package main

import (
    "context"
    "fmt"
    "log"
    "io/ioutil"
    "os"
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

    var client k8s.Client
    var err error

    if kubernetes_host == "" {
        client, err = k8s.NewInClusterClient()
        if err != nil {
            log.Fatal(err)
        }
    } else if kubeconfig != "" {
        client, err = loadClient(kubeconfig)
        if err != nil {
            log.Fatal(err)
        }
    } else if kubeconfig == "" {
	client = k8s.Client{
                    Endpoint:  "http://127.0.0.1:8080",
                    Namespace: "default",
                    Client: &http.Client{} }
    }
//    client := k8s.Client{}

    nodes, err := client.CoreV1().ListNodes(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    for _, node := range nodes.Items {
        fmt.Printf("name=%q schedulable=%t\n", *node.Metadata.Name, !*node.Spec.Unschedulable)
    }
}
