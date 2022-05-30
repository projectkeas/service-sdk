package configuration

import (
	"fmt"
	"sync"
)

type ConfigurationRoot struct {
	Providers        []ConfigurationProvider
	onChangeHandlers []func(ConfigurationRoot)
	mutex            *sync.Mutex
}

const SERVICE_NAME string = "Config"

func (config *ConfigurationRoot) addProvider(provider ConfigurationProvider) {
	config.Providers = append(config.Providers, provider)
}

func (config *ConfigurationRoot) addObservableProvider(provider ObservableConfigurationProvider) {
	config.addProvider(provider)

	go observeChanges(config, provider)
}

func (config *ConfigurationRoot) GetStringValueOrDefault(key string, defaultValue string) string {
	for _, provider := range config.Providers {
		found, value := provider.TryGetValue(key)
		if found {
			return value
		}
	}

	return defaultValue
}

func (config *ConfigurationRoot) RegisterChangeNotificationHandler(handler func(ConfigurationRoot)) *ConfigurationRoot {
	config.onChangeHandlers = append(config.onChangeHandlers, handler)
	handler(*config)
	return config
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

			for _, handler := range config.onChangeHandlers {
				handler(*config)
			}
		}()
	}
}
