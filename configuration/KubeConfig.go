package configuration

import (
	"flag"
	"os"
	"path/filepath"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const (
	K8S_NS_FILE = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
)

func GetKubernetesConfig() (*rest.Config, error) {

	if _, err := os.Stat(K8S_NS_FILE); err == nil {
		return rest.InClusterConfig()
	}

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	return clientcmd.BuildConfigFromFlags("", *kubeconfig)
}
