package aicost

import (
	"fmt"
	"log"
	"strings"

	"github.com/pkoukk/tiktoken-go"

	tiktokenloader "github.com/pkoukk/tiktoken-go-loader"
)

// Model represents a model with its cost
type Model struct {
	Provider string `json:"provider" yaml:"provider"`
	Model    string `json:"model" yaml:"model"`
	// Releases is a list of releases for the model
	// empty list means exact model match
	// * means any release in consecutive release order
	Releases []string `json:"releases" yaml:"releases"`
	// CostInput is the cost (usually) per 1k tokens for a query message
	CostInput Money `json:"cost_input" yaml:"cost_input"`
	// CostOutput is the cost (usually) per tokens for an output message
	CostOutput Money `json:"cost_output" yaml:"cost_output"`
}

// Accountant is an interface for model cost calculation
type Accountant interface {
	TokenCount(provider, model string, content string) (int64, error)
	CostForModelInput(provider, model string, userCurrency string, tokens int64) (*Money, *Money, error)
	CostForModelOutput(provider, model string, userCurrency string, tokens int64) (*Money, *Money, error)
	Models() []Model
}

var ErrPricingModelNotFound = fmt.Errorf("model not supported")
var ErrTokenizerNotFound = fmt.Errorf("tokenizer not found")

// Counter is a pricing calculator
type Counter struct {
	models    []Model
	converter CurrencyConversion
}

var _ Accountant = (*Counter)(nil)

// NewAccountant returns a new pricing
func NewAccountant(models []Model, converter CurrencyConversion, bpe bool) *Counter {
	// TODO: we may need a way to reset this
	if bpe {
		log.Printf("setting offline bpe loader")
		loader := tiktokenloader.NewOfflineLoader()
		tiktoken.SetBpeLoader(loader)
	} else {
		log.Printf("setting nil bpe loader")
		tiktoken.SetBpeLoader(nil)
	}

	return &Counter{
		models:    models,
		converter: converter,
	}
}

// Models returns the list of models
func (p *Counter) Models() []Model {
	return p.models
}

// TokenCount returns the token count for a message
func (p *Counter) TokenCount(provider, model string, content string) (int64, error) {
	tkm, err := tiktoken.EncodingForModel(model)
	if err != nil {
		// no encoding for model
		if strings.Contains(err.Error(), "no encoding for model") {
			return 0, fmt.Errorf("%w: %w", err, ErrTokenizerNotFound)
		}

		return 0, fmt.Errorf("failed to get encoding for model %s: %w", model, err)
	}
	tokens := len(tkm.Encode(content, nil, nil))
	return int64(tokens), nil
}

// CostForModelInput returns the cost for a model query
func (p *Counter) CostForModelInput(provider, model string, userCurrency string, tokens int64) (*Money, *Money, error) {
	pricingModel := p.findModel(provider, model)
	if pricingModel == nil {
		return nil, nil, fmt.Errorf("failed to find model for input cost %s: %w", model, ErrPricingModelNotFound)
	}

	cost, convertedCost, err := p.calculateCost(tokens, pricingModel.CostInput, userCurrency)
	if err != nil {
		return nil, nil, err
	}

	return cost, convertedCost, nil
}

// CostForModelOutput returns the cost for a model output
func (p *Counter) CostForModelOutput(provider, model string, userCurrency string, tokens int64) (*Money, *Money, error) {
	pricingModel := p.findModel(provider, model)
	if pricingModel == nil {
		return nil, nil, fmt.Errorf("failed to find model for output cost %s: %w", model, ErrPricingModelNotFound)
	}

	cost, convertedCost, err := p.calculateCost(tokens, pricingModel.CostOutput, userCurrency)
	if err != nil {
		return nil, nil, err
	}

	return cost, convertedCost, nil
}

func (p *Counter) findModel(provider, model string) *Model {
	var mod *Model
	for _, m := range p.models {
		// "" allows us to skip the provider check
		if provider != "" && m.Provider != provider {
			continue
		}

		cnt := len(m.Releases)
		// no releases means exact model match
		if cnt == 0 {
			if m.Model == model {
				mod = &m
				break
			}
			continue
		}

		if p.matchModelRelease(model, m) {
			// log.Printf("matched model by release")
			mod = &m
			break
		}
	}

	if mod == nil {
		return nil
	}

	return mod
}

func (p *Counter) matchModelRelease(givenModel string, model Model) bool {
	// make sure the given model starts with the base model
	// "gpt-4-0125-preview", "gpt-4"
	// log.Printf("givenModel: %s, model.Model: %s", givenModel, model.Model)
	if !strings.HasPrefix(givenModel, model.Model) {
		// log.Printf("model does not start with base model")
		return false
	}
	// log.Printf("model found: %s", model.Model)
	// log.Printf("model releases: %v", model.Releases)

	// * means any release in consecutive release order
	// e.g. gpt-4-0125-preview, gpt-4-0125, gpt-4-1106-preview
	// gpt-4 is the base model
	for _, release := range model.Releases {
		if release == "*" {
			return true
		}
		if model.Model+"-"+release == givenModel {
			return true
		}
	}

	return false
}

func (p *Counter) calculateCost(tokens int64, costPerToken Money, userCurrency string) (*Money, *Money, error) {
	// calculate cost and take into consideration that we need to convert the cost per thousand
	cost := float64(tokens) * MoneyToFloat64(costPerToken)

	// cost := float32(tokens) * costPerThousand
	converted, err := p.converter.Convert(CurrencyAmount(cost), costPerToken.CurrencyCode, userCurrency)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert cost from %s to %s: %w", costPerToken.CurrencyCode, userCurrency, err)
	}

	originalCost := NewMoneyFromFloat(costPerToken.CurrencyCode, cost)
	convertedCost := NewMoneyFromFloat(userCurrency, float64(converted))

	return &originalCost, &convertedCost, nil
}
