package aicost

import (
	"errors"
	"fmt"
	"math"
)

// Money represents a monetary value
// can be:
// cost per single token
// cost per 1k tokens
// cost per custom amount of tokens
// cost per character
type Money struct {
	// Units is the whole units of the currency
	Units int64 `json:"units" yaml:"units"`
	// Nanos is the fractional part of the currency, in nanoseconds
	Nanos int32 `json:"nanos" yaml:"nanos"`
	// CurrencyCode is the ISO 4217 currency code
	CurrencyCode string `json:"currency_code" yaml:"currency_code"`
}

// NewMoney creates a new money object with validation
func NewMoney(currency string, units int64, nanos int32) (*Money, error) {
	// Validate that nanos is within range
	if nanos < -999999999 || nanos > 999999999 {
		return nil, errors.New("nanos must be between -999999999 and 999999999")
	}

	// Validate that units and nanos have the same sign
	if (units < 0 && nanos > 0) || (units > 0 && nanos < 0) {
		return nil, fmt.Errorf("units and nanos must have the same sign: units[%d], nanos[%d]", units, nanos)
	}

	if currency == "" {
		return nil, errors.New("currency code cannot be empty")
	}

	return &Money{
		Nanos:        nanos,
		CurrencyCode: currency,
		Units:        units,
	}, nil
}

// NewMoneyUnsafe creates a new money object without validation
// Only use when you're certain the inputs are valid
func NewMoneyUnsafe(currency string, units int64, nanos int32) *Money {
	return &Money{
		Nanos:        nanos,
		CurrencyCode: currency,
		Units:        units,
	}
}

// Add adds two Money objects together
func (m *Money) Add(n *Money) (*Money, error) {
	if m.CurrencyCode != n.CurrencyCode {
		return nil, fmt.Errorf("currency codes do not match: %s != %s", m.CurrencyCode, n.CurrencyCode)
	}

	// Calculate total in nanos to avoid sign issues
	totalNanos := int64(m.Units)*1e9 + int64(m.Nanos) + int64(n.Units)*1e9 + int64(n.Nanos)

	// Convert back to units and nanos
	units := totalNanos / 1e9
	nanos := int32(totalNanos % 1e9)

	return NewMoney(m.CurrencyCode, units, nanos)
}

// Times multiplies Money by an integer factor
func (m *Money) Times(n int64) (*Money, error) {
	if n == 0 {
		return NewMoney(m.CurrencyCode, 0, 0)
	}

	// Check for overflow in units
	if m.Units != 0 && n > math.MaxInt64/m.Units {
		return nil, errors.New("integer overflow in units multiplication")
	}

	// Calculate total in nanos to handle sign correctly
	totalNanos := int64(m.Units)*1e9*n + int64(m.Nanos)*n

	// Convert back to units and nanos
	units := totalNanos / 1e9
	nanos := int32(totalNanos % 1e9)

	return NewMoney(m.CurrencyCode, units, nanos)
}

// TimesFloat multiplies Money by a floating-point factor
func (m *Money) TimesFloat(rate float64) (*Money, error) {
	if rate == 0 {
		money, err := NewMoney(m.CurrencyCode, 0, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to create zero rate float money: %w", err)
		}
		return money, nil
	}

	// calculate total in nanos to handle sign correctly
	totalNanos := (float64(m.Units)*1e9 + float64(m.Nanos)) * rate

	// check for overflow
	if totalNanos > math.MaxInt64 || totalNanos < math.MinInt64 {
		return nil, errors.New("overflow in float multiplication")
	}

	roundedNanos := math.Round(totalNanos)
	// convert back to units and nanos
	units := int64(roundedNanos / 1e9)
	nanos := int32(roundedNanos - float64(units*1e9))

	money, err := NewMoney(m.CurrencyCode, units, nanos)
	if err != nil {
		return nil, fmt.Errorf("failed to create money by multiplying by float: %w", err)
	}

	return money, nil
}

// NewMoneyFromFloat converts a float64 cost to a Money struct.
// it makes the creation of Money instances more human-readable.
func NewMoneyFromFloat(currencyCode string, amount float64) (*Money, error) {
	// handle potential floating-point precision issues by working with a scaled integer
	scaledAmount := amount * 1e9

	units := int64(scaledAmount / 1e9)
	nanos := int32(scaledAmount - float64(units)*1e9)

	money, err := NewMoney(currencyCode, units, nanos)
	if err != nil {
		return nil, fmt.Errorf("failed to create money from float: %w", err)
	}
	return money, err
}

// MoneyToString converts Money to a string representation.
func MoneyToString(m Money) string {
	return fmt.Sprintf("%s %d.%09d", m.CurrencyCode, m.Units, int(math.Abs(float64(m.Nanos))))
}

// MoneyToFloat64 converts Money to a float64 representation.
func MoneyToFloat64(m Money) float64 {
	sign := 1.0
	if m.Units < 0 || m.Nanos < 0 {
		sign = -1.0
	}
	return sign * (math.Abs(float64(m.Units)) + math.Abs(float64(m.Nanos))/1e9)
}

// MoneyToInt64 converts Money to an int64 representation with proper rounding.
func MoneyToInt64(m Money) int64 {
	// Calculate with proper rounding
	totalNanos := float64(m.Units)*1e9 + float64(m.Nanos)
	rounded := math.Round(totalNanos / 1e9)
	return int64(rounded)
}
