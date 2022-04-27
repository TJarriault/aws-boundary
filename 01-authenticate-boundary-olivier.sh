#!/bin/bash

AUTHMETHOD="ampw_S9bL7zKsKB"
BOUNDARY_ADDR="http://boundary-t-controller-glider-f4c794b6e55bbe02.elb.us-east-1.amazonaws.com:9200"
USER="olivier"
PWD="foofoofoo"

boundary authenticate password -addr=$BOUNDARY_ADDR -auth-method-id=$AUTHMETHOD       -login-name=$USER -password=$PWD -keyring-type=none -format=json | jq -r ".item.attributes.token" > /tmp/boundary_token.txt
export BOUNDARY_TOKEN= "$(cat /tmp/boundary_token.txt)"
echo export BOUNDARY_TOKEN="$(cat /tmp/boundary_token.txt)"
export BOUNDARY_ADDR="${BOUNDARY_ADDR}"
echo export BOUNDARY_ADDR="${BOUNDARY_ADDR}"
