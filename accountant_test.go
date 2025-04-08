package aicost

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Counter_TokenCount(t *testing.T) {
	// Setup test models
	testModels := []Model{
		{
			Provider: "openai",
			Model:    "gpt-4",
			Version:  "1",
			CostInput: Money{
				Units:        0,
				Nanos:        3000000,
				CurrencyCode: "USD",
			},
			CostOutput: Money{
				Units:        0,
				Nanos:        6000000,
				CurrencyCode: "USD",
			},
		},
	}

	con := NewConverter("USD", testRates)
	accountant := NewAccountant(testModels, con, true)

	tests := []struct {
		name     string
		provider string
		model    string
		content  string
		want     int64
		wantErr  bool
	}{
		{
			name:     "basic token count",
			provider: "openai",
			model:    "gpt-4",
			content:  "Hello, world!",
			want:     4,
			wantErr:  false,
		},
		{
			name:     "longer text",
			provider: "openai",
			model:    "gpt-4",
			content:  "This is a longer text that should have more tokens. It includes multiple sentences and should give us a reasonable count to test with.",
			want:     26,
			wantErr:  false,
		},
		{
			name:     "unsupported model",
			provider: "openai",
			model:    "nonexistent-model",
			content:  "Hello, world!",
			want:     0,
			wantErr:  true,
		},
		{
			name:     "empty string",
			provider: "openai",
			model:    "gpt-4",
			content:  "",
			want:     0,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := accountant.TokenCount(tt.provider, tt.model, tt.content)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_Counter_CostForModelInput(t *testing.T) {
	// Setup test models
	testModels := []Model{
		{
			Provider: "openai",
			Model:    "gpt-4",
			Version:  "1",
			CostInput: Money{
				Units:        0,
				Nanos:        3000000, // $0.003 per 1k tokens
				CurrencyCode: "USD",
			},
			CostOutput: Money{
				Units:        0,
				Nanos:        6000000,
				CurrencyCode: "USD",
			},
		},
		{
			Provider: "anthropic",
			Model:    "claude-3",
			Version:  "1",
			CostInput: Money{
				Units:        0,
				Nanos:        8000000, // $0.008 per 1k tokens
				CurrencyCode: "USD",
			},
			CostOutput: Money{
				Units:        0,
				Nanos:        24000000,
				CurrencyCode: "USD",
			},
		},
	}

	con := NewConverter("USD", testRates)

	accountant := NewAccountant(testModels, con, true)

	tests := []struct {
		name          string
		provider      string
		model         string
		userCurrency  string
		tokens        int64
		wantCost      *Money
		wantConverted *Money
		wantErr       bool
	}{
		{
			name:         "gpt-4 cost for 1000 tokens in USD",
			provider:     "openai",
			model:        "gpt-4",
			userCurrency: "USD",
			tokens:       1000,
			wantCost: &Money{
				Units:        3,
				Nanos:        0, // $0.003 * 1000 = $3
				CurrencyCode: "USD",
			},
			wantConverted: &Money{
				Units:        3,
				Nanos:        0,
				CurrencyCode: "USD",
			},
			wantErr: false,
		},
		{
			name:         "gpt-4 cost for 1000 tokens in EUR",
			provider:     "openai",
			model:        "gpt-4",
			userCurrency: "EUR",
			tokens:       1000,
			wantCost: &Money{
				Units:        3,
				Nanos:        0,
				CurrencyCode: "USD",
			},
			wantConverted: &Money{
				Units:        2,
				Nanos:        550000000, // €2.55
				CurrencyCode: "EUR",
			},
			wantErr: false,
		},
		{
			name:         "claude-3 cost for 1000 tokens in EUR",
			provider:     "anthropic",
			model:        "claude-3",
			userCurrency: "EUR",
			tokens:       1000,
			wantCost: &Money{
				Units:        8,
				Nanos:        0,
				CurrencyCode: "USD",
			},
			wantConverted: &Money{
				Units:        6,
				Nanos:        800000000, // €6.80
				CurrencyCode: "EUR",
			},
			wantErr: false,
		},
		{
			name:          "model not found",
			provider:      "openai",
			model:         "nonexistent-model",
			userCurrency:  "USD",
			tokens:        1000,
			wantCost:      nil,
			wantConverted: nil,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCost, gotConverted, err := accountant.CostForModelInput(tt.provider, tt.model, tt.userCurrency, tt.tokens)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, gotCost)
				assert.Nil(t, gotConverted)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantCost, gotCost)
			assert.Equal(t, tt.wantConverted, gotConverted)
		})
	}
}

