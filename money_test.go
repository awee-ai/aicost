package aicost

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewMoney(t *testing.T) {
	tests := []struct {
		name     string
		currency string
		units    int64
		nanos    int32
		want     *Money
		wantErr  bool
	}{
		{
			name:     "valid positive",
			currency: "USD",
			units:    10,
			nanos:    500000000,
			want: &Money{
				Units:        10,
				Nanos:        500000000,
				CurrencyCode: "USD",
			},
			wantErr: false,
		},
		{
			name:     "valid negative",
			currency: "EUR",
			units:    -5,
			nanos:    -750000000,
			want: &Money{
				Units:        -5,
				Nanos:        -750000000,
				CurrencyCode: "EUR",
			},
			wantErr: false,
		},
		{
			name:     "valid zero",
			currency: "JPY",
			units:    0,
			nanos:    0,
			want: &Money{
				Units:        0,
				Nanos:        0,
				CurrencyCode: "JPY",
			},
			wantErr: false,
		},
		{
			name:     "nanos out of range",
			currency: "USD",
			units:    1,
			nanos:    1200000000, // exceeds 999999999
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "conflicting signs",
			currency: "USD",
			units:    1,
			nanos:    -500000000, // units positive, nanos negative
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "empty currency",
			currency: "",
			units:    1,
			nanos:    500000000,
			want:     nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewMoney(tt.currency, tt.units, tt.nanos)
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

func Test_Money_Add(t *testing.T) {
	tests := []struct {
		name    string
		m1      *Money
		m2      *Money
		want    *Money
		wantErr bool
	}{
		{
			name: "valid addition",
			m1: &Money{
				Units:        5,
				Nanos:        200000000,
				CurrencyCode: "USD",
			},
			m2: &Money{
				Units:        3,
				Nanos:        400000000,
				CurrencyCode: "USD",
			},
			want: &Money{
				Units:        8,
				Nanos:        600000000,
				CurrencyCode: "USD",
			},
			wantErr: false,
		},
		{
			name: "addition with carry",
			m1: &Money{
				Units:        5,
				Nanos:        800000000,
				CurrencyCode: "USD",
			},
			m2: &Money{
				Units:        3,
				Nanos:        400000000,
				CurrencyCode: "USD",
			},
			want: &Money{
				Units:        9,
				Nanos:        200000000,
				CurrencyCode: "USD",
			},
			wantErr: false,
		},
		{
			name: "negative addition",
			m1: &Money{
				Units:        5,
				Nanos:        0,
				CurrencyCode: "USD",
			},
			m2: &Money{
				Units:        -3,
				Nanos:        0,
				CurrencyCode: "USD",
			},
			want: &Money{
				Units:        2,
				Nanos:        0,
				CurrencyCode: "USD",
			},
			wantErr: false,
		},
		{
			name: "currency mismatch",
			m1: &Money{
				Units:        5,
				Nanos:        0,
				CurrencyCode: "USD",
			},
			m2: &Money{
				Units:        3,
				Nanos:        0,
				CurrencyCode: "EUR",
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.m1.Add(tt.m2)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_Money_Times(t *testing.T) {
	tests := []struct {
		name    string
		money   *Money
		factor  int64
		want    *Money
		wantErr bool
	}{
		{
			name: "multiply by positive",
			money: &Money{
				Units:        2,
				Nanos:        500000000,
				CurrencyCode: "USD",
			},
			factor: 3,
			want: &Money{
				Units:        7,
				Nanos:        500000000,
				CurrencyCode: "USD",
			},
			wantErr: false,
		},
		{
			name: "multiply by zero",
			money: &Money{
				Units:        5,
				Nanos:        500000000,
				CurrencyCode: "USD",
			},
			factor: 0,
			want: &Money{
				Units:        0,
				Nanos:        0,
				CurrencyCode: "USD",
			},
			wantErr: false,
		},
		{
			name: "multiply by negative",
			money: &Money{
				Units:        2,
				Nanos:        500000000,
				CurrencyCode: "USD",
			},
			factor: -2,
			want: &Money{
				Units:        -5,
				Nanos:        0,
				CurrencyCode: "USD",
			},
			wantErr: false,
		},
		{
			name: "potential overflow",
			money: &Money{
				Units:        9223372036854775807, // max int64
				Nanos:        0,
				CurrencyCode: "USD",
			},
			factor:  2,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.money.Times(tt.factor)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_Money_TimesFloat(t *testing.T) {
	tests := []struct {
		name    string
		money   *Money
		factor  float64
		want    *Money
		wantErr bool
	}{
		{
			name: "multiply by positive float",
			money: &Money{
				Units:        2,
				Nanos:        500000000,
				CurrencyCode: "USD",
			},
			factor: 2.5,
			want: &Money{
				Units:        6,
				Nanos:        250000000,
				CurrencyCode: "USD",
			},
			wantErr: false,
		},
		{
			name: "multiply by zero",
			money: &Money{
				Units:        5,
				Nanos:        500000000,
				CurrencyCode: "USD",
			},
			factor: 0.0,
			want: &Money{
				Units:        0,
				Nanos:        0,
				CurrencyCode: "USD",
			},
			wantErr: false,
		},
		{
			name: "multiply by negative float",
			money: &Money{
				Units:        2,
				Nanos:        500000000,
				CurrencyCode: "USD",
			},
			factor: -1.5,
			want: &Money{
				Units:        -3,
				Nanos:        -750000000,
				CurrencyCode: "USD",
			},
			wantErr: false,
		},
		{
			name: "float overflow",
			money: &Money{
				Units:        9223372036854775807, // max int64
				Nanos:        0,
				CurrencyCode: "USD",
			},
			factor:  2.0,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mon, err := NewMoney(tt.money.CurrencyCode, tt.money.Units, tt.money.Nanos)
			assert.NoError(t, err)
			got, err := mon.TimesFloat(tt.factor)
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

func Test_NewMoneyFromFloat(t *testing.T) {
	tests := []struct {
		name     string
		currency string
		amount   float64
		want     *Money
		wantErr  bool
	}{
		{
			name:     "positive amount",
			currency: "USD",
			amount:   12.345,
			want: &Money{
				Units:        12,
				Nanos:        345000000,
				CurrencyCode: "USD",
			},
			wantErr: false,
		},
		{
			name:     "negative amount",
			currency: "EUR",
			amount:   -5.75,
			want: &Money{
				Units:        -5,
				Nanos:        -750000000,
				CurrencyCode: "EUR",
			},
			wantErr: false,
		},
		{
			name:     "zero amount",
			currency: "JPY",
			amount:   0.0,
			want: &Money{
				Units:        0,
				Nanos:        0,
				CurrencyCode: "JPY",
			},
			wantErr: false,
		},
		{
			name:     "very small fractional",
			currency: "USD",
			amount:   0.000000001, // 1 nano
			want: &Money{
				Units:        0,
				Nanos:        1,
				CurrencyCode: "USD",
			},
			wantErr: false,
		},
		{
			name:     "empty currency",
			currency: "",
			amount:   10.0,
			want:     nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewMoneyFromFloat(tt.currency, tt.amount)
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

func Test_MoneyToString(t *testing.T) {
	tests := []struct {
		name  string
		money Money
		want  string
	}{
		{
			name: "positive amount",
			money: Money{
				Units:        12,
				Nanos:        345000000,
				CurrencyCode: "USD",
			},
			want: "USD 12.345000000",
		},
		{
			name: "negative amount",
			money: Money{
				Units:        -5,
				Nanos:        -750000000,
				CurrencyCode: "EUR",
			},
			want: "EUR -5.750000000",
		},
		{
			name: "zero amount",
			money: Money{
				Units:        0,
				Nanos:        0,
				CurrencyCode: "JPY",
			},
			want: "JPY 0.000000000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MoneyToString(tt.money)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_MoneyToFloat64(t *testing.T) {
	tests := []struct {
		name  string
		money Money
		want  float64
	}{
		{
			name: "positive amount",
			money: Money{
				Units:        12,
				Nanos:        345000000,
				CurrencyCode: "USD",
			},
			want: 12.345,
		},
		{
			name: "negative amount",
			money: Money{
				Units:        -5,
				Nanos:        -750000000,
				CurrencyCode: "EUR",
			},
			want: -5.75,
		},
		{
			name: "zero amount",
			money: Money{
				Units:        0,
				Nanos:        0,
				CurrencyCode: "JPY",
			},
			want: 0.0,
		},
		{
			name: "very small fractional",
			money: Money{
				Units:        0,
				Nanos:        1,
				CurrencyCode: "USD",
			},
			want: 0.000000001, // 1 nano
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MoneyToFloat64(tt.money)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_MoneyToInt64(t *testing.T) {
	tests := []struct {
		name  string
		money Money
		want  int64
	}{
		{
			name: "positive amount, round down",
			money: Money{
				Units:        12,
				Nanos:        345000000,
				CurrencyCode: "USD",
			},
			want: 12,
		},
		{
			name: "positive amount, round up",
			money: Money{
				Units:        12,
				Nanos:        745000000,
				CurrencyCode: "USD",
			},
			want: 13,
		},
		{
			name: "negative amount, round toward zero",
			money: Money{
				Units:        -5,
				Nanos:        -250000000,
				CurrencyCode: "EUR",
			},
			want: -5,
		},
		{
			name: "negative amount, round away from zero",
			money: Money{
				Units:        -5,
				Nanos:        -750000000,
				CurrencyCode: "EUR",
			},
			want: -6,
		},
		{
			name: "zero amount",
			money: Money{
				Units:        0,
				Nanos:        0,
				CurrencyCode: "JPY",
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MoneyToInt64(tt.money)
			assert.Equal(t, tt.want, got)
		})
	}
}
