#!/usr/bin/env bash

set -ex

NAME=fabula

REPOROOT=$(git rev-parse --show-toplevel)
PROJECT=$(gcloud config get-value project)
SERVICE=$NAME
IMAGE=$NAME-image

#gcloud builds submit $REPOROOT \
#    --project $REPO_PROJECT_ID \
#    --config $REPOROOT/deploy/server/cloudbuild.yaml \
#    --substitutions=_IMAGE="$IMAGE"

gcloud container images list-tags \
    --format='get(digest)' \
    --filter=tags:latest gcr.io/fabula-resources/fabula-image \
    > $REPOROOT/deploy/server/fabula-image-latest.txt
