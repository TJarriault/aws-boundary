#!/bin/bash

AUTHMETHOD="ampw_xxx"
BOUNDARY_ADDR="<BOUNDARY URL>"
USER="username_to_be_define""
PWD="<PWD>"

boundary authenticate password -addr=$BOUNDARY_ADDR -auth-method-id=$AUTHMETHOD       -login-name=$USER -password=$PWD -keyring-type=none -format=json | jq -r ".item.attributes.token" > /tmp/boundary_token.txt
export BOUNDARY_TOKEN= "$(cat /tmp/boundary_token.txt)"
echo export BOUNDARY_TOKEN="$(cat /tmp/boundary_token.txt)"
export BOUNDARY_ADDR="${BOUNDARY_ADDR}"
echo export BOUNDARY_ADDR="${BOUNDARY_ADDR}"
