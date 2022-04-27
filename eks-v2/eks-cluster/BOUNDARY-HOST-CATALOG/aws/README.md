# AWS deployment with CloudFormation or Terraform

This directory contains a template file for deploying a lab environment using the [CloudFormation](https://docs.aws.amazon.com/cloudformation/) and a `terraform/` directory for provisioning the lab environment with Terraform instead. Please note that the template and Terraform are separate deployment strategies.

The template is meant to be used with CloudFormation and can be launched using the [CloudFormation Web Console](https://console.aws.amazon.com/cloudformation) or [CloudFormation CLI](https://docs.aws.amazon.com/cli/latest/reference/cloudformation/index.html).

**Please note that usage of this template will incur costs associated with your AWS subscription.** You are responsible for these costs. See the Dynamic Host Catalogs tutorial section **Cleanup and teardown** to learn about destroying these resources after completing the tutorial.

The lab template creates:

- 4 Amazon Linux instances
  - AMI ID `ami-083602cee93914c0c`
  - Size: `t3.micro`

The VMs are named and tagged as follows:

- boundary-1-dev
    - Tags: `service-type`: `database` and `application`: `dev`
- boundary-2-dev
    - Tags: `service-type`: `database` and `application`: `dev`
- boundary-3-production
    - Tags: `service-type`: `database` and `application`: `production`
- boundary-4-production
    - Tags: `service-type`: `database` and `application`: `prod`

The fourth VM, `boundary-vm-4-production`, is purposefully misconfigured with a tag of `application`: `prod` that is corrected by the learner in the Dynamic Host Catalogs tutorial.
