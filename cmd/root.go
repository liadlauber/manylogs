// Package cmd /*
package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"gopkg.in/gookit/color.v1"
	"io"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	typev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"math/rand"
	"os"
	"strings"
	"time"
)

var namespace, label, container, kubeconfig string
var rootCmd = &cobra.Command{
	Use:   "manylogs [flags]",
	Short: "Print the logs for containers in pods matching label from a given namespace",
	Long: `Print the logs for containers in pods matching label from a given namespace. If the pods has only one container, the container name is
optional.

Example:
  # Return logs from manager container in pods matching label control-plane=controller-manager at namespace sns-system.
  manylogs -namespace="sns-system" -label="control-plane=controller-manager" -container="manager"`,

	Run: func(cmd *cobra.Command, args []string) {
		ch := make(chan string)

		rand.Seed(time.Now().UnixNano())

		clientset, err := getK8sClient()
		if err != nil {
			panic(err.Error())
		}

		if namespace == "" {
			fmt.Println("you must enter namespace flag")
			os.Exit(1)
		}

		if label == "" {
			fmt.Println("you must enter label flag")
			os.Exit(1)
		}

		pods, err := getLabeledPods(label, namespace, clientset.CoreV1())
		if err != nil {
			panic(err.Error())
		}

		for _, pod := range pods.Items {
			podColor := color.C256(uint8(rand.Intn(256)))
			go getPodLogs(namespace, pod.Name, container, clientset, podColor, ch)
		}
		println(<-ch)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&namespace, "namespace", "", "a namespace to get logs from")
	rootCmd.PersistentFlags().StringVar(&label, "label", "", "a label matches the pods to get logs from")
	rootCmd.PersistentFlags().StringVar(&container, "container", "", "(optional) a container to get logs from")
	rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "/root/.kube/config", "(optional) absolute path to the kubeconfig file)")
}

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
			close(ch)
			break
		}
		if err != nil {
			close(ch)
			panic(err.Error())
		}
		message := string(buf[:numBytes])
		podColor.Printf(podName + ":\t" + message)
	}
	close(ch)
	panic(err.Error())
}

func getK8sClient() (*kubernetes.Clientset, error) {

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
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