func Test_Counter_CostForModelOutput(t *testing.T) {
	// Setup test models
	testModels := []Model{
		{
			Provider: "openai",
			Model:    "gpt-4",
			Version:  "1",
			CostInput: Money{
				Units:        0,
				Nanos:        3000,
				CurrencyCode: "USD",
			},
			CostOutput: Money{
				Units:        0,
				Nanos:        6000,
				CurrencyCode: "USD",
			},
		},
		{
			Provider: "anthropic",
			Model:    "claude-3",
			Version:  "1",
			CostInput: Money{
				Units:        0,
				Nanos:        8000,
				CurrencyCode: "USD",
			},
			CostOutput: Money{
				Units:        0,
				Nanos:        24000, // $0.024 per 1k tokens
				CurrencyCode: "USD",
			},
		},
	}

	con := NewConverter("USD", testRates)
	accountant := NewAccountant(testModels, con, true)

	tests := []struct {
		name          string
		provider      string
		model         string
		userCurrency  string
		tokens        int64
		wantCost      *Money
		wantConverted *Money
		wantErr       bool
	}{
		{
			name:         "gpt-4 cost for 1000 tokens in USD",
			provider:     "openai",
			model:        "gpt-4",
			userCurrency: "USD",
			tokens:       1000,
			wantCost: &Money{
				Units:        0,
				Nanos:        6000000,
				CurrencyCode: "USD",
			},
			wantConverted: &Money{
				Units:        0,
				Nanos:        6000000,
				CurrencyCode: "USD",
			},
			wantErr: false,
		},
		{
			name:         "gpt-4 cost for 1000 tokens in EUR",
			provider:     "openai",
			model:        "gpt-4",
			userCurrency: "EUR",
			tokens:       1000,
			wantCost: &Money{
				Units:        0,
				Nanos:        6000000,
				CurrencyCode: "USD",
			},
			wantConverted: &Money{
				Units:        0,
				Nanos:        5100000,
				CurrencyCode: "EUR",
			},
			wantErr: false,
		},
		{
			name:         "claude-3 cost for 1000 tokens in EUR",
			provider:     "anthropic",
			model:        "claude-3",
			userCurrency: "EUR",
			tokens:       1000,
			wantCost: &Money{
				Units:        0,
				Nanos:        24000000,
				CurrencyCode: "USD",
			},
			wantConverted: &Money{
				Units:        0,
				Nanos:        20400000,
				CurrencyCode: "EUR",
			},
			wantErr: false,
		},
		{
			name:          "model not found",
			provider:      "openai",
			model:         "nonexistent-model",
			userCurrency:  "USD",
			tokens:        1000,
			wantCost:      nil,
			wantConverted: nil,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCost, gotConverted, err := accountant.CostForModelOutput(tt.provider, tt.model, tt.userCurrency, tt.tokens)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, gotCost)
				assert.Nil(t, gotConverted)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantCost, gotCost)
			assert.Equal(t, tt.wantConverted, gotConverted)
		})
	}
}

