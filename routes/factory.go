package routes

import (
	"fmt"
	"os"
)

// ProviderType identifies a routes provider implementation.
type ProviderType string

const (
	ProviderGoogle ProviderType = "google"
	ProviderMapbox ProviderType = "mapbox"
	ProviderMock   ProviderType = "mock"
)

// ProviderFactory selects routes providers from env / explicit type.
type ProviderFactory struct {
	defaultProvider ProviderType
}

// NewProviderFactory builds a factory using ROUTES_PROVIDER (defaults to google).
func NewProviderFactory() *ProviderFactory {
	p := os.Getenv("ROUTES_PROVIDER")
	if p == "" {
		p = string(ProviderGoogle)
	}
	return &ProviderFactory{defaultProvider: ProviderType(p)}
}

// GetProvider returns a provider for the given type.
func (f *ProviderFactory) GetProvider(t ProviderType) (Provider, error) {
	if t == "" {
		t = f.defaultProvider
	}
	switch t {
	case ProviderGoogle:
		return NewGoogleRoutesProvider(), nil
	case ProviderMapbox:
		return NewMapboxRouteProvider(), nil
	case ProviderMock:
		return NewMockRouteProvider(), nil
	default:
		return nil, fmt.Errorf("routes: unsupported provider %q", t)
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
