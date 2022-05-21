package configuration

type InMemoryConfigurationProvider struct {
	name string
	data map[string]string
}

func NewInMemoryConfigurationProvider(name string, data map[string]string) *InMemoryConfigurationProvider {
	provider := &InMemoryConfigurationProvider{
		data: data,
		name: name,
	}
	return provider
}

func (provider InMemoryConfigurationProvider) Name() string {
	return provider.name
}
func (provider InMemoryConfigurationProvider) Type() string {
	return "InMemory"
}

func (provider InMemoryConfigurationProvider) TryGetValue(key string) (bool, string) {
	data, found := provider.data[key]

	if found {
		return true, data
	}

	return false, ""
}
