module github.com/vsekhar/fabula

go 1.15

require (
	cloud.google.com/go v0.74.0
	cloud.google.com/go/logging v1.1.2
	cloud.google.com/go/pubsub v1.9.1
	cloud.google.com/go/storage v1.12.0
	contrib.go.opencensus.io/exporter/stackdriver v0.13.4
	// https://github.com/GoogleCloudPlatform/opentelemetry-operations-go/issues/122
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric v0.13.1-0.20201210214614-d8551be9a708
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace v0.13.1-0.20201210214614-d8551be9a708
	github.com/armon/go-metrics v0.3.3 // indirect
	github.com/dchest/siphash v1.2.2 // indirect
	github.com/dgryski/go-maglev v0.0.0-20200611225407-8961b9b1b8e6
	github.com/dustin/go-humanize v1.0.0
	github.com/go-cmd/cmd v1.3.0
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e
	github.com/golang/protobuf v1.4.3
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/hashicorp/go-hclog v0.15.0
	github.com/hashicorp/go-immutable-radix v1.2.0 // indirect
	github.com/hashicorp/go-msgpack v0.5.5 // indirect
	github.com/hashicorp/go-sockaddr v1.0.2 // indirect
	github.com/hashicorp/go-uuid v1.0.2 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/serf v0.9.5
	github.com/k0kubun/go-ansi v0.0.0-20180517002512-3bf9e2903213
	github.com/kr/text v0.2.0 // indirect
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/miekg/dns v1.1.29 // indirect
	github.com/pa-m/sklearn v0.0.0-20200711083454-beb861ee48b1
	github.com/schollz/progressbar/v3 v3.7.2
	github.com/sirupsen/logrus v1.7.0
	go.opencensus.io v0.22.5
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.15.1
	go.opentelemetry.io/contrib/instrumentation/host v0.15.1
	go.opentelemetry.io/otel v0.15.0
	go.opentelemetry.io/otel/sdk v0.15.0
	gocloud.dev v0.21.0
	golang.org/x/crypto v0.0.0-20201208171446-5f87f3452ae9
	golang.org/x/oauth2 v0.0.0-20201208152858-08078c50e5b5
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a
	golang.org/x/sys v0.0.0-20201214210602-f9fddec55a1e
	golang.org/x/term v0.0.0-20201210144234-2321bbc49cbf // indirect
	golang.org/x/tools v0.0.0-20201211185031-d93e913c1a58 // indirect
	gonum.org/v1/gonum v0.8.2
	google.golang.org/api v0.36.0
	google.golang.org/genproto v0.0.0-20201214200347-8c77b98c765d
	google.golang.org/grpc v1.34.0
	google.golang.org/protobuf v1.25.0
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200603094226-e3079894b1e8 // indirect
	k8s.io/apimachinery v0.20.0
	k8s.io/client-go v0.20.0
)
