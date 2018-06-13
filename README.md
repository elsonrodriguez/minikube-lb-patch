# Setting up Minikube to have routable Cluster IPs and External IPs

This guide will show you how to access services within your minikube instance.

## Accessing Cluster IPs

First, we will need to add a route to minikube.

This oneliner gets the ClusterIP range from minikube's config and adds a route.

```
sudo route -n add -net $(cat ~/.minikube/profiles/minikube/config.json | jq -r ".KubernetesConfig.ServiceCIDR") $(minikube ip)
```

To test, deploy nginx.

```
kubectl run nginx --image=nginx --replicas=1
kubectl expose deployment nginx --port=80 --target-port=80 --type=LoadBalancer
```

And connect to it:

```
nginx_ip=$(kubectl get svc nginx -o jsonpath='{.spec.clusterIP}')
curl $nginx_ip
```

# Assigning External IPs

Minikube does not provision an external IP when `type=LoadBalancer`. This prevents some apps that talk to the k8s API from working properly. 

A custom controller must be run on your Minikube cluster to enable this functionality.

DO NOT RUN THIS ON A NON-MINIKUBE CLUSTER!!

```
kubectl run minikube-lb-patch --replicas=1 --image=elsonrodriguez/minikube-lb-patch:0.1 --namespace=kube-system
```

This controller will assign the External IP to the ClusterIP, which has been made routable in the previous section.

```
kubectl  get svc
NAME         CLUSTER-IP   EXTERNAL-IP   PORT(S)        AGE
kubernetes   10.0.0.1     <none>        443/TCP        18d
nginx        10.0.0.98    10.0.0.98     80:31834/TCP   40s
```

Nothing should change connectivity-wise, however your app can now just scrape the usual field for external ip:

```
nginx_external_ip=$(kubectl get svc nginx -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
curl $nginx_external_ip
```

## Teardown

To undo all this:
```
kubectl  delete deployment nginx
kubectl  delete svc nginx
kubectl  delete deployment minikube-lb-patch -nkube-system
sudo route -n delete -net $(cat ~/.minikube/profiles/minikube/config.json | jq -r ".KubernetesConfig.ServiceCIDR") $(minikube ip)
```
