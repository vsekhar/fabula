package streamreduce_test

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	gocmd "github.com/go-cmd/cmd"
)

// TODO: use gcloud beta emulators pubsub
// https://cloud.google.com/pubsub/docs/emulator

// The unique, user-assigned ID of the Project. It must be 6 to 30 lowercase
// letters, digits, or hyphens. It must start with a letter. Trailing hyphens
// are prohibited.
//
// https://cloud.google.com/resource-manager/reference/rest/v1/projects#Project
const projectId = "streamreduce-test-project-for-emulator-only"

func signal(c *gocmd.Cmd, sig os.Signal) {
	s := c.Status()
	if s.StartTs > 0 {
		p, err := os.FindProcess(s.PID)
		if err != nil {
			log.Printf("failed to find process %d", s.PID)
		}
		err = p.Signal(sig)
		if err != nil {
			log.Printf("error signalling process %d: %s", s.PID, err)
		}
	}
}

func setup(ctx context.Context) (client *pubsub.Client, err error) {
	gcloud, err := exec.LookPath("gcloud")
	if err != nil {
		return nil, err
	}
	// Don't use CommandContext as it kills the process, rather than Ctrl-C
	cmd := gocmd.NewCmdOptions(
		gocmd.Options{Streaming: true, Buffered: true},
		gcloud,
		"beta", // TODO: still beta?
		"emulators",
		"pubsub",
		"start",
		"--project="+projectId,
	)
	cmd.Start()
	go func() {
		<-ctx.Done()
		log.Print("killing emulator")
		signal(cmd, os.Interrupt)
	}()
	defer func() {
		if err != nil {
			log.Print("killing emulator")
			signal(cmd, os.Interrupt)
		}
	}()

	// wait up to 5 seconds for server to come up
	serverPreamble := "[pubsub] INFO: Server started, listening on "
	timeoutCh := time.After(5 * time.Second)
	port := 0
loop:
	for {
		select {
		case line := <-cmd.Stderr:
			if strings.HasPrefix(line, serverPreamble) {
				p := strings.TrimPrefix(line, serverPreamble)
				var err error
				port, err = strconv.Atoi(p)
				if err != nil {
					return nil, fmt.Errorf("error parsing port '%s'", p)
				}
				break loop
			}
		case <-timeoutCh:
			return nil, fmt.Errorf("timed out waiting for emulator to come up")
		}
	}

	// FYI: For some reason running "gcloud beta emulators pubsub env-init"
	// gives a wrong port number, hence parsing it out from server output above.

	host := fmt.Sprintf("localhost:%d", port)
	copts := []option.ClientOption{
		option.WithEndpoint(host),
		option.WithoutAuthentication(),
	}

	os.Setenv("PUBSUB_EMULATOR_HOST", host)
	client, err = pubsub.NewClient(ctx, projectId, copts...)
	os.Unsetenv("PUBSUB_EMULATOR_HOST")
	if err != nil {
		return nil, err
	}
	return client, nil
}

func TestClient(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client, err := setup(ctx)
	if err != nil {
		t.Fatal(err)
	}
	topics := client.Topics(ctx)
	n := 0
	for {
		_, err := topics.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		n++
	}
	if n != 0 {
		t.Errorf("expected 0 topics, got %d", n)
	}
	topic, err := client.CreateTopic(ctx, "testtopic")
	if err != nil {
		t.Fatal(err)
	}
	topic.EnableMessageOrdering = true
	sub, err := client.CreateSubscription(ctx, "testsub", pubsub.SubscriptionConfig{
		Topic:                 topic,
		RetentionDuration:     10 * time.Minute,
		ExpirationPolicy:      24 * time.Hour,
		EnableMessageOrdering: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	m1 := &pubsub.Message{
		Data:        []byte("123"),
		OrderingKey: "123",
	}
	res := topic.Publish(ctx, m1)
	_, err = res.Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	sub.Receive(ctx, func(ctx context.Context, m2 *pubsub.Message) {
		if !bytes.Equal(m1.Data, m2.Data) {
			t.Errorf("mismatched data: %s and %s", m1.Data, m2.Data)
		}
		if m1.OrderingKey != m2.OrderingKey {
			t.Errorf("mismatched ordering keys: %s, %s", m1.OrderingKey, m2.OrderingKey)
		}
		t.Logf("received message: %+v", m2)
		m2.Ack()
		cancel()
	})
}
