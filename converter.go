package aicost

import "fmt"

// CurrencyAmount represents the amount of currency as a float32 to maintain precision.
type CurrencyAmount float64

// CurrencyConversion defines the methods that any type of currency converter must implement.
type CurrencyConversion interface {
	Convert(amount CurrencyAmount, fromCurrency, toCurrency string) (CurrencyAmount, error)
}

// Converter holds the conversion rates and scale factors for different currencies.
type Converter struct {
	baseCurrency string
	rates        map[string]CurrencyAmount // rates are scaled up to preserve precision
}

// NewConverter initializes a new Converter struct with default rates.
// In a real-world application, you might fetch these rates from a financial API service.
func NewConverter(baseCurrency string, rates map[string]CurrencyAmount) *Converter {
	return &Converter{
		baseCurrency: baseCurrency,
		rates:        rates,
	}
}

// Convert takes an amount in a source currency and converts it to the target currency.
// It returns the converted amount in the target currency.
func (c *Converter) Convert(amount CurrencyAmount, fromCurrency, toCurrency string) (CurrencyAmount, error) {
	// If the source and target currencies are the same, return the amount as is.
	if fromCurrency == toCurrency {
		return amount, nil
	}

	// Convert the amount to the base currency first.
	baseAmount, err := c.convertToBase(amount, fromCurrency)
	if err != nil {
		return 0, err
	}

	// Now convert from the base currency to the target currency.
	targetRate, ok := c.rates[toCurrency]
	if !ok {
		return 0, fmt.Errorf("conversion rate for target currency %s not found", toCurrency)
	}

	convertedAmount := baseAmount * targetRate
	return convertedAmount, nil
}

// convertToBase is a helper function that converts an amount to the base currency.
func (c *Converter) convertToBase(amount CurrencyAmount, currency string) (CurrencyAmount, error) {
	if currency == c.baseCurrency {
		return amount, nil
	}

	rate, ok := c.rates[currency]
	if !ok {
		return 0, fmt.Errorf("conversion rate for currency %s not found", currency)
	}

	// To convert to the base currency, divide by the currency rate.
	return amount / rate, nil
}
