package configuration

import (
	"os"
	"strings"
)

type EnvironmentConfigurationProvider struct {
	prefix string
	data   map[string]string
}

func NewEnvironmentConfigurationProvider(prefix string) *EnvironmentConfigurationProvider {
	data := map[string]string{}

	for _, env := range os.Environ() {
		envPair := strings.SplitN(env, "=", 2)
		key := envPair[0]
		value := envPair[1]

		if strings.HasPrefix(key, prefix) {
			// normalize the format of the keys
			key = strings.TrimPrefix(key, prefix)
			key = strings.Replace(key, "_", ".", -1)
			key = strings.Replace(key, "-", ".", -1)
			key = strings.Replace(key, "..", ".", -1)
			data[strings.ToLower(key)] = strings.TrimSpace(value)
		}
	}

	provider := &EnvironmentConfigurationProvider{
		data:   data,
		prefix: prefix,
	}
	return provider
}

func (provider EnvironmentConfigurationProvider) Name() string {
	return provider.prefix
}
func (provider EnvironmentConfigurationProvider) Type() string {
	return "InMemory"
}

func (provider EnvironmentConfigurationProvider) TryGetValue(key string) (bool, string) {
	data, found := provider.data[key]

	if found {
		return true, data
	}

	return false, ""
}
