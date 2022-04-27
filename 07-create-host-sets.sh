. env.sh

catalogid="hc_xxx"

boundary host-sets create plugin -name database -host-catalog-id $catalogid -attr filters=tag:service-type=database > /tmp/hosts-sets


cat /tmp/hosts-sets

DATABASE=`cat /tmp/hosts-sets |head -10 | grep ID |grep -v "Host" | awk '{print $2}'`

export DATABASE_HOST_SET_ID=$DATABASE


# DEV
boundary host-sets create plugin \
  -name dev \
  -host-catalog-id $catalogid \
  -attr filters=tag:application=dev > /tmp/plugin-appdev

cat /tmp/plugin-appdev

DEV_HOST_SET=`cat /tmp/plugin-appdev |head -10 | grep ID |grep -v "Host" | awk '{print $2}'`
export DEV_HOST_SET_ID=$DEV_HOST_SET

#PROD
boundary host-sets create plugin \
  -name prod \
  -host-catalog-id $catalogid \
  -attr filters=tag:application=production > /tmp/plugin-appprod

cat /tmp/plugin-appprod

PROD_HOST_SET=`cat /tmp/plugin-appprod |head -10 | grep ID |grep -v "Host" | awk '{print $2}'`
export PROD_HOST_SET_ID=$PROD_HOST_SET

