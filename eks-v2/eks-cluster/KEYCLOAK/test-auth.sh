#!/bin/bash

if [ $# -ne 5 ]; then
  echo 1>&2 "Usage: . $0 hostname realm username clientid"
  echo 1>&2 "  options:"
  echo 1>&2 "    hostname: keycloak.k8s.esecpocgcp.com"
  echo 1>&2 "    realm:poc-sogeti"
  echo 1>&2 "    clientid:sonarqube"
  echo 1>&2 "    For verify ssl: use 'y' (otherwise it will send curl post with --insecure)"
  
  return
fi

HOSTNAME="keycloak.k8s.esecpocgcp.com"
REALM_NAME="poc-sogeti"
USERNAME="toto"
CLIENT_ID="toto"
SECURE="n"



KEYCLOAK_URL=https://keycloak.k8s.esecpocgcp.com/auth/realms/poc-sogeti/protocol/openid-connect/token


TOKEN_URL="${KEYCLOAK_URL}/auth/realms/poc-sogeti/protocol/openid-connect/token"
#AUTH="Authorization: bearer $(curl -d client_id=admin-cli -d username=admin -d password=admin -d grant_type=password ${TOKEN_URL} | sed -n 's|.*"access_token":"\([^"]*\)".*|\1|p')"
AUTH="Authorization: bearer $(curl -d client_id=poc-sogeti -d username=toto -d password=toto -d grant_type=password ${TOKEN_URL} | jq -r '.access_token') -k"

CONVERTER_URL="${KEYCLOAK_URL}/auth/admin/realms/poc-sogeti/client-description-converter"
SAML_XML="@/tmp/saml_export/iam_saml.xml"
CLIENT_JSON=$(curl -X POST -H "${AUTH}"  -H 'content-type: application/json' ${CONVERTER_URL} --data-binary ${SAML_XML=} -k)

CLIENTS_URL="${KEYCLOAK_URL}/auth/admin/realms/poc-sogeti/clients"
curl -X POST -H "${AUTH}"  -H 'content-type: application/json' ${CLIENTS_URL} -d ${CLIENT_JSON} -k
