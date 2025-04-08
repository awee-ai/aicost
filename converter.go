package aicost

import (
	"fmt"
)

// Converter defines the methods that any type of currency converter must implement.
type Converter interface {
	Convert(amount Money, toCurrency string) (*Money, error)
	Rates(rates map[string]float64) error
}

// converter holds the conversion rates and scale factors for different currencies.
type converter struct {
	baseCurrency string
	rates        map[string]float64
}

var _ Converter = (*converter)(nil)

// NewConverter initializes a new converter struct with default rates.
func NewConverter(baseCurrency string, rates map[string]float64) *converter {
	return &converter{
		baseCurrency: baseCurrency,
		rates:        rates,
	}
}

// Rates sets the conversion rates for the converter.
func (c *converter) Rates(rates map[string]float64) error {
	if len(rates) == 0 {
		return fmt.Errorf("conversion rates cannot be empty")
	}

	for cur, rate := range rates {
		if rate <= 0 {
			return fmt.Errorf("conversion rate %s must be greater than 0: %f", cur, rate)
		}
	}

	c.rates = rates

	return nil
}

// Convert takes an amount in a source currency and converts it to the target currency
// it returns the converted amount in the target currency
func (c *converter) Convert(providedMoney Money, toCurrency string) (*Money, error) {
	// if the source and target currencies are the same, return the amount as is
	if providedMoney.CurrencyCode == toCurrency {
		return &providedMoney, nil
	}

	var baseAmount = &providedMoney

	// if the base currency is not the same as the source currency
	// convert the amount to the base currency first
	if providedMoney.CurrencyCode != c.baseCurrency {
		var err error
		baseAmount, err = c.convertToBase(providedMoney)
		if err != nil {
			return nil, fmt.Errorf("error converting to base currency: %w", err)
		}
	}

	if baseAmount.CurrencyCode == toCurrency {
		return baseAmount, nil
	}

	// convert from the base currency to the target currency
	// if the target currency is the base currency, flip the rate.
	targetRate, ok := c.rates[toCurrency]
	if !ok {
		return nil, fmt.Errorf("conversion rate for currency %s not found", toCurrency)
	}

	converted, err := baseAmount.TimesFloat(targetRate)
	if err != nil {
		return nil, fmt.Errorf("error converting to target currency: %w", err)
	}
	converted.CurrencyCode = toCurrency

	return converted, nil
}

// convertToBase is a helper function that converts an amount to the base currency.
func (c *converter) convertToBase(amount Money) (*Money, error) {
	if amount.CurrencyCode == c.baseCurrency {
		return &amount, nil
	}

	rate, ok := c.rates[amount.CurrencyCode]
	if !ok {
		return nil, fmt.Errorf("conversion rate for currency %s not found", amount.CurrencyCode)
	}
	// invert the rate
	rate = 1 + (1 - rate)

	// To convert to the base currency, divide by the currency rate.
	float, err := amount.TimesFloat(rate)
	if err != nil {
		return nil, fmt.Errorf("error converting to base currency: %w", err)
	}
	float.CurrencyCode = c.baseCurrency

	return float, nil
}
