#!/usr/bin/env bash

set -ex

NAME=fabula

PROJECT=$(gcloud config get-value project)
SERVICE=$NAME
IMAGE=$NAME-image
SPANNER_INSTANCE=$NAME-spanner-instance
SPANNER_DB=$NAME-spanner-db
DB_STRING=projects/$PROJECT/instances/$SPANNER_INSTANCE/databases/$SPANNER_DB

# spanner
gcloud deployment-manager deployments update $NAME --config dmconfig.yaml || \
gcloud deployment-manager deployments create $NAME --config dmconfig.yaml

gcloud spanner databases create $SPANNER_DB --instance=$SPANNER_INSTANCE || true

gcloud spanner databases ddl update $SPANNER_DB \
    --instance=$SPANNER_INSTANCE \
    --ddl="$(cat storage/spanner.sdl)" || \
    read -p "Press Ctrl-C to cancel or any key to continue..."

# server in us-central1
(cd server && exec go mod tidy)
gcloud builds submit server --tag gcr.io/$PROJECT/$IMAGE
gcloud run deploy $SERVICE \
    --image gcr.io/$PROJECT/$IMAGE \
    --set-env-vars=DB=$DB_STRING \
    --allow-unauthenticated \
    --region=us-central1 \
    --platform managed
