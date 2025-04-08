package aicost

import (
	"fmt"
	"strings"

	"github.com/pkoukk/tiktoken-go"

	tiktokenloader "github.com/pkoukk/tiktoken-go-loader"
)

// Model represents a model with its cost
type Model struct {
	Provider string `json:"provider" yaml:"provider"`
	Model    string `json:"model" yaml:"model"`
	Version  string `json:"version" yaml:"version"`
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
	Models(models []Model) []Model
}

var ErrPricingModelNotFound = fmt.Errorf("model not supported")
var ErrTokenizerNotFound = fmt.Errorf("tokenizer not found")

// Counter is a pricing calculator
type Counter struct {
	models    []Model
	converter Converter
}

var _ Accountant = (*Counter)(nil)

// NewAccountant returns a new pricing
func NewAccountant(models []Model, converter Converter, bpe bool) *Counter {
	if bpe {
		loader := tiktokenloader.NewOfflineLoader()
		tiktoken.SetBpeLoader(loader)
	} else {
		tiktoken.SetBpeLoader(nil)
	}

	return &Counter{
		models:    models,
		converter: converter,
	}
}

// Models returns or sets the models
func (p *Counter) Models(models []Model) []Model {
	if models != nil {
		p.models = models

		return p.models
	}

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
		if m.Provider != provider {
			continue
		}
		if m.Model == model {
			mod = &m
			break
		}
	}
	return mod
}

func (p *Counter) calculateCost(tokens int64, costPerToken Money, userCurrency string) (*Money, *Money, error) {
	cost, err := costPerToken.Times(tokens)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to multiply tokens %d: %w", tokens, err)
	}

	converted, err := p.converter.Convert(*cost, userCurrency)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert cost from %s to %s: %w", costPerToken.CurrencyCode, userCurrency, err)
	}

	return cost, converted, nil
}
