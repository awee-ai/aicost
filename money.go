package aicost

import (
	"fmt"
	"math"
)

type Money struct {
	Units        int64
	Nanos        int32
	CurrencyCode string
}

// NewMoney creates a new money object
func NewMoney(currency string, units int64, nanos int32) *Money {
	return &Money{
		Nanos:        nanos,
		CurrencyCode: currency,
		Units:        units,
	}
}

func (m *Money) Add(n *Money) (*Money, error) {
	if m.CurrencyCode != n.CurrencyCode {
		return nil, fmt.Errorf("currency codes do not match: %s != %s", m.CurrencyCode, n.CurrencyCode)
	}

	totalNanos := m.Nanos + n.Nanos
	extraUnits := totalNanos / 1e9
	remainingNanos := totalNanos % 1e9

	money := NewMoney(m.CurrencyCode, m.Units+n.Units+int64(extraUnits), remainingNanos)

	return money, nil
}

// NewMoneyFromFloat converts a float64 cost to a Money struct.
// It makes the creation of Money instances more human-readable.
func NewMoneyFromFloat(currencyCode string, amount float64) Money {
	units := int64(amount)                                      // Extract whole units
	nanos := int32(math.Round((amount - float64(units)) * 1e9)) // Extract nanos
	return Money{
		CurrencyCode: currencyCode,
		Units:        units,
		Nanos:        nanos,
	}
}

// MoneyToString converts Money to a string representation.
func MoneyToString(m Money) string {
	return fmt.Sprintf("%s %d.%09d", m.CurrencyCode, m.Units, m.Nanos)
}

// MoneyToFloat64 converts Money to a float64 representation.
func MoneyToFloat64(m Money) float64 {
	return float64(m.Units) + float64(m.Nanos)/1e9
}

func MoneyToInt64(m Money) int64 {
	return m.Units + int64(m.Nanos)
}
