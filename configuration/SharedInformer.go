package configuration

import (
	"sync"
	"time"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
)

var (
	informer informers.SharedInformerFactory
	lock     = &sync.Mutex{}
)

func GetInformer() informers.SharedInformerFactory {
	if informer != nil {
		return informer
	}

	// lock ensures that we only ever have one factory
	// the lock is after the initial check as it can be lock free for most cases
	lock.Lock()
	defer lock.Unlock()

	// there's a slight race condition between the first check and the lock,
	// so check again inside a synchronised context
	if informer != nil {
		return informer
	}

	config, err := GetKubernetesConfig()
	if err != nil {
		panic(err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	informer = informers.NewSharedInformerFactory(client, 5*time.Minute)
	return informer
}
