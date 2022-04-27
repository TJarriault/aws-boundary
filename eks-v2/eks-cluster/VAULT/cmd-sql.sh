USER="postgres"

PWD=$(kubectl get secret --namespace postgresql postgresql-boundary-devoxx -o jsonpath="{.data.postgres-password}" | base64 --decode)
export POSTGRES_PASSWORD=$PWD

export SERVICE_IP=$(kubectl get svc --namespace postgresql postgresql-boundary-devoxx --template "{{ range (index .status.loadBalancer.ingress 0) }}{{ . }}{{ end }}")
SERVICE_IP=$(kubectl get svc --namespace postgresql postgresql-boundary-devoxx --template "{{ range (index .status.loadBalancer.ingress 0) }}{{ . }}{{ end }}")

PG_DB="boundary"
PG_DB="northwind"

export PG_DB=$PG_DB;export PG_URL="postgres://${USER}:${PWD}@${HOST}:5432/${PG_DB}?sslmode=disable"


psql -d $PG_URL -f sql/northwind-database.sql --quiet
psql -d $PG_URL -f sql/northwind-roles.sql
