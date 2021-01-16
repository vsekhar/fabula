#!/usr/bin/bash

set -ex

# TODO: get name of group, pick an instance, ssh into it
GROUP=$(terraform state get)
INSTANCE=$(gcloud compute instance-groups list-instances)
