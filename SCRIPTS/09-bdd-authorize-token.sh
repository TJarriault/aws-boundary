

ANALYST_TARGET_ID=ttcp_xxx

boundary targets authorize-session -id $ANALYST_TARGET_ID
boundary targets authorize-session -id $ANALYST_TARGET_ID -format json | jq .
