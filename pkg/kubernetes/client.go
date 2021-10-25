package kubernetes

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/yaml"
)

const (
	deployRunningThreshold     = time.Second * 60
	deployRunningCheckInterval = time.Second * 2
	nginxManifest              = `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: nginx
  name: nginx
  namespace: default
spec:
  replicas: 10
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - image: bitnami/nginx
        name: nginx
`
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

func GetRestConfig() (*rest.Config, error) {
	kubeconfig, err := getKubeconfig()
	if err != nil {
		return nil, err
	}
	// use the current context in kubeconfig
	return clientcmd.BuildConfigFromFlags("", kubeconfig)
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

func CreateNginxDeployment() (*appsv1.Deployment, error) {
	cs, err := GetK8sClientset()
	if err != nil {
		return nil, fmt.Errorf("getting clientset, %v", err)
	}

	// Create a deployment for port-forwarding
	d := &appsv1.Deployment{}
	if err := yaml.Unmarshal([]byte(nginxManifest), d); err != nil {
		return nil, fmt.Errorf("failed to unmarshal deployment manifest: %v", err)
	}

	if _, err := cs.AppsV1().Deployments("default").Create(context.TODO(), d, metav1.CreateOptions{}); err != nil {
		return nil, fmt.Errorf("failed to create nginx deployment: %v", err)
	}

	fmt.Println("nginx Deployment created")

	if err := waitForPodsRunning(cs, "app=nginx"); err != nil {
		return nil, fmt.Errorf("timed out waiting for pods to be in Running state: %v", err)
	}

	fmt.Println("nginx pods in Running state, continuing")

	return d, nil
}

func DeleteDeployment(d *appsv1.Deployment) error {
	cs, err := GetK8sClientset()
	if err != nil {
		return fmt.Errorf("getting clientset, %v", err)
	}

	if err := cs.AppsV1().Deployments("default").Delete(context.TODO(), d.Name, metav1.DeleteOptions{}); err != nil {
		return fmt.Errorf("failed to create nginx deployment: %v", err)
	}

	return nil
}

func waitForPodsRunning(cs *kubernetes.Clientset, label string) error {
	end := time.Now().Add(deployRunningThreshold)

	for true {
		<-time.NewTimer(deployRunningCheckInterval).C

		var err error
		running, err := allPodsRunning(cs, label)
		if running {
			return nil
		}

		if err != nil {
			println(fmt.Sprintf("Encountered an error checking for running pods: %s", err))
		}

		if time.Now().After(end) {
			return fmt.Errorf("Failed to get all running containers")
		}
	}
	return nil
}

func allPodsRunning(cs *kubernetes.Clientset, label string) (bool, error) {
	pods, err := cs.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{
		LabelSelector: label,
	})

	if err != nil {
		return false, err
	}

	for _, pod := range pods.Items {
		if pod.Status.Phase != corev1.PodRunning {
			return false, nil
		}
	}

	return true, nil
}
