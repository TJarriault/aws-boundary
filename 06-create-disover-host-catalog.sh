. env.sh
boundary host-catalogs create plugin \
  -scope-id p_xxx \
  -plugin-name aws \
  -attr disable_credential_rotation=true \
  -attr region=us-east-1 \
  -secret access_key_id=env://BOUNDARY_ACCESS_KEY_ID \
  -secret secret_access_key=env://BOUNDARY_SECRET_ACCESS_KEY

