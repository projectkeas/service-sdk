package configuration

import (
	"fmt"
	"sync"
)

type ConfigurationRoot struct {
	providers []ConfigurationProvider
	onChange  func(ConfigurationRoot)
	mutex     *sync.Mutex
}

func (config *ConfigurationRoot) addProvider(provider ConfigurationProvider) {
	config.providers = append(config.providers, provider)
}

func (config *ConfigurationRoot) addObservableProvider(provider ObservableConfigurationProvider) {
	config.providers = append(config.providers, provider)

	go observeChanges(config, provider)
}

func (config *ConfigurationRoot) GetStringValueOrDefault(key string, defaultValue string) string {
	for _, provider := range config.providers {
		found, value := provider.TryGetValue(key)
		if found {
			return value
		}
	}

	return defaultValue
}

func observeChanges(config *ConfigurationRoot, provider ObservableConfigurationProvider) {
	for {
		func() {
			<-provider.getChannel()
			config.mutex.Lock()

			defer func() {
				if err := recover(); err != nil {
					fmt.Println("panic at the disco!", err)
				}

				config.mutex.Unlock()
			}()

			config.onChange(*config)
		}()
	}
}
