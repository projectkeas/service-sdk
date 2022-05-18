package configuration

type KubernetesConfigMapConfigurationProvider struct {
	name    string
	watcher kubernetesObjectWatcher
}

func NewKubernetesConfigMapConfigurationProvider(name string) *KubernetesConfigMapConfigurationProvider {
	provider := &KubernetesConfigMapConfigurationProvider{
		name:    name,
		watcher: newKubernetesObjectWatcher("ConfigMap", name),
	}
	provider.watcher.Watch()
	return provider
}

func (provider *KubernetesConfigMapConfigurationProvider) Name() string {
	return provider.name
}

func (provider *KubernetesConfigMapConfigurationProvider) Type() string {
	return "KubernetesConfigMap"
}

func (provider *KubernetesConfigMapConfigurationProvider) TryGetValue(key string) (bool, string) {

	data, found := provider.watcher.data[key]

	if found {
		return true, data
	}

	return false, ""
}

func (provider *KubernetesConfigMapConfigurationProvider) getChannel() chan map[string]string {
	return provider.watcher.Channel
}
