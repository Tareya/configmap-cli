package common

import (
	"log"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// init the rest client
func InitClient() *kubernetes.Clientset {
	// create a kubeconfig instance
	c := LoadConfig()

	// init the rest client
	clientset, err := kubernetes.NewForConfig(c)
	if err != nil {
		log.Fatal(err)
	}

	return clientset
}

// load the kubeconfig file
func LoadConfig() *rest.Config {

	config, err := clientcmd.BuildConfigFromFlags("", "./admin.conf")
	if err != nil {
		log.Fatal(err)
	}

	return config
}
