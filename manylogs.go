package main

import (
	"context"
	"flag"
	"gopkg.in/gookit/color.v1"
	"io"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	typev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"math/rand"
	"path/filepath"
	"strings"
	"time"
)

func getPodLogs(namespace string, podName string, containerName string, clientset *kubernetes.Clientset, podColor color.Color256, ch chan string) {
	count := int64(100)
	podLogOptions := corev1.PodLogOptions{
		Follow:    true,
		TailLines: &count,
		Container: containerName,
	}

	podLogRequest := clientset.CoreV1().
		Pods(namespace).
		GetLogs(podName, &podLogOptions)
	stream, err := podLogRequest.Stream(context.TODO())
	if err != nil {
		println("error in opening stream")
		panic(err.Error())
	}
	defer func() {
		err = stream.Close()
		if err != nil {
			println("error in closing stream")
			panic(err.Error())
		}
		ch <- "Done"
	}()

	for {
		buf := make([]byte, 2000)
		numBytes, err := stream.Read(buf)
		if numBytes == 0 {
			continue
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err.Error())
		}
		message := string(buf[:numBytes])
		podColor.Printf(podName + ":\t" + message)
	}
	panic(err.Error())
}

func getK8sClient() (*kubernetes.Clientset, error) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	return kubernetes.NewForConfig(config)
}

func getLabeledPods(label string, namespace string, k8sClient typev1.CoreV1Interface) (*corev1.PodList, error) {
	podsLabel := map[string]string{strings.Split(label, "=")[0]: strings.Split(label, "=")[1]}
	set := labels.Set(podsLabel)
	listOptions := metav1.ListOptions{LabelSelector: set.AsSelector().String()}
	pods, err := k8sClient.Pods(namespace).List(context.TODO(), listOptions)
	return pods, err
}

func main() {
	//namespace := "openshift-etcd"
	//namespace := "sns-system"
	//label := "k8s-app=etcd-quorum-guard"
	//label := "control-plane=controller-manager"
	//namespace := "labertest"
	//label := "app=laber"
	//container := ""

	namespace := flag.String("namespace", "labertest", "a namespace to get logs from")
	label := flag.String("label", "app=laber", "a label matches the pods to get logs from")
	container := flag.String("container", "", "a container to get logs from")

	ch := make(chan string)

	rand.Seed(time.Now().UnixNano())

	clientset, err := getK8sClient()
	if err != nil {
		panic(err.Error())
	}

	pods, err := getLabeledPods(*label, *namespace, clientset.CoreV1())
	if err != nil {
		panic(err.Error())
	}

	for _, pod := range pods.Items {
		podColor := color.C256(uint8(rand.Intn(256)))
		go getPodLogs(*namespace, pod.Name, *container, clientset, podColor, ch)
	}
	println(<-ch)
}
