package configuration

import (
	"time"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"

	log "github.com/projectkeas/sdks-service/logger"
	types "k8s.io/api/core/v1"
)

type KubernetesSecretConfigurationProvider struct {
	name          string
	data          map[string]string
	updateChannel chan map[string]string
	Exists        bool
}

func NewKubernetesSecretConfigurationProvider(name string) *KubernetesSecretConfigurationProvider {
	provider := &KubernetesSecretConfigurationProvider{
		name:          name,
		data:          map[string]string{},
		updateChannel: make(chan map[string]string),
	}

	informer := GetInformer()
	configInformer := informer.Core().V1().Secrets().Informer()

	configInformer.AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc:    onNewSecret(provider),
		UpdateFunc: onUpdatedSecret(provider),
		DeleteFunc: onDeletedSecret(provider),
	}, 2*time.Minute)

	informer.Start(wait.NeverStop)
	informer.WaitForCacheSync(wait.NeverStop)

	return provider
}

func (provider *KubernetesSecretConfigurationProvider) Name() string {
	return provider.name
}

func (provider *KubernetesSecretConfigurationProvider) Type() string {
	return "KubernetesSecret"
}

func (provider *KubernetesSecretConfigurationProvider) TryGetValue(key string) (bool, string) {

	data, found := provider.data[key]

	if found {
		return true, data
	}

	return false, ""
}

func (provider *KubernetesSecretConfigurationProvider) getChannel() chan map[string]string {
	return provider.updateChannel
}

func onNewSecret(provider *KubernetesSecretConfigurationProvider) func(newSecret interface{}) {
	return func(newSecret interface{}) {
		secret, successfulCast := newSecret.(*types.Secret)
		if successfulCast && secret.Name == provider.name {
			addOrUpdateSecret(provider, secret)
			if log.Logger != nil {
				log.Logger.Debug("Secret added", zap.Any("secret", map[string]string{
					"name":      secret.Name,
					"namespace": secret.Namespace,
				}))
			}
		} else if !successfulCast {
			log.Logger.Error("could not cast config map")
		}
	}
}

func onUpdatedSecret(provider *KubernetesSecretConfigurationProvider) func(oldSecret interface{}, newSecret interface{}) {
	return func(oldSecret interface{}, newSecret interface{}) {
		secret, successfulCast := newSecret.(*types.Secret)
		if successfulCast && secret.Name == provider.name {
			addOrUpdateSecret(provider, secret)
			if log.Logger != nil {
				log.Logger.Debug("Secret updated", zap.Any("secret", map[string]string{
					"name":      secret.Name,
					"namespace": secret.Namespace,
				}))
			}
		} else if !successfulCast {
			log.Logger.Error("could not cast config map")
		}
	}
}

func addOrUpdateSecret(provider *KubernetesSecretConfigurationProvider, secret *types.Secret) {

	if secret.StringData != nil {
		provider.data = secret.StringData
	} else {
		data := map[string]string{}
		for key, value := range secret.Data {
			data[key] = string(value)
		}

		provider.data = data
	}

	if provider.data == nil {
		provider.data = map[string]string{}
	}

	provider.Exists = true
	provider.updateChannel <- provider.data
}

func onDeletedSecret(provider *KubernetesSecretConfigurationProvider) func(deletedSecret interface{}) {
	return func(deletedSecret interface{}) {
		secret, successfulCast := deletedSecret.(*types.Secret)
		if successfulCast && secret.Name == provider.name {
			provider.data = map[string]string{}
			provider.Exists = false
			if log.Logger != nil {
				log.Logger.Debug("Secret deleted", zap.Any("secret", map[string]string{
					"name":      secret.Name,
					"namespace": secret.Namespace,
				}))
			}
			provider.updateChannel <- provider.data
		} else if !successfulCast {
			log.Logger.Error("could not cast config map")
		}
	}
}
