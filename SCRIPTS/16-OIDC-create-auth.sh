ISSUER="https://ISSUER-DOMAIN-NAME/auth/realms/master"
CLIENT_ID="boundary"
CLIENT_SECRET="xxxx"
PREFIX_URL="http://<BOUNDARY-URL>:9200"



boundary auth-methods create oidc \
  -issuer "$ISSUER" \
  -client-id $CLIENT_ID \
  -client-secret $CLIENT_SECRET \
  -signing-algorithm RS256 \
  -api-url-prefix "$PREFIX_URL" \
  -name "keycloak"

