package concurrentconnections

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"sync"
	"time"

	k8s "github.com/ipochi/konnscen/pkg/kubernetes"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	numberOfConcurrentUsers = 5
	numberOfTimes           = 10
	contextTimeout          = 30
	Name                    = "concurrent-connections"
)

type ConcurrentConnections struct {
	NumberOfConcurrentUsers int `yaml:"number_of_concurrent_users"`
	NumberOfTimes           int `yaml:"number_of_times"`
}

func NewConcurrentConnections() *ConcurrentConnections {
	return &ConcurrentConnections{
		NumberOfConcurrentUsers: numberOfConcurrentUsers,
		NumberOfTimes:           numberOfTimes,
	}
}

func (c *ConcurrentConnections) Run() error {
	//	if !konnectivity.IsInstalled() {
	//		return fmt.Errorf("Konnectivity Server/Agent, not found")

	//TODO: get metrics of Konnectivity server, before the start of scenario
	ctx := context.Background()
	var wg sync.WaitGroup
	errChan := make(chan error)
	wgDone := make(chan bool)
	errCount := 0
	wg.Add(c.NumberOfConcurrentUsers)
	for i := 0; i < c.NumberOfConcurrentUsers; i++ {
		go c.getLogs(ctx, &wg, errChan)
	}

	go func() {
		wg.Wait()
		wgDone <- true
	}()

	isDone := false
	for {
		if isDone {
			break
		}

		select {
		case <-wgDone:
			isDone = true
			close(wgDone)
			close(errChan)
		case err := <-errChan:
			errCount++
			fmt.Println(err)
		default:
		}
	}

	fmt.Println("Total errors when processing logs --- ", errCount)

	return nil
}

func (c *ConcurrentConnections) getLogs(ctx context.Context, wg *sync.WaitGroup, errChan chan<- error) {
	cs, err := k8s.GetK8sClientset()
	if err != nil {
		errChan <- fmt.Errorf("getting clientset, %v", err)
	}

	defer wg.Done()
	for i := 0; i < c.NumberOfTimes; i++ {
		randomSleep(15)
		podLogOpts := corev1.PodLogOptions{}

		// Get all the pods in the cluster.
		pods, err := cs.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			errChan <- fmt.Errorf("retreiving all pods in the cluster: %q", err)
			continue
		}

		if len(pods.Items) == 0 {
			errChan <- fmt.Errorf("No pods found in the cluster")
			continue
		}

		index := getRandomIndex(len(pods.Items))
		pod := pods.Items[index]

		// To avoid error in getting logs of a pod with multiple containers.
		// Default to the first container in the pod.Spec.Container list.
		if len(pod.Spec.Containers) > 1 {
			podLogOpts.Container = pod.Spec.Containers[0].Name
		}

		req := cs.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &podLogOpts)
		podLogs, err := req.Stream(ctx)

		if err != nil {
			errChan <- fmt.Errorf("opening stream: %q", err)
			continue
		}

		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, podLogs)
		if err != nil {
			errChan <- fmt.Errorf("copying information from podLogs to buf: %q", err)
			continue
		}
		str := buf.String()
		fmt.Println(str)
		podLogs.Close()
	}
}

func randomSleep(seconds int) {
	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(seconds)
	fmt.Printf("Sleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	fmt.Println("Done")
}

func getRandomIndex(max int) int {
	rand.Seed(time.Now().UnixNano())

	return rand.Intn(max)
}

func (c *ConcurrentConnections) Cleanup() error {

	return nil
}
