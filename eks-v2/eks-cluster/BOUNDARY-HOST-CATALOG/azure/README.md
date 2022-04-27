# Azure Resource Manager deployment

This directory contains a template file for deploying a lab environment using the [Azure Resource Manager (ARM)](https://docs.microsoft.com/en-us/azure/azure-resource-manager/management/overview) and a `terraform/` directory for provisioning the lab environment with Terraform instead. Please note that the template and Terraform are separate deployment strategies.

The template is meant to be used with a [Template deployment](https://docs.microsoft.com/en-us/azure/azure-resource-manager/templates/syntax) that can be launched using the Azure Portal or Azure CLI.

**Please note that usage of this template will incur costs associated with your Azure subscription.** You are responsible for these costs. See the Dynamic Host Catalogs tutorial section **Cleanup and teardown** to learn about destroying these resources after completing the tutorial.

The lab template creates:

- 4 Centos VMs
  - Publisher: OpenLogin
  - SKU: 7_9-gen2
  - Size: Standard_B1ls

The VMs are named and tagged as follows:

- boundary-vm-1-dev
    - Tags: `service-type`: `database` and `application`: `dev`
- boundary-vm-2-dev
    - Tags: `service-type`: `database` and `application`: `dev`
- boundary-vm-3-production
    - Tags: `service-type`: `database` and `application`: `production`
- boundary-vm-4-production
    - Tags: `service-type`: `database` and `application`: `prod`

The fourth VM, `boundary-vm-4-production`, is purposefully misconfigured with a tag of `application`: `prod` that is corrected by the learner in the Dynamic Host Catalogs tutorial.
