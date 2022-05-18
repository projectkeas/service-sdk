package configuration

type KubernetesSecretConfigurationProvider struct {
	name    string
	watcher kubernetesObjectWatcher
}

func NewKubernetesSecretConfigurationProvider(name string) *KubernetesSecretConfigurationProvider {
	provider := &KubernetesSecretConfigurationProvider{
		name:    name,
		watcher: newKubernetesObjectWatcher("Secret", name),
	}
	provider.watcher.Watch()
	return provider
}

func (provider *KubernetesSecretConfigurationProvider) Name() string {
	return provider.name
}

func (provider *KubernetesSecretConfigurationProvider) Type() string {
	return "KubernetesSecret"
}

func (provider *KubernetesSecretConfigurationProvider) TryGetValue(key string) (bool, string) {
	data, found := provider.watcher.data[key]

	if found {
		return true, data
	}

	return false, ""
}

func (provider *KubernetesSecretConfigurationProvider) getChannel() chan map[string]string {
	return provider.watcher.Channel
}
