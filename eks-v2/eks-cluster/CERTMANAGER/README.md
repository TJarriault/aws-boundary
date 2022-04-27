


# Cert Manager
cert-manager provides Helm charts as a first-class method of installation on Kubernetes

## Helm Cert-Manager
- [ ] [Ref] (https://cert-manager.io/docs/installation/helm/)
```bash
helm repo add jetstack https://charts.jetstack.io
helm repo update
```

## Install Cert-Manager
- Install the cert-manager Helm chart
```bash
helm install \
  cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --version v1.6.1 \
  --set installCRDs=true
```


- Validate cert-manager deployment

```bash
kubectl get pods --namespace cert-manager
```

## Configure CLUSTER cert-manager to use ACME as an issuer
- [ ] [Ref] (https://cert-manager.io/docs/configuration/) 
- [ ] [Ref] (https://cert-manager.io/docs/configuration/acme/http01/) with HTTP01 configuration

Create the deployment file : letencrypt-prod-cluster.yaml
```bash
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    # You must replace this email address with your own.
    # Let's Encrypt will use this to contact you about expiring
    # certificates, and issues related to your account.
    email: tony.jarriault@capgemini.com
    server: https://acme-v02.api.letsencrypt.org/directory
    privateKeySecretRef:
      # Secret resource that will be used to store the account's private key.
      name: prod-issuer-letsencrypt
    # Add a single challenge solver, HTTP01 using nginx
    solvers:
    - http01:
        ingress:
          class: nginx
```

privateKeySecretRef.name value will be reuse on each application to generate SSL certificate


Create issuer
```bash
 kubectl apply -f letencrypt-prod-cluster.yaml
```

Control deployment
```bash
kubectl get ClusterIssuer -A

NAME               READY   AGE
letsencrypt-prod   True    96m

```


