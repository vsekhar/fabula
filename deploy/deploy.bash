#!/usr/bin/env bash

set -ex

NAME=fabula

PROJECT=$(gcloud config get-value project)
SERVICE=$NAME
IMAGE=$NAME-image


# server in us-central1
(cd server && exec go mod tidy)
gcloud builds submit server --tag gcr.io/$PROJECT/$IMAGE
gcloud run deploy $SERVICE \
    --image gcr.io/$PROJECT/$IMAGE \
    --set-env-vars=DB=$DB_STRING \
    --allow-unauthenticated \
    --region=us-central1 \
    --platform managed
