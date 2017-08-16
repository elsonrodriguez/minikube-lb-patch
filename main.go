package main

import (
	"fmt"
	"os"
	"os/user"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/rest"

	"log"
)

func main() {
	kubernetes_host := os.Getenv("KUBERNETES_SERVICE_HOST")
	kubeconfig := os.Getenv("KUBECONFIG")
	in_cluster := os.Getenv("MINIKUBELB_IN_CLUSTER")

	usr, err := user.Current()
	if err != nil {
		log.Fatalf("Error getting current user: %s", err)
	}

	if kubeconfig == "" {
		kubeconfig = usr.HomeDir + "/.kube/config"
	}

	if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
		fmt.Printf("kubeconfig not found at %v\n", kubeconfig)
		kubeconfig = ""
	}

	var config *rest.Config

	if kubernetes_host != "" || in_cluster == "True" {
		// creates the in-cluster config
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Fatalf("Error configuring in-cluster client: %s", err)
		}
	} else if kubeconfig != "" {
		// use the current context in kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			log.Fatalf("Error configuring kubeconfig client: %s", err)
		}
	} else {
		config = &rest.Config{
			Host: "http://127.0.0.1:8001",
		}
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error preparing connection to kubernetes cluster: %s", err)
	}

	//need sanity check to see if there's existing external-lb functionality, or maybe at least an overridable option to exit out if not running on minikube.

	for {
		svc, err := clientset.CoreV1().Services("").List(metav1.ListOptions{})
		if err != nil {
			log.Fatalf("Error getting services from kubernetes cluster: %s", err)
		}
		for _, svc := range svc.Items {
			if svc.Spec.Type == "LoadBalancer" && len(svc.Status.LoadBalancer.Ingress) == 0 {
				fmt.Printf("Service %q has no ingress for its loadbalancer\n", svc.Name)
				patch := []byte(fmt.Sprintf(`[{"op": "add", "path": "/status/loadBalancer/ingress", "value":  [ { "ip": "%s" } ] }]`, svc.Spec.ClusterIP))
				err := clientset.CoreV1().RESTClient().Patch(types.JSONPatchType).Resource("services").Namespace(svc.Namespace).Name(svc.Name).SubResource("status").Body(patch).Do().Error()
				if err != nil {
					log.Fatalf("Error patching service %s: %s", svc.Name, err)
				}
			} else if svc.Spec.Type == "LoadBalancer" {
				fmt.Printf("Service %q has IP %v for its loadbalancer\n", svc.Name, svc.Status.LoadBalancer.Ingress[0].IP) //need to check for hostnames and other types too.
			} else {
				fmt.Printf("Service %q is not type LoadBalancer\n", svc.Name)
			}
		}
		time.Sleep(1 * time.Second)
	}
}

	func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}