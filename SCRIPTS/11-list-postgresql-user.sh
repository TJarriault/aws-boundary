SERVICE_IP=$(kubectl get svc --namespace postgresql postgresql-boundary-devoxx --template "{{ range (index .status.loadBalancer.ingress 0) }}{{ . }}{{ end }}")
psql -h $SERVICE_IP -U postgres -c "\du"
