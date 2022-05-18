package configuration

type ConfigurationProvider interface {
	Name() string
	Type() string
	TryGetValue(key string) (bool, string)
}

type ObservableConfigurationProvider interface {
	Name() string
	Type() string
	TryGetValue(key string) (bool, string)

	getChannel() chan map[string]string
}
