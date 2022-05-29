package configuration

import (
	"flag"
	"os"
	"path/filepath"
	"sync"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const (
	K8S_NS_FILE = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
)

var (
	config     *rest.Config
	configLock = &sync.Mutex{}
)

func GetKubernetesConfig() (*rest.Config, error) {

	if config != nil {
		return config, nil
	}

	// lock ensures that we only ever have one config file
	// the lock is after the initial check as it can be lock free for most cases
	configLock.Lock()
	defer configLock.Unlock()

	// there's a slight race condition between the first check and the lock,
	// so check again inside a synchronised context
	if config != nil {
		return config, nil
	}

	var err error

	if _, err = os.Stat(K8S_NS_FILE); err == nil {
		config, err = rest.InClusterConfig()
		return config, err
	}

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	return config, err
}
