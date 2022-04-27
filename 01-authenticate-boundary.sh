#boundary authenticate password \
#    -addr=http://boundary-t-controller-werewolf-2945817b2d73f929.elb.us-east-1.amazonaws.com:9200 \
#    -auth-method-id=ampw_rsn57FCOqL \
#    -login-name=jim \
#    -password=foofoofoo


AUTHMETHOD="ampw_XXXX"
BOUNDARY_ADDR="<BOUNDARY_URL>:9200"

boundary authenticate password -addr=$BOUNDARY_ADDR -auth-method-id=$AUTHMETHOD       -login-name=jim -password=<PASSWORD> -keyring-type=none -format=json | jq -r ".item.attributes.token" > /tmp/boundary_token.txt
echo "BOUNDARY_TOKEN=$(cat /tmp/boundary_token.txt)"
export BOUNDARY_TOKEN=$(cat /tmp/boundary_token.txt)
echo "BOUNDARY_ADDR=$BOUNDARY_ADDR"
export BOUNDARY_ADDR=$BOUNDARY_ADDR

