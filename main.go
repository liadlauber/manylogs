package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/TwinProduction/go-color"
	"io"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
)

func GetPodLogs(namespace string, podName string, clientset *kubernetes.Clientset) {
	count := int64(100)
	podLogOptions := corev1.PodLogOptions{
		Follow:    true,
		TailLines: &count,
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
		fmt.Print(color.Ize(color.Blue, message))
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

func main() {
	pods := []string{
		"app", "app2", "app3", "app4",
	}
	clientset, err := getK8sClient()
	if err != nil {

	}
	for _, pod := range pods {
		go GetPodLogs("default", pod, clientset)
	}
	for {
	}
}
