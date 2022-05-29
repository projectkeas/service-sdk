package configuration

import (
	"fmt"
	"sync"
)

type ConfigurationRoot struct {
	Providers []ConfigurationProvider
	onChange  func(ConfigurationRoot)
	mutex     *sync.Mutex
}

const SERVICE_NAME string = "Config"

func (config *ConfigurationRoot) addProvider(provider ConfigurationProvider) {
	config.Providers = append(config.Providers, provider)
}

func (config *ConfigurationRoot) addObservableProvider(provider ObservableConfigurationProvider) {
	config.Providers = append(config.Providers, provider)

	go observeChanges(config, provider)
}

func (config ConfigurationRoot) GetStringValueOrDefault(key string, defaultValue string) string {
	for _, provider := range config.Providers {
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
