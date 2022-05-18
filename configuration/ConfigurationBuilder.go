package configuration

import (
	"fmt"
	"sync"
)

type ConfigurationBuilder struct {
	isDevelopment       bool
	providers           []ConfigurationProvider
	observableProviders []ObservableConfigurationProvider
}

func NewConfigurationBuilder(development bool) ConfigurationBuilder {
	builder := ConfigurationBuilder{isDevelopment: development}
	return builder
}

func (builder *ConfigurationBuilder) AddConfigurationProvider(provider ConfigurationProvider) *ConfigurationBuilder {
	builder.providers = append(builder.providers, provider)
	return builder
}

func (builder *ConfigurationBuilder) AddObservableConfigurationProvider(provider ObservableConfigurationProvider) *ConfigurationBuilder {
	builder.observableProviders = append(builder.observableProviders, provider)
	return builder
}

func (builder *ConfigurationBuilder) ClearProviders() *ConfigurationBuilder {
	builder.providers = []ConfigurationProvider{}
	return builder
}

func (builder *ConfigurationBuilder) Build(callback func(ConfigurationRoot)) ConfigurationRoot {
	config := ConfigurationRoot{
		onChange: callback,
		mutex:    &sync.Mutex{},
	}

	for _, provider := range builder.providers {
		config.addProvider(provider)
		if builder.isDevelopment {
			fmt.Printf("Loaded provider: %s (%s)\n", provider.Name(), provider.Type())
		}
	}

	for _, provider := range builder.observableProviders {
		config.addObservableProvider(provider)
		if builder.isDevelopment {
			fmt.Printf("Loaded observable provider: %s (%s)\n", provider.Name(), provider.Type())
		}
	}

	return config
}
