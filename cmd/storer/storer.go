package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"

	"cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/pubsub"
	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
	"github.com/vsekhar/fabula/internal/initlogrus"
	"gocloud.dev/server"
	"gocloud.dev/server/health"

	pb "github.com/vsekhar/fabula/internal/api/storer"
	"github.com/vsekhar/fabula/internal/storer"
)

var (
	pubsubSubscription = flag.String("pubsubSubscription", "", "pubsub subscription to read from")
	bucket             = flag.String("bucket", "", "bucket to write to")
	folder             = flag.String("folder", "", "object folder to write to")
	prefix             = flag.String("prefix", "", "filename prefix")
	healthCheckPort    = flag.Int("healthCheckPort", 0, "port for http health check")
	projectFlag        = flag.String("project", "", "project")
	verbose            = flag.Bool("verbose", false, "verbose log level")
	dev                = flag.Bool("dev", false, "developer mode")
)

const hashLen = 64

type healthCheck struct{}

func (healthCheck) CheckHealth() error { return nil }

func main() {
	flag.Parse()
	if *pubsubSubscription == "" {
		log.Fatal("pubsub_subscription required")
	}
	if *bucket == "" {
		log.Fatal("bucket required")
	}
	initlogrus.Init("storer", "v0.1.0", *dev, *verbose)

	project := *projectFlag
	if project == "" {
		var err error
		project, err = metadata.ProjectID()
		if err != nil {
			log.WithError(err).Fatal("getting project ID")
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// health check
	srvOptions := &server.Options{
		HealthChecks: []health.Checker{new(healthCheck)},
	}
	srv := server.New(http.DefaultServeMux, srvOptions)
	go func(ctx context.Context) {
		go func() {
			log.Info("Starting health check server...")
			if err := srv.ListenAndServe(fmt.Sprintf(":%d", *healthCheckPort)); err != nil {
				log.WithError(err).Error("health check HTTP listener")
			}
		}()
		<-ctx.Done()
		if err := srv.Shutdown(context.Background()); err != nil {
			log.WithError(err).Error("http check HTTP shutdown")
		}
	}(ctx)

	pscli, err := pubsub.NewClient(ctx, project)
	if err != nil {
		log.WithError(err).Fatal("creating pubsub client")
	}
	sub := pscli.Subscription(*pubsubSubscription)
	s, err := storer.New(ctx, *bucket, *folder, *prefix)
	if err != nil {
		log.WithError(err).Fatal("creating storer")
	}
	log.Info("Starting pubsub receive...")
	err = sub.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		defer m.Nack() // first call wins, so just Ack where appropriate below.

		fields := make(log.Fields)
		fields["pubsub_id"] = m.ID
		fields["pubsub_publish_time"] = m.PublishTime.String()
		fields["pubsub_data"] = m.Data

		var req pb.StoreRequest
		if err := proto.Unmarshal(m.Data, &req); err != nil {
			log.WithFields(fields).WithError(err).Error("unmarshalling")
			return
		}
		fields["pubsub_batches"] = len(req.Batch.Batches)
		fields["pubsub_entries"] = len(req.Batch.Entries)
		fields["pubsub_sha3512"] = req.Ref.BatchSha3512

		objName, stored, err := s.StoreOrVerify(ctx, &req)
		fields["objName"] = objName
		if err != nil {
			log.WithFields(fields).WithError(err).Error("storing")
			return
		}
		m.Ack()
		if stored {
			log.WithFields(fields).Error("batcher failed to write object, storer wrote it")
		} else {
			log.WithFields(fields).Info("storer verified")
		}
	})
	if err != nil {
		log.WithError(err).Error("receive terminated")
	}
	log.Info("Terminating")
}
