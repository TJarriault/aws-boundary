USER="postgres"
PWD="boundarydemo"
#HOST="devoxxdemo.cusgj5wygqg1.us-east-1.rds.amazonaws.com"
HOST="aa4f61b5642794175838f036c43cb703-462774920.us-east-1.elb.amazonaws.com"
PG_DB="boundary"

export PG_DB=$PG_DB;export PG_URL="postgres://${USER}:${PWD}@devoxxdemo.cusgj5wygqg1.us-east-1.rds.amazonaws.com:5432/${PG_DB}?sslmode=disable"


psql -d $PG_URL -f northwind-database.sql --quiet
psql -h ad23350bd67594dd0b5b4517ff599f77-1675789865.us-east-1.elb.amazonaws.com -U postgres -W boundary -f northwind-roles.sql
