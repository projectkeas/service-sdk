package configuration

import (
	"flag"
	"io/ioutil"
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
	namespace  = "keas"
)

func GetKubernetesConfig() (*rest.Config, string, error) {

	if config != nil {
		return config, namespace, nil
	}

	// lock ensures that we only ever have one config file
	// the lock is after the initial check as it can be lock free for most cases
	configLock.Lock()
	defer configLock.Unlock()

	// there's a slight race condition between the first check and the lock,
	// so check again inside a synchronised context
	if config != nil {
		return config, namespace, nil
	}

	var err error

	if _, err = os.Stat(K8S_NS_FILE); err == nil {
		ns, _ := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
		namespace = string(ns)
		config, err = rest.InClusterConfig()
		return config, namespace, err
	}

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	return config, namespace, err
}
