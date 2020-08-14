# FROM golang:1.13.6-buster as build
FROM golang@sha256:93a56423351235e070b3630e0a8b3e27d5e868883d4dff591f676315f208a574 as builder
WORKDIR /app
ADD . /app
RUN addgroup fabulagroup
RUN adduser --disabled-password --gecos "" --no-create-home --ingroup fabulagroup fabulauser
RUN go mod download
RUN go mod verify
RUN CGO_ENABLED=0 go build -v -ldflags="-w -s" -o fabula-server
FROM gcr.io/distroless/static
COPY --from=builder /app/fabula-server /fabula-server
COPY --from=builder /etc/passwd /etc/passwd
USER fabulauser
ENTRYPOINT ["/fabula-server"]
