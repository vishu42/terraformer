#!/bin/bash

action=${1}
plan_file=${2}

# if number of arguments is less than 1, exit
if [ $# -lt 2 ]; then
    echo "Usage: $0 <action> <plan_output_file>"
    exit 1
fi

# create a directory to persist logs
LOG_DIR=/usr/local/var/log/terraformer
mkdir -p ${LOG_DIR}

# run terraform in ci mode
TF_IN_AUTOMATION=true

useAzureCredentials="false"
echo $ARM_CLIENT_ID
# if ARM_CLIENT_ID and ARM_CLIENT_SECRET and ARM_TENANT_ID and ARM_SUBSCRIPTION_ID are set, set useAzureCredentials to true
if [ -n "${ARM_CLIENT_ID}" ] && [ -n "${ARM_CLIENT_SECRET}" ] && [ -n "${ARM_TENANT_ID}" ] && [ -n "${ARM_SUBSCRIPTION_ID}" ]; then
    useAzureCredentials="true"
fi

# if useAzureCredentials is false, exit
if [ "${useAzureCredentials}" == "false" ]; then
    echo "ARM_CLIENT_ID and ARM_CLIENT_SECRET and ARM_TENANT_ID and ARM_SUBSCRIPTION_ID are not set"
    exit 1
fi

# # env var TERRAFORM_WORKDIR is required
# if [ -z "${TERRAFORM_WORKDIR}" ]; then
#     echo "TERRAFORM_WORKDIR is not set"
#     exit 1
# fi

# env var AUTO_APPROVE is optional
if [ -z "${AUTO_APPROVE}" ]; then
    AUTO_APPROVE="false"
fi

# if AUTO_APPROVE is true, run terraform in auto-approve mode and exit 0
if [ "${AUTO_APPROVE}" == "true" ]; then
    terraform init -input=false
    terraform apply -input=false -auto-approve
    exit 0
fi

# run terraform in ci mode
terraform init -input=false

if [ "${action}" == "plan" ]; then
    terraform plan -input=false -out=${LOG_DIR}/${plan_file}
elif [ "${action}" == "apply" ]; then
    terraform apply -input=false ${LOG_DIR}/${plan_file}
elif [ "${action}" == "destroy" ]; then
    terraform destroy -input=false
else
    echo "Invalid action: ${action}"
    exit 1
fi

