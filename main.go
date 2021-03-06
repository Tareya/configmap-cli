package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	// get pod information like project name and image version
	p := PodGet()
	project, version, err := GetPodInfo(p)
	if err != nil {
		log.Fatal(err)
	}

	re := strings.ToLower(version)

	// get necessary fieldss
	namespace := os.Getenv("POD_NAMESPACE")
	filename := strings.Join([]string{project, re}, "-")                // joint the configmap name
	filepath := filepath.Join("/data/apps", project, "tmp/config.json") // joint the tmp file path

	// fmt.Println(namespace, filename, filepath)

	// check the tmp file exists
	_, e := os.Stat(filepath)
	if os.IsNotExist(e) {
		fmt.Println("start to sleep")
		time.Sleep(10 * time.Second) //
		fmt.Println("sleep 8 sec")
	}

	// judge the configmap exsit status, and create the configmap when not exsits.
	cl := ConfigmapList(namespace)

	status := ConfigmapStatus(cl, filename)

	if status == true {
		fmt.Printf("Configmap %v/%v already exists.", namespace, filename)
	} else {
		c := ConfigmapGenerate(namespace, filename, filepath)
		e := ConfigmapCreate(namespace, c)
		if e == nil {
			fmt.Println("Create configmap failure.")
			return
		} else {
			fmt.Printf("Configmap %v/%v created successfully", namespace, filename)
		}
	}

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

	kubeconfig := JudgeConfig()

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	return config
}

// select the kubeconfig which env
func JudgeConfig() string {

	namespace := os.Getenv("POD_NAMESPACE")

	env := strings.Split(namespace, "-app")[0]
	kubeconfig := filepath.Join("./config", strings.Join([]string{env, ".conf"}, ""))

	return kubeconfig
}

// list the configmaps
func ConfigmapList(namespace string) *corev1.ConfigMapList {
	// init the clientset instance
	client := InitClient()

	// define the api interface
	api := client.CoreV1()
	opts := metav1.ListOptions{}

	// list configmaps operation
	configmapList, err := api.ConfigMaps(namespace).List(context.TODO(), opts)
	if err != nil {
		log.Fatal(err)
	}

	return configmapList
}

// judge the configmap exsit status
func ConfigmapStatus(cl *corev1.ConfigMapList, name string) bool {

	// init the regexp instance
	re := regexp.MustCompile(name)

	num := 0
	// traverse the configmap list to find the input configmap and return the result
	for _, configmap := range cl.Items {
		result := re.MatchString(configmap.Name)
		if result == false {
			num += 0
		} else {
			num += 1
		}
	}
	if num > 0 {
		return true
	} else {
		return false
	}
}

// create the configmap
func ConfigmapCreate(namespace string, c *corev1.ConfigMap) *corev1.ConfigMap {
	// init the clientset instance
	client := InitClient()

	// define the api interface
	api := client.CoreV1()
	opts := metav1.CreateOptions{}

	// create configmap operation
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
