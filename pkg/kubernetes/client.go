package kubernetes

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func getKubeconfig() (string, error) {
	var kubeconfig string

	kubeconfig = os.Getenv("KUBECONFIG")
	if kubeconfig != "" {
		return kubeconfig, nil
	}

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
		return kubeconfig, nil
	}

	return "", fmt.Errorf("kubeconfig not found, nor KUBECONFIG env was set")
}

func GetK8sClientset() (*kubernetes.Clientset, error) {
	kubeconfig, err := getKubeconfig()
	if err != nil {
		return nil, err
	}
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("building kubeconfig")
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("getting clientset")
	}

	return clientset, nil
}
