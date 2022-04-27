# tja-featurepoc

======
# Prerequisite Helm Controller
For security reasons, we will privilege the use of dedicated controller by namespace.
We deployed one ingress controller per namespace
![Ref](https://kubernetes.github.io/ingress-nginx/user-guide/multiple-ingress/)

- [_] Init helm chart
```
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update
```

- [_] Deploy helm with restriction
specific **class name : classkeycloak** , this one will be used on the keycloak ingress specification

```
helm install ingress1 ingress-nginx/ingress-nginx --namespace keycloak --set controller.ingressClassResource.name=class1 --set controller.scope.namespace=keycloak --set controller.scope.enabled=true
```

- [_] Control controller deployment
```bash
kubectl get pods -n keycloak

NAME                                                             READY   STATUS    RESTARTS   AGE
keycloak-controller-ingress-nginx-controller-5d46cd95b5-d9lwn   1/1     Running   0          6h50m
keycloak-ingress-nginx-controller-6fc6459b95-dr9kt              1/1     Running   0          6h45m
```


```bash
kubectl get svc -n keycloak

NAME                                                      TYPE           CLUSTER-IP     EXTERNAL-IP      PORT(S)                      AGE
keycloak-controller-ingress-nginx-controller             LoadBalancer   10.12.14.102   34.77.91.68      80:30393/TCP,443:32625/TCP   6h51m
keycloak-controller-ingress-nginx-controller-admission   ClusterIP      10.12.13.141   <none>           443/TCP                      6h51m
```

```bash
kubectl get svc -n keycloak -o json | jq -r '.items[] |  select(.status.loadBalancer.ingress != null) | [.metadata.name,.status.loadBalancer.ingress[].ip]'

[
  "ingressdefault-ingress-nginx-controller",
  "34.140.97.235"
]
```

keycloak should be DNS on the address hold by the SERVICE



======
# Getting started to deploy Keycloak
To make this deployment, we'll use kubernetes helm charts

https://github.com/bitnami/charts/tree/master/bitnami/keycloak

- Init helm repository
```
helm repo add bitnami https://charts.bitnami.com/bitnami
helm install keycloak bitnami/keycloak -f values.yaml
```

- Specific chart configuration on values.yaml
```
ingress:
  ## @param ingress.enabled Enable ingress controller resource
  ##
  enabled: true
  ## DEPRECATED: Use ingress.annotations instead of ingress.certManager
  ## certManager: false
  ##

  ## @param ingress.hostname Default host for the ingress resource
  ##
  hostname: keycloak4.k8s.esecpocgcp.com
  ## @param ingress.apiVersion Force Ingress API version (automatically detected if not set)
  ##
  apiVersion: ""
  ## @param ingress.ingressClassName IngressClass that will be be used to implement the Ingress (Kubernetes 1.18+)
  ## This is supported in Kubernetes 1.18+ and required if you have more than one IngressClass marked as the default for your cluster
  ## ref: https://kubernetes.io/blog/2020/04/02/improvements-to-the-ingress-api-in-kubernetes-1.18/
  ##
  ingressClassName: "class1"
  ## @param ingress.path Ingress path
  ##
  path: /
  ## @param ingress.pathType Ingress path type
  ##
  pathType: ImplementationSpecific
  ## @param ingress.annotations Additional annotations for the Ingress resource. To enable certificate autogeneration, place here your cert-manager annotations.
  ## For a full list of possible ingress annotations, please see
  ## ref: https://github.com/kubernetes/ingress-nginx/blob/master/docs/user-guide/nginx-configuration/annotations.md
  ## Use this parameter to set the required annotations for cert-manager, see
  ## ref: https://cert-manager.io/docs/usage/ingress/#supported-annotations
  ##
  ## e.g:
  ## annotations:
  ##   kubernetes.io/ingress.class: nginx
  ##   cert-manager.io/cluster-issuer: cluster-issuer-name
  ##
  annotations:
    #kubernetes.io/ingress.class: "nginx"
    #nginx.org/redirect-to-https: "True"
    #ingress.kubernetes.io/ssl-redirect: "True"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.org/server-snippets: |
      location / {
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
      }
  ## @param ingress.tls Enable TLS configuration for the hostname defined at `ingress.hostname` parameter
  ## TLS certificates will be retrieved from a TLS secret with name: {{- printf "%s-tls" .Values.ingress.hostname }}
  ## You can use the ingress.secrets parameter to create this TLS secret, relay on cert-manager to create it, or
  ## let the chart create self-signed certificates for you
  ##
  tls: true
....
```


```
helm repo add bitnami https://charts.bitnami.com/bitnami
helm upgrade keycloak bitnami/keycloak -f values.yaml
```
