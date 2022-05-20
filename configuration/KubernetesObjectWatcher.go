package configuration

import (
	"context"
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	"github.com/projectkeas/sdks-service/logger"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	NS_FILE = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
)

type kubernetesObjectWatcher struct {
	Namespace string
	Type      string
	Name      string
	Channel   chan map[string]string

	data       map[string]string
	client     *kubernetes.Clientset
	mutex      *sync.Mutex
	runningK8s bool
}

func newKubernetesObjectWatcher(objectType string, objectName string) kubernetesObjectWatcher {

	var (
		namespace string
		clientCfg *rest.Config
		client    *kubernetes.Clientset
	)

	nsBytes, err := ioutil.ReadFile(NS_FILE)
	if err != nil {
		fmt.Printf("Unable to read namespace file at %s\n", NS_FILE)
	} else {
		namespace = string(nsBytes)
	}

	clientCfg, err = rest.InClusterConfig()
	if err != nil {
		fmt.Println("Unable to get our client configuration")
	} else {
		client, err = kubernetes.NewForConfig(clientCfg)
		if err != nil {
			fmt.Println("Unable to create our clientset")
		}
	}

	watcher := kubernetesObjectWatcher{
		client:     client,
		mutex:      &sync.Mutex{},
		data:       map[string]string{},
		runningK8s: err == nil,

		Channel:   make(chan map[string]string, 10),
		Name:      objectName,
		Namespace: namespace,
		Type:      objectType,
	}

	return watcher
}

func (watcher *kubernetesObjectWatcher) Watch() {

	if !watcher.runningK8s {
		return
	}

	go watchForChange(*watcher, func(currentWatcher kubernetesObjectWatcher, data map[string]string) {
		watcher.data = data
		watcher.Channel <- data

		if logger.Logger != nil {
			logger.Logger.Debug("Resource Changed.",
				zap.String("name", currentWatcher.Name),
				zap.String("namespace", currentWatcher.Namespace),
				zap.String("type", currentWatcher.Type))
		}
	})
}

func watchForChange(kow kubernetesObjectWatcher, callback func(kubernetesObjectWatcher, map[string]string)) {

	var (
		watcher watch.Interface
		err     error
	)

	// At somepoint the server will close the connection, so we need to loop continually
	for {
		switch kow.Type {
		case "ConfigMap":
			watcher, err = kow.client.CoreV1().ConfigMaps(kow.Namespace).Watch(
				context.TODO(), // provide an empty context
				metav1.SingleObject(metav1.ObjectMeta{
					Name:      kow.Name,
					Namespace: kow.Namespace,
				}))
		case "Secret":
			watcher, err = kow.client.CoreV1().Secrets(kow.Namespace).Watch(
				context.TODO(), // provide an empty context
				metav1.SingleObject(metav1.ObjectMeta{
					Name:      kow.Name,
					Namespace: kow.Namespace,
				}))
		default:
			if logger.Logger != nil {
				logger.Logger.Warn("Unsupported object type. Aborting watch sequence.",
					zap.String("name", kow.Name),
					zap.String("namespace", kow.Namespace),
					zap.String("type", kow.Type),
				)
			}
			return
		}

		if logger.Logger != nil && err != nil {
			logger.Logger.Error("Cannot generate watcher",
				zap.String("name", kow.Name),
				zap.String("namespace", kow.Namespace),
				zap.String("error", err.Error()),
				zap.String("type", kow.Type),
			)

			time.Sleep(5 * time.Second)
		} else if err != nil {
			fmt.Println(err.Error())
		} else if watcher != nil {
			waitForChannelThenPublishUpdate(watcher.ResultChan(), callback, kow)
		}
	}
}

func waitForChannelThenPublishUpdate(eventChannel <-chan watch.Event, callback func(kubernetesObjectWatcher, map[string]string), watcher kubernetesObjectWatcher) {
	for {
		event, open := <-eventChannel
		if open {
			switch event.Type {
			case watch.Added:
				fallthrough
			case watch.Modified:
				watcher.mutex.Lock()

				switch watcher.Type {
				case "ConfigMap":
					if updatedMap, ok := event.Object.(*corev1.ConfigMap); ok {
						callback(watcher, updatedMap.Data)
					}
				case "Secret":
					if updatedSecret, ok := event.Object.(*corev1.Secret); ok {
						callback(watcher, updatedSecret.StringData)
					}
				}

				watcher.mutex.Unlock()
			case watch.Deleted:
				watcher.mutex.Lock()
				// Fall back to the default value
				callback(watcher, map[string]string{})
				watcher.mutex.Unlock()
			default:
				// Do nothing
			}
		} else {
			// If eventChannel is closed, it means the server has closed the connection
			return
		}
	}
}
