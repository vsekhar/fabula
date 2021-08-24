REPO_PROJECT_ID=fabula-resources

.PHONY:

all: fabula-digest.txt hello-digest.txt

%-digest.txt: %.image_on_gcr
	gcloud container images list-tags \
		--format='get(digest)' \
		--filter=tags:latest gcr.io/${REPO_PROJECT_ID}/$* \
		> $@

%.image_on_gcr: .PHONY Dockerfile.% reporoot project
	gcloud builds submit ${REPOROOT} \
		--project ${REPO_PROJECT_ID} \
		--config ${REPOROOT}/deploy/containers/cloudbuild.yaml \
		--substitutions=_DOCKERFILE=deploy/containers/Dockerfile.$*,_IMAGE=$* \

reporoot: .PHONY
	$(eval REPOROOT=$(shell git rev-parse --show-toplevel))

project: .PHONY
	$(eval PROJECT=$(shell gcloud config get-value project))
