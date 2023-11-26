package llm_pricing

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type test struct {
	model string

	currency string
	tokens   int

	expected       float32
	expectedOutput string
}

var testProvider = "llm-api"

var testModelsUSD = []Model{
	{
		Provider:    testProvider,
		Model:       "model-1",
		Currency:    CurrencyUSD,
		PriceQuery:  0.05 / 1000,
		PriceOutput: 0.1 / 1000,
	},
	{
		Provider:    testProvider,
		Model:       "model-2",
		Currency:    CurrencyUSD,
		PriceQuery:  0.1 / 1000,
		PriceOutput: 0.2 / 1000,
	},
}

var converter = &Conversion{
	CurrencyUSD,
	map[string]CurrencyAmount{
		CurrencyUSD: 1,
		CurrencyEUR: 0.91492366,
	},
}

func Test_USD_Model_Pricing_ForModelQuery(t *testing.T) {
	price := NewPricing(testModelsUSD, converter)

	tests := []test{
		{
			model:          "model-1",
			tokens:         10_000,
			currency:       CurrencyUSD,
			expected:       0.5,
			expectedOutput: "$0.5000",
		},
		{
			model:          "model-2",
			tokens:         50_000,
			currency:       CurrencyUSD,
			expected:       5,
			expectedOutput: "$5.0000",
		},
		{
			model:          "model-1",
			tokens:         10_000,
			currency:       CurrencyEUR,
			expected:       0.45746183,
			expectedOutput: "€0.4575",
		},
		{
			model:          "model-2",
			tokens:         50_000,
			currency:       CurrencyEUR,
			expected:       4.5746183,
			expectedOutput: "€4.5746",
		},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%s_%s_%s_%d", testProvider, tt.model, tt.currency, tt.tokens)
		t.Run(name, func(t *testing.T) {
			actual, err := price.ForModelQuery(testProvider, tt.model, tt.currency, tt.tokens)
			assert.NoError(t, err)

			assert.Equal(t, tt.expected, actual)
			assert.Equal(t, tt.expectedOutput, FormatPrice(actual, tt.currency))
		})
	}
}

func Test_USD_Model_Pricing_ForModelOutput(t *testing.T) {
	price := NewPricing(testModelsUSD, converter)

	tests := []test{
		{
			model:          "model-1",
			tokens:         10_000,
			currency:       CurrencyUSD,
			expected:       1,
			expectedOutput: "$1.0000",
		},
		{
			model:          "model-2",
			tokens:         50_000,
			currency:       CurrencyUSD,
			expected:       10,
			expectedOutput: "$10.0000",
		},
		{
			model:          "model-1",
			tokens:         10_000,
			currency:       CurrencyEUR,
			expected:       0.91492367,
			expectedOutput: "€0.9149",
		},
		{
			model:          "model-2",
			tokens:         50_000,
			currency:       CurrencyEUR,
			expected:       9.149237,
			expectedOutput: "€9.1492",
		},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%s_%s_%s_%d", testProvider, tt.model, tt.currency, tt.tokens)
		t.Run(name, func(t *testing.T) {
			actual, err := price.ForModelOutput(testProvider, tt.model, tt.currency, tt.tokens)
			assert.NoError(t, err)

			assert.Equal(t, tt.expected, actual)
			assert.Equal(t, tt.expectedOutput, FormatPrice(actual, tt.currency))
		})
	}
}

//
// var testModelsEUR = []Model{
// 	{
// 		Provider:    testProvider,
// 		Model:       "model-1",
// 		Currency:    CurrencyEUR,
// 		PriceQuery:  0.05 / 1000,
// 		PriceOutput: 0.1 / 1000,
// 	},
// 	{
// 		Provider:    testProvider,
// 		Model:       "model-2",
// 		Currency:    CurrencyEUR,
// 		PriceQuery:  0.1 / 1000,
// 		PriceOutput: 0.2 / 1000,
// 	},
// }
