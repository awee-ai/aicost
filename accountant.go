package llm_pricing

import (
	"fmt"
	"log"
	"strings"

	"github.com/pkoukk/tiktoken-go"

	tiktokenloader "github.com/pkoukk/tiktoken-go-loader"
)

// Model represents a model with its price
type Model struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
	// Releases is a list of releases for the model
	// empty list means exact model match
	// * means any release in consecutive release order
	Releases []string `json:"releases"`
	// PriceQuery is the price per token for a query message
	// per 1000 tokens
	PriceQuery Money `json:"price_query"`
	// PriceOutput is the price per token for an output message
	// per 1000 tokens
	PriceOutput Money `json:"price_output"`
}

// Accountant is an interface for model price calculation
type Accountant interface {
	TokenCount(provider, model string, content string) (int64, error)
	PriceForModelInput(provider, model string, userCurrency string, tokens int64) (*Money, *Money, error)
	PriceForModelOutput(provider, model string, userCurrency string, tokens int64) (*Money, *Money, error)
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

// PriceForModelQuery returns the price for a model query
func (p *Counter) PriceForModelInput(provider, model string, userCurrency string, tokens int64) (*Money, *Money, error) {
	pricingModel := p.findModel(provider, model)
	if pricingModel == nil {
		return nil, nil, fmt.Errorf("failed to find model for query price %s: %w", model, ErrPricingModelNotFound)
	}

	price, convertedPrice, err := p.calculateCost(tokens, pricingModel.PriceQuery, userCurrency)
	if err != nil {
		return nil, nil, err
	}

	return price, convertedPrice, nil
}

// PriceForModelOutput returns the price for a model output
func (p *Counter) PriceForModelOutput(provider, model string, userCurrency string, tokens int64) (*Money, *Money, error) {
	pricingModel := p.findModel(provider, model)
	if pricingModel == nil {
		return nil, nil, fmt.Errorf("failed to find model for output price %s: %w", model, ErrPricingModelNotFound)
	}

	price, convertedPrice, err := p.calculateCost(tokens, pricingModel.PriceOutput, userCurrency)
	if err != nil {
		return nil, nil, err
	}

	return price, convertedPrice, nil
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

func (p *Counter) calculateCost(tokens int64, pricePerToken Money, userCurrency string) (*Money, *Money, error) {
	// calculate price and take into consideration that we need to convert the price per thousand
	price := float64(tokens) * MoneyToFloat64(pricePerToken)

	// price := float32(tokens) * pricePerThousand
	converted, err := p.converter.Convert(CurrencyAmount(price), pricePerToken.CurrencyCode, userCurrency)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert price from %s to %s: %w", pricePerToken.CurrencyCode, userCurrency, err)
	}

	originalCost := NewMoneyFromFloat(pricePerToken.CurrencyCode, price)
	convertedCost := NewMoneyFromFloat(userCurrency, float64(converted))

	return &originalCost, &convertedCost, nil
}
