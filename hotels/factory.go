package hotels

import (
	"fmt"
	"os"
)

// ProviderType identifies a hotels provider implementation.
type ProviderType string

const (
	ProviderAmadeus ProviderType = "amadeus"
	ProviderMock    ProviderType = "mock"
)

// ProviderFactory selects a hotels provider from env / explicit type.
type ProviderFactory struct {
	defaultProvider ProviderType
}

// NewProviderFactory builds a factory using HOTELS_PROVIDER (defaults to amadeus).
func NewProviderFactory() *ProviderFactory {
	p := os.Getenv("HOTELS_PROVIDER")
	if p == "" {
		p = string(ProviderAmadeus)
	}
	return &ProviderFactory{defaultProvider: ProviderType(p)}
}

// GetProvider returns a provider for the given type.
func (f *ProviderFactory) GetProvider(t ProviderType) (Provider, error) {
	if t == "" {
		t = f.defaultProvider
	}
	switch t {
	case ProviderAmadeus:
		return NewAmadeusHotelProvider(), nil
	case ProviderMock:
		return NewMockHotelProvider(), nil
	default:
		return nil, fmt.Errorf("hotels: unsupported provider %q", t)
	}
}

// GetDefaultProvider returns the configured default.
func (f *ProviderFactory) GetDefaultProvider() (Provider, error) {
	return f.GetProvider(f.defaultProvider)
}

// DefaultProviderType returns the configured default type.
func (f *ProviderFactory) DefaultProviderType() ProviderType {
	return f.defaultProvider
}
