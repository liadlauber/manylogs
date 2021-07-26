package main

import (
	"context"
	"flag"
	"gopkg.in/gookit/color.v1"
	"io"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"math/rand"
	"path/filepath"
	"time"
)

func getPodLogs(namespace string, podName string, clientset *kubernetes.Clientset, podColor color.Color256) {
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
		podColor.Println(podName + ":\n" + message)
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

/*func getRandomColor() string {
	rand.Seed(time.Now().UnixNano())
	colors := []string{
		"\033[31m","\033[32m","\033[33m","\033[34m","\033[35m","\033[36m","\033[37m","\033[97m",
	}
	return colors[rand.Intn(len(colors))]
}*/

func main() {
	rand.Seed(time.Now().UnixNano())
	pods := []string{
		"app", "app2", "app3", "app4",
	}
	clientset, err := getK8sClient()
	if err != nil {
		panic(err.Error())
	}
	//println(getRandomColor())
	//println(rand.Intn(100))
	for _, pod := range pods {
		podColor := color.C256(uint8(rand.Intn(256)))
		go getPodLogs("default", pod, clientset, podColor)
	}
	for {
	}
}
