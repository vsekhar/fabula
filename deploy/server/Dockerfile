# Invoke with repo root as context and this file as a parameter
#
#    deploy/server$ docker build --file Dockerfile ../../

# FROM golang:1.15.2-alpine amd64 as builder
FROM golang@sha256:fc801399d044a8e01f125eeb5aa3f160a0d12d6e03ba17a1d0b22ce50dfede81 as builder

# create non-root user from here (distroless/static doesn't have a shell)
RUN addgroup fabulagroup
RUN adduser --disabled-password --gecos "" --no-create-home --ingroup fabulagroup fabulauser

WORKDIR /app

# Get dependencies early so they can be cached by Docker
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
ADD go.mod /app
ADD go.sum /app
RUN go mod download
RUN go mod verify

# build app
ADD . /app
RUN go build -v -ldflags='-w -s -extldflags "-static"' -o fabula-server ./cmd/server

# runtime image
FROM gcr.io/distroless/static
COPY --from=builder /app/fabula-server /fabula-server
COPY --from=builder /etc/passwd /etc/passwd
USER fabulauser
ENTRYPOINT ["/fabula-server"]
