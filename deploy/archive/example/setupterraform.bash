#!/usr/bin/env bash

set -ex

gcloud projects create ${TF_ADMIN} \
#  --organization ${TF_VAR_org_id} \
  --set-as-default

gcloud beta billing projects link ${TF_ADMIN} \
  --billing-account ${TF_VAR_billing_account}

gcloud iam service-accounts create terraform \
  --display-name "Terraform admin account"

gcloud iam service-accounts keys create ${TF_CREDS} \
  --iam-account terraform@${TF_ADMIN}.iam.gserviceaccount.com

gcloud projects add-iam-policy-binding ${TF_ADMIN} \
  --member serviceAccount:terraform@${TF_ADMIN}.iam.gserviceaccount.com \
  --role roles/viewer

gcloud projects add-iam-policy-binding ${TF_ADMIN} \
  --member serviceAccount:terraform@${TF_ADMIN}.iam.gserviceaccount.com \
  --role roles/storage.admin

gcloud services enable cloudresourcemanager.googleapis.com
gcloud services enable cloudbilling.googleapis.com
gcloud services enable iam.googleapis.com
gcloud services enable compute.googleapis.com
gcloud services enable serviceusage.googleapis.com

# remote state
# gsutil mb -p ${TF_ADMIN} gs://${TF_ADMIN}
# gsutil versioning set on gs://${TF_ADMIN}

# vars for later
# export GOOGLE_APPLICATION_CREDENTIALS=${TF_CREDS}
# export GOOGLE_PROJECT=${TF_ADMIN}
# export TF_VAR_admin_project_name=${TF_ADMIN}
