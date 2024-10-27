package aicost

import "testing"

func TestMoneyToString(t *testing.T) {
	tests := []struct {
		name     string
		money    Money
		expected string
	}{
		{
			name:     "Basic USD value",
			money:    Money{CurrencyCode: "USD", Units: 1, Nanos: 0},
			expected: "USD 1.000000000",
		},
		{
			name:     "Small value",
			money:    Money{CurrencyCode: "USD", Units: 0, Nanos: 5000000},
			expected: "USD 0.005000000",
		},
		{
			name:     "Complex value",
			money:    Money{CurrencyCode: "EUR", Units: 123, Nanos: 456789012},
			expected: "EUR 123.456789012",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MoneyToString(tt.money)
			if result != tt.expected {
				t.Errorf("MoneyToString(%v) = %v; want %v", tt.money, result, tt.expected)
			}
		})
	}
}

func TestMoneyToFloat64(t *testing.T) {
	tests := []struct {
		name     string
		money    Money
		expected float64
	}{
		{
			name:     "Basic USD value",
			money:    Money{CurrencyCode: "USD", Units: 1, Nanos: 0},
			expected: 1.0,
		},
		{
			name:     "Small value",
			money:    Money{CurrencyCode: "USD", Units: 0, Nanos: 5000000},
			expected: 0.005,
		},
		{
			name:     "Complex value",
			money:    Money{CurrencyCode: "EUR", Units: 123, Nanos: 456789012},
			expected: 123.456789012,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MoneyToFloat64(tt.money)
			if result != tt.expected {
				t.Errorf("MoneyToFloat64(%v) = %v; want %v", tt.money, result, tt.expected)
			}
		})
	}
}
