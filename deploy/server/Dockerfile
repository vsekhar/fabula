# Invoke with repo root as context and this file as a parameter
#
#    deploy/server$ docker build --file Dockerfile ../../

# FROM golang:1.13.6-buster as build
FROM golang@sha256:93a56423351235e070b3630e0a8b3e27d5e868883d4dff591f676315f208a574 as builder

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