func Test_Counter_Models(t *testing.T) {
	// Setup test models
	testModels := []Model{
		{
			Provider: "openai",
			Model:    "gpt-4",
			Version:  "1",
			CostInput: Money{
				Units:        0,
				Nanos:        3000000,
				CurrencyCode: "USD",
			},
			CostOutput: Money{
				Units:        0,
				Nanos:        6000000,
				CurrencyCode: "USD",
			},
		},
		{
			Provider: "anthropic",
			Model:    "claude-3",
			Version:  "1",
			CostInput: Money{
				Units:        0,
				Nanos:        8000000,
				CurrencyCode: "USD",
			},
			CostOutput: Money{
				Units:        0,
				Nanos:        24000000,
				CurrencyCode: "USD",
			},
		},
	}

	con := NewConverter("USD", testRates)
	accountant := NewAccountant(testModels, con, true)

	tests := []struct {
		name string
		want []Model
	}{
		{
			name: "get all models",
			want: testModels,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := accountant.Models(nil)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_Counter_findModel(t *testing.T) {
	// Setup test models
	testModels := []Model{
		{
			Provider: "openai",
			Model:    "gpt-4",
			Version:  "1",
			CostInput: Money{
				Units:        0,
				Nanos:        3000000,
				CurrencyCode: "USD",
			},
			CostOutput: Money{
				Units:        0,
				Nanos:        6000000,
				CurrencyCode: "USD",
			},
		},
		{
			Provider: "anthropic",
			Model:    "claude-3",
			Version:  "1",
			CostInput: Money{
				Units:        0,
				Nanos:        8000000,
				CurrencyCode: "USD",
			},
			CostOutput: Money{
				Units:        0,
				Nanos:        24000000,
				CurrencyCode: "USD",
			},
		},
		{
			Provider: "anthropic",
			Model:    "claude-2",
			Version:  "1",
			CostInput: Money{
				Units:        0,
				Nanos:        6000000,
				CurrencyCode: "USD",
			},
			CostOutput: Money{
				Units:        0,
				Nanos:        18000000,
				CurrencyCode: "USD",
			},
		},
	}

	con := NewConverter("USD", testRates)
	accountant := NewAccountant(testModels, con, true)

	tests := []struct {
		name     string
		provider string
		model    string
		want     *Model
	}{
		{
			name:     "find gpt-4 with provider",
			provider: "openai",
			model:    "gpt-4",
			want:     &testModels[0],
		},
		{
			name:     "find claude-3 with provider",
			provider: "anthropic",
			model:    "claude-3",
			want:     &testModels[1],
		},
		{
			name:     "find claude-2 with provider",
			provider: "anthropic",
			model:    "claude-2",
			want:     &testModels[2],
		},
		{
			name:     "cant find gpt-4 without provider",
			provider: "",
			model:    "gpt-4",
			want:     nil,
		},
		{
			name:     "provider mismatch",
			provider: "anthropic",
			model:    "gpt-4",
			want:     nil,
		},
		{
			name:     "model not found",
			provider: "openai",
			model:    "nonexistent-model",
			want:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := accountant.findModel(tt.provider, tt.model)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_Counter_calculateCost(t *testing.T) {
	// Setup test models
	testModels := []Model{
		{
			Provider: "openai",
			Model:    "gpt-4",
			Version:  "1",
			CostInput: Money{
				Units:        0,
				Nanos:        3000000,
				CurrencyCode: "USD",
			},
		},
	}

	con := NewConverter("USD", testRates)
	accountant := NewAccountant(testModels, con, false)

	tests := []struct {
		name          string
		tokens        int64
		costPerToken  Money
		userCurrency  string
		wantCost      *Money
		wantConverted *Money
		wantErr       bool
	}{
		{
			name:   "1000 tokens at 0.003 USD per token to EUR",
			tokens: 1000,
			costPerToken: Money{
				Units:        0,
				Nanos:        10000,
				CurrencyCode: "USD",
			},
			userCurrency: "EUR",
			wantCost: &Money{
				Units:        0,
				Nanos:        10000000,
				CurrencyCode: "USD",
			},
			wantConverted: &Money{
				Units:        0,
				Nanos:        8500000,
				CurrencyCode: "EUR",
			},
			wantErr: false,
		},
		{
			name:   "500 tokens at 0.003 USD per token to EUR",
			tokens: 500,
			costPerToken: Money{
				Units:        0,
				Nanos:        10000,
				CurrencyCode: "USD",
			},
			userCurrency: "EUR",
			wantCost: &Money{
				Units:        0,
				Nanos:        5000000,
				CurrencyCode: "USD",
			},
			wantConverted: &Money{
				Units:        0,
				Nanos:        4250000,
				CurrencyCode: "EUR",
			},
			wantErr: false,
		},
		{
			name:   "conversion error",
			tokens: 1000,
			costPerToken: Money{
				Units:        0,
				Nanos:        1000000, // $0.001 per token
				CurrencyCode: "USD",
			},
			userCurrency:  "CAD", // Will cause conversion error
			wantCost:      nil,
			wantConverted: nil,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCost, gotConverted, err := accountant.calculateCost(tt.tokens, tt.costPerToken, tt.userCurrency)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, gotCost)
				assert.Nil(t, gotConverted)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantCost, gotCost)
			assert.Equal(t, tt.wantConverted, gotConverted)
		})
	}
}
