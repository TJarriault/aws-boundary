USER=boundary4
aws iam put-user-policy \
  --user-name $USER \
  --policy-name BoundaryDescribeInstances \
  --policy-document file://./eks-v2/eks-cluster/BOUNDARY-HOST-CATALOG/aws/boundary-describe-instances-policy.json

aws iam get-user-policy --user-name $USER --policy-name BoundaryDescribeInstances
