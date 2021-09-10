package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// namespace := os.Getenv("POD_NAMESPACE")
	// name := os.Getenv("POD_NAME")

	// filepath := filepath.Join("/data/apps", "")

	// configmap := ConfigmapGenerate(namespace, name, filepath)
	// ConfigmapCreate(configmap)
	p := PodGet()
	project, version, err := GetPodInfo(p)
	if err != nil {
		log.Fatal(err)
	}

	namespace := os.Getenv("POD_NAMESPACE")
	filename := strings.Join([]string{project, version}, "-")           // joint the configmap name
	filepath := filepath.Join("/data/apps", project, "tmp/config.json") // joint the tmp file path

	// fmt.Println(namespace, filename, filepath)

	c := ConfigmapGenerate(namespace, filename, filepath)

	fmt.Println(c)
}

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

	config, err := clientcmd.BuildConfigFromFlags("", "./cluster-admin.conf")
	if err != nil {
		log.Fatal(err)
	}

	return config
}

// create the configmap
func ConfigmapCreate(c *corev1.ConfigMap) *corev1.ConfigMap {
	// init the clientset instance
	client := InitClient()

	// define the api iinterface
	api := client.CoreV1()
	opts := metav1.CreateOptions{}

	namespace := os.Getenv("POD_NAMESPACE")

	configmapCreate, err := api.ConfigMaps(namespace).Create(context.TODO(), c, opts)
	if err != nil {
		log.Fatal(err)
	}

	return configmapCreate
}

// generate the configmap struct body
func ConfigmapGenerate(namespace string, name string, filepath string) *corev1.ConfigMap {

	// define the configmap necessary metadata
	type ConfigMap struct {
		Metadata interface{}
		Data     map[string]string
	} // *corev1.ConfigMap

	type ObjectMeta struct {
		Name      string
		Namespace string
	} // *metav1.ObjectMeta

	// read the tmp json config file
	f, e := ioutil.ReadFile(filepath)
	if e != nil {
		log.Fatal(e)
	}

	jsonStr := string(f)

	// convert the json string to map
	m := make(map[string]string)
	m["config.json"] = jsonStr

	metadata := ObjectMeta{
		Name:      name,
		Namespace: namespace,
	}

	// generate the k8s configmap data
	result := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      metadata.Name,
			Namespace: metadata.Namespace,
		},
		Data: m,
	}

	return result
}

func PodGet() *corev1.Pod {

	// init the clientset instance
	client := InitClient()

	// define the api iinterface
	api := client.CoreV1()
	opts := metav1.GetOptions{}

	namespace := os.Getenv("POD_NAMESPACE")
	name := os.Getenv("POD_NAME")

	podGet, err := api.Pods(namespace).Get(context.TODO(), name, opts)
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Println(podGet)

	// GetPodInfo(podGet)

	return podGet

}

func GetPodInfo(pod *corev1.Pod) (project, version string, err error) {

	var image string

	for _, v := range pod.Spec.Containers {
		project = v.Name
		image = v.Image
	}

	version = strings.Split(image, ":")[1]

	// fmt.Println(project, image, version)

	return project, version, err
}
