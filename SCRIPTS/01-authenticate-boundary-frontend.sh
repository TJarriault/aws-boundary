

AUTHMETHOD="ampw_XXXX"
BOUNDARY_ADDR="<BOUNDARY_URL>:9200"

boundary authenticate password -addr=$BOUNDARY_ADDR -auth-method-id=$AUTHMETHOD       -login-name=<USERNAME> -password=<PASSWORD> -keyring-type=none -format=json | jq -r ".item.attributes.token" > /tmp/boundary_token.txt
echo "BOUNDARY_TOKEN=$(cat /tmp/boundary_token.txt)"
export BOUNDARY_TOKEN=$(cat /tmp/boundary_token.txt)
echo "BOUNDARY_ADDR=$BOUNDARY_ADDR"
export BOUNDARY_ADDR=$BOUNDARY_ADDR

