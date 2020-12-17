package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/serf/cmd/serf/command/agent"
	"github.com/hashicorp/serf/serf"
	log "github.com/sirupsen/logrus"
	k8sapiv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

var errNotInCluster = rest.ErrNotInCluster

func getPodName() (string, error) {
	name, ok := os.LookupEnv("K8S_POD_NAME")
	if !ok || name == "" {
		return "", errors.New("K8S_POD_NAME environment variable not set")
	}
	return name, nil
}

func populateSerfFromK8s(ctx context.Context, a *agent.Agent) error {
	// TODO: get service account from metadata server
	// TODO: get namespace from service account
	/*
		kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			clientcmd.NewDefaultClientConfigLoadingRules(),
			&clientcmd.ConfigOverrides{},
		)
		namespace, _, err := kubeconfig.Namespace()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Namespace from clientcmd: %s", namespace)
	*/

	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	k8sclient, err := corev1client.NewForConfig(config)
	if err != nil {
		return err
	}
	namespace, ok := os.LookupEnv("K8S_NAMESPACE")
	if !ok || namespace == "" {
		return errors.New("K8S_NAMESPACE environment variable not set")
	}
	myIP := a.Serf().LocalMember().Addr.String()
	log.Printf("[DEBUG] local IP according to Serf: %s", myIP)

	go func() {
		const maxSleep = 30 * time.Second
		const sleepPerLiveNode = 1 * time.Second

		for {
			ipSet := make(map[string]struct{})
			all := 0
			alive := 0
			for _, m := range a.Serf().Members() {
				all++
				if m.Status == serf.StatusAlive {
					alive++
				}
				ipSet[m.Addr.String()] = struct{}{}
			}
			toAdd := make([]string, 0)

			// TODO: prefer old-ish pods (>5 mins)
			// TODO: handle lots of pods (randomly grab some)
			list, err := k8sclient.Pods(namespace).List(ctx, k8sapiv1.ListOptions{})
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("[DEBUG] Found %d total pods via k8s", len(list.Items))
			for _, p := range list.Items {
				log.Printf("[DEBUG] main: got k8s pod IP: '%s'", p.Status.PodIP)
				if _, ok := ipSet[p.Status.PodIP]; !ok {
					toAdd = append(toAdd, p.Status.PodIP)
				}
			}
			log.Printf("[DEBUG] Joining %d new pods via k8s", len(toAdd))
			for i := range toAdd {
				toAdd[i] = fmt.Sprintf("%s:%d", toAdd[i], *controlPort)
			}
			joined, err := a.Serf().Join(toAdd, false)
			if err != nil {
				log.Printf("[ERROR] %s", err)
			}
			log.Printf("[DEBUG] Joined %d new pods via k8s", joined)

			sleepTime := time.Duration(alive) * sleepPerLiveNode
			if sleepTime > maxSleep {
				sleepTime = maxSleep
			}
			select {
			case <-ctx.Done():
			case <-time.After(sleepTime):
			}
		}
	}()

	return nil
}
