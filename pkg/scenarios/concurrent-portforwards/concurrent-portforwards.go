package concurrentportforwards

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	k8s "github.com/ipochi/konnscen/pkg/kubernetes"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

const (
	numberOfConcurrentPortForwards = 5
	contextTimeout                 = 30
	Name                           = "concurrent-portforwards"
)

type ConcurrentPortForwards struct {
	NumberOfConcurrentPortForwards int `yaml:"number_of_concurrent_portforwards"`
	KeepConnectedForSeconds        int `yaml:"keep_connected_for_seconds"`
	StartPort                      int `yaml:"start_port"`
}

type PortForwardAPodRequest struct {
	// RestConfig is the kubernetes config
	RestConfig *rest.Config
	// Pod is the selected pod for this port forwarding
	Pod corev1.Pod
	// LocalPort is the local port that will be selected to expose the PodPort
	LocalPort int
	// PodPort is the target port for the pod
	PodPort int
	// Steams configures where to write or read input from
	Streams genericclioptions.IOStreams
	// StopCh is the channel used to manage the port forward lifecycle
	StopCh <-chan struct{}
	// ReadyCh communicates when the tunnel is ready to receive traffic
	ReadyCh chan struct{}
}

func NewConcurrentPortForwards() *ConcurrentPortForwards {
	return &ConcurrentPortForwards{
		NumberOfConcurrentPortForwards: numberOfConcurrentPortForwards,
	}
}

func (c *ConcurrentPortForwards) Run() error {
	//	if !konnectivity.IsInstalled() {
	//		return fmt.Errorf("Konnectivity Server/Agent, not found")

	//TODO: get metrics of Konnectivity server, before the start of scenario
	var wg sync.WaitGroup
	errChan := make(chan error)
	wgDone := make(chan bool)

	wg.Add(c.NumberOfConcurrentPortForwards)

	d, err := k8s.CreateNginxDeployment()
	if err != nil {
		return err
	}

	for i := 0; i < c.NumberOfConcurrentPortForwards; i++ {
		go getPortForwards(&wg, errChan, c.KeepConnectedForSeconds, c.StartPort+i)
	}

	go func() {
		wg.Wait()
		wgDone <- true
	}()

	isDone := false
	for {
		if isDone {
			if err := k8s.DeleteDeployment(d); err != nil {
				return err
			}
			break
		}

		select {
		case <-wgDone:
			isDone = true
			close(wgDone)
			close(errChan)
		case err = <-errChan:
			close(wgDone)
			close(errChan)
			isDone = true
		default:
		}
	}

	return err
}

func getPortForwards(wg *sync.WaitGroup, errChan chan<- error, connectionSeconds, port int) {
	errCh := make(chan error, 1)
	config, err := k8s.GetRestConfig()
	if err != nil {
		errChan <- fmt.Errorf("getting rest config: %v", err)
		wg.Done()
	}

	cs, err := k8s.GetK8sClientset()
	if err != nil {
		errChan <- fmt.Errorf("getting clientset, %v", err)
		wg.Done()
	}

	stopCh := make(chan struct{}, 1)
	readyCh := make(chan struct{}, 1)

	stream := genericclioptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		errCh <- fmt.Errorf("interrupt recevied")
	}()

	go func() {
		err := <-errCh
		close(stopCh)
		errChan <- err
		wg.Done()
	}()

	pods, err := cs.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{
		LabelSelector: "app=nginx",
	})

	if err != nil {
		errCh <- fmt.Errorf("retreiving all pods in the cluster: %q", err)
	}

	if len(pods.Items) == 0 {
		errCh <- fmt.Errorf("No pods found in the cluster")
	}

	pod := pods.Items[getRandomIndex(len(pods.Items))]

	go func() {
		err = PortForwardAPod(PortForwardAPodRequest{
			RestConfig: config,
			Pod:        pod,
			LocalPort:  port,
			PodPort:    8080,
			Streams:    stream,
			StopCh:     stopCh,
			ReadyCh:    readyCh,
		})
		if err != nil {
			errCh <- fmt.Errorf("could not port forward: %v", err)
		}
	}()

	timeout := time.Second * time.Duration(connectionSeconds)
	isDone := false
	timer := time.NewTimer(timeout)
	for {
		if isDone {
			break
		}
		select {
		case <-timer.C:
			fmt.Println("Reached timeout of keep_connected_for_seconds; stopping")
			errCh <- nil
			isDone = true
			timer.Stop()
		default:
			time.Sleep(3 * time.Second)
			uri := fmt.Sprintf("http://localhost:%d", port)
			resp, err := http.Get(uri)
			if err != nil {
				errCh <- err
				isDone = true
			} else {
				io.Copy(ioutil.Discard, resp.Body)
				resp.Body.Close()
			}
			fmt.Println("I'm curling ....")
		}
	}
}

func PortForwardAPod(req PortForwardAPodRequest) error {
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward",
		req.Pod.Namespace, req.Pod.Name)
	hostIP := strings.TrimPrefix(req.RestConfig.Host, "https://")
	hostIP = strings.TrimSuffix(hostIP, "/")

	transport, upgrader, err := spdy.RoundTripperFor(req.RestConfig)
	if err != nil {
		return err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, &url.URL{Scheme: "https", Path: path, Host: hostIP})
	fw, err := portforward.New(dialer, []string{fmt.Sprintf("%d:%d", req.LocalPort, req.PodPort)}, req.StopCh, req.ReadyCh, req.Streams.Out, req.Streams.ErrOut)
	if err != nil {
		return err
	}
	return fw.ForwardPorts()
}

func (c *ConcurrentPortForwards) Cleanup() error {

	return nil
}

func getRandomIndex(max int) int {
	rand.Seed(time.Now().UnixNano())

	return rand.Intn(max)
}
