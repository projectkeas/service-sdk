package configuration

import (
	"time"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"

	log "github.com/projectkeas/sdks-service/logger"
	types "k8s.io/api/core/v1"
)

type KubernetesConfigMapConfigurationProvider struct {
	name          string
	data          map[string]string
	updateChannel chan map[string]string
	Exists        bool
}

func NewKubernetesConfigMapConfigurationProvider(name string) *KubernetesConfigMapConfigurationProvider {
	provider := &KubernetesConfigMapConfigurationProvider{
		name:          name,
		data:          map[string]string{},
		updateChannel: make(chan map[string]string),
	}

	informer := GetInformer()
	configInformer := informer.Core().V1().ConfigMaps().Informer()

	configInformer.AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc:    onNewConfigMap(provider),
		UpdateFunc: onUpdatedConfigMap(provider),
		DeleteFunc: onDeletedConfigMap(provider),
	}, 2*time.Minute)

	informer.Start(wait.NeverStop)
	informer.WaitForCacheSync(wait.NeverStop)

	return provider
}

func (provider *KubernetesConfigMapConfigurationProvider) Name() string {
	return provider.name
}

func (provider *KubernetesConfigMapConfigurationProvider) Type() string {
	return "KubernetesConfigMap"
}

func (provider *KubernetesConfigMapConfigurationProvider) TryGetValue(key string) (bool, string) {

	data, found := provider.data[key]

	if found {
		return true, data
	}

	return false, ""
}

func (provider *KubernetesConfigMapConfigurationProvider) getChannel() chan map[string]string {
	return provider.updateChannel
}

func onNewConfigMap(provider *KubernetesConfigMapConfigurationProvider) func(newConfigMap interface{}) {
	return func(newConfigMap interface{}) {
		configMap, successfulCast := newConfigMap.(*types.ConfigMap)
		if successfulCast && configMap.Name == provider.name {
			addOrUpdateConfigMap(provider, configMap)
			if log.Logger != nil {
				log.Logger.Debug("ConfigMap added", zap.Any("configMap", map[string]string{
					"name":      configMap.Name,
					"namespace": configMap.Namespace,
				}))
			}
		} else if !successfulCast {
			log.Logger.Error("could not cast config map")
		}
	}
}

func onUpdatedConfigMap(provider *KubernetesConfigMapConfigurationProvider) func(oldConfigMap interface{}, newConfigMap interface{}) {
	return func(oldConfigMap interface{}, newConfigMap interface{}) {
		configMap, successfulCast := newConfigMap.(*types.ConfigMap)
		if successfulCast && configMap.Name == provider.name {
			addOrUpdateConfigMap(provider, configMap)
			if log.Logger != nil {
				log.Logger.Debug("ConfigMap updated", zap.Any("configMap", map[string]string{
					"name":      configMap.Name,
					"namespace": configMap.Namespace,
				}))
			}
		} else if !successfulCast {
			log.Logger.Error("could not cast config map")
		}
	}
}

func addOrUpdateConfigMap(provider *KubernetesConfigMapConfigurationProvider, configMap *types.ConfigMap) {
	provider.data = configMap.Data
	provider.Exists = true
	provider.updateChannel <- provider.data
}

func onDeletedConfigMap(provider *KubernetesConfigMapConfigurationProvider) func(deletedConfigMap interface{}) {
	return func(deletedConfigMap interface{}) {
		configMap, successfulCast := deletedConfigMap.(*types.ConfigMap)
		if successfulCast && configMap.Name == provider.name {
			provider.data = map[string]string{}
			provider.Exists = false
			if log.Logger != nil {
				log.Logger.Debug("ConfigMap deleted", zap.Any("configMap", map[string]string{
					"name":      configMap.Name,
					"namespace": configMap.Namespace,
				}))
			}
			provider.updateChannel <- provider.data
		} else if !successfulCast {
			log.Logger.Error("could not cast config map")
		}
	}
}
