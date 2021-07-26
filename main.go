//no follow
/*podLogOpts := corev1.PodLogOptions{}
req := clientset.CoreV1().Pods(myPod.Namespace).GetLogs(myPod.Name, &podLogOpts)
podLogs, err := req.Stream(context.TODO())
if err != nil {
	println("error in opening stream")
	panic(err.Error())
}
defer podLogs.Close()

buf := new(bytes.Buffer)
_, err = io.Copy(buf, podLogs)
if err != nil {
	println("error in copy information from podLogs to buf")
	panic(err.Error())
}
str := buf.String()

println(str)
*/

package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	//
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
)

func main() {
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

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	namespace := "default"
	pod := "loki-0"
	myPod, err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), pod, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		fmt.Printf("Pod %s in namespace %s not found\n", pod, namespace)
	} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
		fmt.Printf("Error getting pod %s in namespace %s: %v\n",
			pod, namespace, statusError.ErrStatus.Message)
	} else if err != nil {
		panic(err.Error())
	} else {
		fmt.Printf("Found pod %s in namespace %s\n", pod, namespace)
	}

	count := int64(100)
	podLogOpts := corev1.PodLogOptions{Follow: true, TailLines: &count}
	req := clientset.CoreV1().Pods(myPod.Namespace).GetLogs(myPod.Name, &podLogOpts)
	podLogs, err := req.Stream(context.TODO())
	if err != nil {
		println("error in opening stream")
		panic(err.Error())
	}
	defer func() {
		err = podLogs.Close()
		if err != nil {
			println("error in closing stream")
			panic(err.Error())
		}
	}()

	for {
		buf := make([]byte, 2000)
		numBytes, err := podLogs.Read(buf)
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
		fmt.Print(message)
	}
}