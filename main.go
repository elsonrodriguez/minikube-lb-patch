package main

import (
    "context"
    "fmt"
    "log"
    "io/ioutil"
    "os"
    "os/user"
    "net/http"

    "github.com/ghodss/yaml"
    "github.com/ericchiang/k8s"
)

// loadClient parses a kubeconfig from a file and returns a Kubernetes
// client. It does not support extensions or client auth providers.
func loadClient(kubeconfigPath string) (*k8s.Client, error) {
	data, err := ioutil.ReadFile(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("read kubeconfig: %v", err)
	}

	// Unmarshal YAML into a Kubernetes config object.
	var config k8s.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("unmarshal kubeconfig: %v", err)
	}
	return k8s.NewClient(&config)
}

func main() {
    kubernetes_host := os.Getenv("KUBERNETES_SERVICE_HOST")
    kubeconfig := os.Getenv("KUBECONFIG")
    in_cluster := os.Getenv("MINIKUBELB_IN_CLUSTER")

    usr, err := user.Current()
    if err != nil {
        log.Fatal( err )
    }


    if kubeconfig == "" {
        kubeconfig = usr.HomeDir + "/.kube/config"
    }

    if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
        fmt.Printf("kubeconfig not found at %v\n", kubeconfig)
        kubeconfig = "" 
    }


    var client *k8s.Client

    if kubernetes_host != "" || in_cluster == "True" {
        client, err = k8s.NewInClusterClient()
        if err != nil {
            log.Fatal(err)
        }

    } else if kubeconfig != "" {
        client, err = loadClient(kubeconfig)
        if err != nil {
            log.Fatal(err)
        }
    } else {
	client = &k8s.Client{
                    Endpoint:  "http://127.0.0.1:8001",
                    Namespace: "default",
                    Client: &http.Client{} }
    }

    svc, err := client.CoreV1().ListServices(context.Background(), "")
    if err != nil {
        log.Fatal(err)
    }
    for _, svc := range svc.Items {
        //fmt.Println(reflect.TypeOf(svc.Status.LoadBalancer.Ingress))
        if *svc.Spec.Type == "LoadBalancer" && len(svc.Status.LoadBalancer.Ingress) == 0 {
          fmt.Printf("Service %q has no ingress for its loadbalancer\n", *svc.Metadata.Name)
        } else if *svc.Spec.Type == "LoadBalancer" {
           fmt.Printf("Service %q has IP %v for its loadbalancer\n", *svc.Metadata.Name, *svc.Status.LoadBalancer.Ingress[0].Ip) //need to check for hostnames and other types too.
        } else {
           fmt.Printf("Service %q is not type LoadBalancer\n", *svc.Metadata.Name)
        }
    }
}

