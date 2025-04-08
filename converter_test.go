package aicost

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var testRates = map[string]float64{
	"EUR": 0.85,
	"GBP": 0.75,
	"JPY": 110.0,
}

func Test_NewConverter(t *testing.T) {
	tests := []struct {
		name         string
		baseCurrency string
		rates        map[string]float64
		want         *converter
	}{
		{
			name:         "basic initialization",
			baseCurrency: "USD",
			rates:        testRates,
			want: &converter{
				baseCurrency: "USD",
				rates:        testRates,
			},
		},
		{
			name:         "empty rates",
			baseCurrency: "USD",
			rates:        map[string]float64{},
			want: &converter{
				baseCurrency: "USD",
				rates:        map[string]float64{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewConverter(tt.baseCurrency, tt.rates)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_Converter_Convert(t *testing.T) {
	// Setup test converter with USD as base
	testConverter := NewConverter("USD", testRates)

	tests := []struct {
		name       string
		amount     Money
		toCurrency string
		want       *Money
		wantErr    bool
	}{
		{
			name: "same currency",
			amount: Money{
				Units:        100,
				Nanos:        500000000,
				CurrencyCode: "USD",
			},
			toCurrency: "USD",
			want: &Money{
				Units:        100,
				Nanos:        500000000,
				CurrencyCode: "USD",
			},
			wantErr: false,
		},
		{
			name: "convert from base to another currency",
			amount: Money{
				Units:        100,
				Nanos:        0,
				CurrencyCode: "USD",
			},
			toCurrency: "EUR",
			want: &Money{
				Units:        85,
				Nanos:        0,
				CurrencyCode: "EUR",
			},
			wantErr: false,
		},
		{
			name: "convert from non-base to base",
			amount: Money{
				Units:        100,
				Nanos:        0,
				CurrencyCode: "EUR",
			},
			toCurrency: "USD",
			want: &Money{
				Units:        115,
				Nanos:        0,
				CurrencyCode: "USD",
			},
			wantErr: false,
		},
		{
			name: "convert between two non-base currencies",
			amount: Money{
				Units:        100,
				Nanos:        0,
				CurrencyCode: "EUR",
			},
			toCurrency: "GBP",
			want: &Money{
				Units:        86,
				Nanos:        250000000,
				CurrencyCode: "GBP",
			},
			wantErr: false,
		},
		{
			name: "source currency not found",
			amount: Money{
				Units:        100,
				Nanos:        0,
				CurrencyCode: "CAD", // Not in rates
			},
			toCurrency: "USD",
			want:       nil,
			wantErr:    true,
		},
		{
			name: "target currency not found",
			amount: Money{
				Units:        100,
				Nanos:        0,
				CurrencyCode: "USD",
			},
			toCurrency: "CAD", // Not in rates
			want:       nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := testConverter.Convert(tt.amount, tt.toCurrency)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_Converter_ConvertToBase(t *testing.T) {
	// Setup test converter with USD as base
	testConverter := NewConverter("USD", testRates)

	tests := []struct {
		name    string
		amount  Money
		want    *Money
		wantErr bool
	}{
		{
			name: "already base currency",
			amount: Money{
				Units:        100,
				Nanos:        500000000,
				CurrencyCode: "USD",
			},
			want: &Money{
				Units:        100,
				Nanos:        500000000,
				CurrencyCode: "USD",
			},
			wantErr: false,
		},
		{
			name: "convert EUR to base (USD)",
			amount: Money{
				Units:        100,
				Nanos:        0,
				CurrencyCode: "EUR",
			},
			want: &Money{
				Units:        115,
				Nanos:        0,
				CurrencyCode: "USD",
			},
			wantErr: false,
		},
		{
			name: "convert GBP to base (USD)",
			amount: Money{
				Units:        100,
				Nanos:        0,
				CurrencyCode: "GBP",
			},
			want: &Money{
				Units:        125,
				Nanos:        0,
				CurrencyCode: "USD",
			},
			wantErr: false,
		},
		{
			name: "currency not found",
			amount: Money{
				Units:        100,
				Nanos:        0,
				CurrencyCode: "CAD", // Not in rates
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := testConverter.convertToBase(tt.amount)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
