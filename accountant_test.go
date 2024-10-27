package aicost

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type test struct {
	name  string
	model string

	currency string
	tokens   int64

	expected       float64
	expectedOutput string
}

var testProvider = "llm-api"

var testModelsUSD = []Model{
	{
		Provider: testProvider,
		Model:    "model-1",

		PriceQuery:  NewMoneyFromFloat(CurrencyUSD, 0.05/1000),
		PriceOutput: NewMoneyFromFloat(CurrencyUSD, 0.1/1000),
	},
	{
		Provider: testProvider,
		Model:    "model-2",

		PriceQuery:  NewMoneyFromFloat(CurrencyUSD, 0.1/1000),
		PriceOutput: NewMoneyFromFloat(CurrencyUSD, 0.2/1000),
	},
	{
		Provider: testProvider,
		Model:    "gpt-4",

		PriceQuery:  NewMoneyFromFloat(CurrencyUSD, 0.03/1000),
		PriceOutput: NewMoneyFromFloat(CurrencyUSD, 0.06/1000),
	},
}

var converter = &Converter{
	CurrencyUSD,
	map[string]CurrencyAmount{
		CurrencyUSD: 1,
		CurrencyEUR: 1.5,
	},
}

func Test_USD_Model_Pricing_ForModelQuery(t *testing.T) {
	price := NewAccountant(testModelsUSD, converter, false)

	tests := []test{
		{
			name:           "model-1",
			model:          "model-1",
			tokens:         10_000,
			currency:       CurrencyUSD,
			expected:       0.5,
			expectedOutput: "$0.5000",
		},
		{
			name:           "model-2",
			model:          "model-2",
			tokens:         50_000,
			currency:       CurrencyUSD,
			expected:       5,
			expectedOutput: "$5.0000",
		},
		{
			name:           "model-1",
			model:          "model-1",
			tokens:         10_000,
			currency:       CurrencyEUR,
			expected:       0.5,
			expectedOutput: "€0.5000",
		},
		{
			name:           "model-2",
			model:          "model-2",
			tokens:         50_000,
			currency:       CurrencyEUR,
			expected:       4.5746183,
			expectedOutput: "€4.5746",
		},
		{
			name:           "gpt-4",
			model:          "gpt-4",
			tokens:         1,
			currency:       CurrencyUSD,
			expected:       0.00003,
			expectedOutput: "$0.0000",
		},
		{
			name:           "gpt-4",
			model:          "gpt-4",
			tokens:         1,
			currency:       CurrencyEUR,
			expected:       0.00002744771,
			expectedOutput: "€0.0000",
		},
		{
			name:           "gpt-4",
			model:          "gpt-4",
			tokens:         20,
			currency:       CurrencyUSD,
			expected:       0.00059999997,
			expectedOutput: "$0.0006",
		},
		{
			name:           "gpt-4",
			model:          "gpt-4",
			tokens:         20,
			currency:       CurrencyEUR,
			expected:       0.0005489542,
			expectedOutput: "€0.0005",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, actualConverted, err := price.PriceForModelInput(testProvider, tt.model, tt.currency, tt.tokens)
			assert.NoError(t, err)

			assert.Equal(t, NewMoneyFromFloat(CurrencyUSD, tt.expected), *actual)
			assert.NotNil(t, actualConverted)
			assert.Equal(t, tt.expectedOutput, FormatPrice(*actual))
		})
	}
}

// FormatPrice formats a price with currency
func FormatPrice(price Money) string {
	switch price.CurrencyCode {
	case CurrencyUSD:
		return fmt.Sprintf("$%.4f", MoneyToFloat64(price))
	case CurrencyEUR:
		return fmt.Sprintf("€%.4f", MoneyToFloat64(price))
	default:
		return fmt.Sprintf("%s", MoneyToString(price))
	}
}

func Test_USD_Model_Pricing_ForModelOutput(t *testing.T) {
	price := NewAccountant(testModelsUSD, converter, false)

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
		{
			model:          "gpt-4",
			tokens:         1,
			currency:       CurrencyUSD,
			expected:       0.00006,
			expectedOutput: "$0.0001",
		},
		{
			model:          "gpt-4",
			tokens:         1,
			currency:       CurrencyEUR,
			expected:       0.00005489542,
			expectedOutput: "€0.0001",
		},
		{
			model:          "gpt-4",
			tokens:         9,
			currency:       CurrencyUSD,
			expected:       0.00054,
			expectedOutput: "$0.0005",
		},
		{
			model:          "gpt-4",
			tokens:         9,
			currency:       CurrencyEUR,
			expected:       0.00049405877,
			expectedOutput: "€0.0005",
		},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%s_%s_%s_%d", testProvider, tt.model, tt.currency, tt.tokens)
		t.Run(name, func(t *testing.T) {
			actualUSD, actualConverted, err := price.PriceForModelOutput(testProvider, tt.model, tt.currency, tt.tokens)
			assert.NoError(t, err)

			expectedUSD := NewMoneyFromFloat(CurrencyUSD, tt.expected)
			assert.Equal(t, expectedUSD, actualUSD)
			assert.NotNil(t, actualConverted)
			assert.Equal(t, tt.expectedOutput, FormatPrice(*actualUSD))
		})
	}
}
