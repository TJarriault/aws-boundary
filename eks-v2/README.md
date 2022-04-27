aws sts get-caller-identity
aws eks update-kubeconfig --region us-east-1 --name boundary-devoxx

# Init EKS
1) Deploy EKS cluster
```
cd eks-cluster
terraform apply
```

2) Deploy HELM
add repo on Helm
```
helm repo add bitnami https://charts.bitnami.com/bitnami
helm search repo bitnami

helm repo update
```

# Deploy NGINX PROXY
```
CD NGINX
helm install ingresstest ingress-nginx/ingress-nginx --namespace ingress-nginx --set controller.ingressClassResource.name=classdtest --set controller.scope.namespace=test
```


# Deploy POSGRESQL
```
cd POSTGRESQL

helm upgrade --install postgresql-boundary-devoxx bitnami/postgresql -n postgresql --create-namespace -f values.yaml
```


```
helm install postgresql-boundary-devoxx bitnami/postgresql -f values.yaml
NAME: postgresql-boundary-devoxx


    postgresql-boundary-devoxx.postgresql.svc.cluster.local - Read/Write connection

    export POSTGRES_PASSWORD=$(kubectl get secret --namespace postgresql postgresql-boundary-devoxx -o jsonpath="{.data.postgres-password}" | base64 --decode)

To connect to your database run the following command:

    kubectl run postgresql-boundary-devoxx-client --rm --tty -i --restart='Never' --namespace postgresql --image docker.io/bitnami/postgresql:14.2.0-debian-10-r45 --env="PGPASSWORD=$POSTGRES_PASSWORD" \
      --command -- psql --host postgresql-boundary-devoxx -U postgres -d boundary -p 5432

    > NOTE: If you access the container using bash, make sure that you execute "/opt/bitnami/scripts/entrypoint.sh /bin/bash" in order to avoid the error "psql: local user with ID 1001} does not exist"

To connect to your database from outside the cluster execute the following commands:

  NOTE: It may take a few minutes for the LoadBalancer IP to be available.
        Watch the status with: 'kubectl get svc --namespace postgresql -w postgresql-boundary-devoxx'

    export SERVICE_IP=$(kubectl get svc --namespace postgresql postgresql-boundary-devoxx --template "{{ range (index .status.loadBalancer.ingress 0) }}{{ . }}{{ end }}")
    PGPASSWORD="$POSTGRES_PASSWORD" psql --host $SERVICE_IP --port 5432 -U postgres -d boundary

```



# Deploy VAULT
Ref : https://www.vaultproject.io/docs/platform/k8s/helm

```
cd VAULT

helm repo add hashicorp https://helm.releases.hashicorp.com

helm upgrade --install -n vault vault hashicorp/vault  -f values.yaml --create-namespace
```

## inject data/role on postgresql for test to populate database
./cmd-sql.sh

##AUTH
```
export VAULT_ADDR=https://vault.devoxx.esecpocgcp.com
vault login
export VAULT_TOKEN=xxxxx
export POSTGRES_PASSWORD=$(kubectl get secret --namespace postgresql postgresql-boundary-devoxx -o jsonpath="{.data.postgres-password}" | base64 --decode)
export SERVICE_IP=$(kubectl get svc --namespace postgresql postgresql-boundary-devoxx --template "{{ range (index .status.loadBalancer.ingress 0) }}{{ . }}{{ end }}")

vault policy write admin vault-policy/admin-policy.hcl
vault write auth/userpass/users/TJarriault  policies=admin
```

[*]- Create the boundary-controller policy
```
vault policy write boundary-controller vault-policy/boundary-controller-policy.hcl
```


[*]- Configure Vault with the postgres-database-plugin, connection information and allowed roles of dba and analyst:
```
vault write database/config/northwind \
      plugin_name=postgresql-database-plugin \
      connection_url="postgresql://postgres:{{POSTGRES_ADMIN_PASSWORD}}@{{SERVICE_IP}}:5432/postgres?sslmode=disable" \
      allowed_roles=dba,analyst \
      username="vault" \
      password="vault-password"
```


[*]- Create DBA role
```
vault write database/roles/dba \
      db_name=northwind \
      creation_statements=@vault-policy/dba.sql.hcl \
      default_ttl=3m \
      max_ttl=60m
```

[*]- Generate & Read postgresql credentials
```
vault read database/creds/dba
```

Response
```
Key                Value
---                -----
lease_id           database/creds/dba/x95VXP4zXCpUlOBcDjwdoPvj
lease_duration     3m
lease_renewable    true
password           -2KSNaXcT5pfDdUVNZMi
username           v-root-dba-q5Xpp1HDFu1XWNIXqZmk-1648811218
```


[*]- Create analyst role
```
vault write database/roles/analyst \
      db_name=northwind \
      creation_statements=@vault-policy/analyst.sql.hcl \
      default_ttl=3m \
      max_ttl=60m
```


[*]- Generate & Read postgresql credentials
```
vault read database/creds/analyst
```


Response
```
Key                Value
---                -----
lease_id           database/creds/analyst/SbavgaipDXWc3lnz9kPup0NZ
lease_duration     3m
lease_renewable    true
password           1jFj6Y-w6wLMkOrjgDOd
username           v-root-analyst-kgI37hYkHs9RAcw0t7t3-1648811427
```

[*]- Create northwind-database policy
    read database/creds/analyst
    read database/creds/dba
```
vault policy write northwind-database northwind-database-policy.hcl
```


[*]- Create Vault token for Boundary

A Vault token is needed to access the Boundary credential store that will be configured when setting up Boundary.

It's very important that the token is:
    periodic
    orphan
    renewable

```
vault token create \
  -no-default-policy=true \
  -policy="boundary-controller" \
  -policy="northwind-database" \
  -orphan=true \
  -period=20m \
  -renewable=true
```




# Configure BOUNDARY/VAULT : Credential Brokering
Ref : https://learn.hashicorp.com/tutorials/boundary/vault-cred-brokering-quickstart?in=boundary/configuration
![brokering](https://mktg-content-api-hashicorp.vercel.app/api/assets?product=tutorials&version=main&asset=public%2Fimg%2Fboundary%2Fboundary-vault-quickstart-overview.png)

git clone https://github.com/hashicorp/learn-boundary-vault-quickstart
cd learn-boundary-vault-quickstart






# Deploy keycloak

```
cd KEYCLOAK

helm install keycloak bitnami/keycloak -f values.yaml
```
NAME: keycloak

Keycloak can be accessed through the following DNS name from within your cluster:
    keycloak.default.svc.cluster.local (port 80)


1. Get the Keycloak URL and associate its hostname to your cluster external IP:
   export CLUSTER_IP=$(minikube ip) # On Minikube. Use: `kubectl cluster-info` on others K8s clusters
   echo "Keycloak URL: https://keycloak.devoxx.esecpocgcp.com/auth"
   echo "$CLUSTER_IP  keycloak.devoxx.esecpocgcp.com" | sudo tee -a /etc/hosts

2. Access Keycloak using the obtained URL.
3. Access the Administration Console using the following credentials:

  echo Username: user
  echo Password: $(kubectl get secret --namespace default keycloak -o jsonpath="{.data.admin-password}" | base64 --decode)

```
kubectl get secret keycloak -n default -o json | jq -r '.data."admin-password"' |base64 -d
```
