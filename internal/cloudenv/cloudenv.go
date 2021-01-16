// Package cloudenv provides instance identity and peer information.
package cloudenv

import (
	"context"
	"expvar"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/compute/metadata"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/compute/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

const (
	localEnvString  = "local"
	k8sEnvString    = "k8s"
	gceMigEnvString = "gceMig"
)

var expvarMap *expvar.Map
var peerMap *expvar.Map

func init() {
	expvarMap = expvar.NewMap("cloudenv")
	peerMap = expvar.NewMap("cloudenv.last_peers")
}

// CloudEnv provides methods to describe the current running instance and its
// peers using information provided by the cloud runtime environment.
type CloudEnv interface {
	CloudEnv() string
	Hostname() string
	Peers(ctx context.Context, filter string) ([]string, error)
}

type localEnv struct {
	hostname string
}

func (localEnv) CloudEnv() string                                { return localEnvString }
func (l *localEnv) Hostname() string                             { return l.hostname }
func (localEnv) Peers(context.Context, string) ([]string, error) { return nil, nil }

type k8sEnv struct {
	hostname  string
	namespace string
	k8sclient *corev1client.CoreV1Client
}

func (k8sEnv) CloudEnv() string    { return k8sEnvString }
func (k *k8sEnv) Hostname() string { return k.hostname }
func (k *k8sEnv) Peers(ctx context.Context, filter string) ([]string, error) {
	podList, err := k.k8sclient.Pods(k.namespace).List(ctx, v1.ListOptions{
		LabelSelector: filter,
	})
	if err != nil {
		return nil, err
	}
	log.WithField("n", len(podList.Items)).Debug("cloudenv.Peers: fetched Kubernetes pod list")
	r := make([]string, podList.Size())
	now := time.Now()
	for i, pod := range podList.Items {
		r[i] = pod.Status.PodIP
		peerMap.Set(pod.Status.PodIP, now)
	}
	return r, nil
}

type gceMigEnv struct {
	hostname   string
	project    string
	createdBy  string
	zones      []string
	inMig      bool
	computeSvc *compute.Service
}

func (gceMigEnv) CloudEnv() string    { return gceMigEnvString }
func (g *gceMigEnv) Hostname() string { return g.hostname }
func (g *gceMigEnv) Peers(ctx context.Context, filter string) ([]string, error) {
	if !g.inMig {
		return nil, nil
	}

	// Search for instances in all zones
	zctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	zch := make(chan []string)
	for _, zone := range g.zones {
		go func(zone string) {
			r := make([]string, 0)

			// TODO: can't get instance list with filter using g.createdBy:
			//   metadata.items.key['created-by']['value']='projects/523215995107/regions/us-central1/instanceGroupManagers/svc-test-external'
			// So get full list in all zones and filter manually?
			// Or get list of instance names from rigm then look up IPs (maybe
			// cache names and IPs somewhere).
			req := g.computeSvc.Instances.List(g.project, zone).Filter(
				"metadata.items.key['created-by']['value']='" + g.createdBy + "'",
			)
			err := req.Pages(zctx, func(page *compute.InstanceList) error {
				for _, instance := range page.Items {
					// TODO: filter network interfaces to match the network we're on
					for _, iface := range instance.NetworkInterfaces {
						r = append(r, iface.NetworkIP)
					}
				}
				return nil
			})
			if err != nil {
				log.WithFields(log.Fields{
					"hostname":    g.hostname,
					"createdBy":   g.createdBy,
					"currentZone": zone,
					"zones":       g.zones,
				}).WithError(err).WithContext(zctx).Error("fetching instances")
				return
			}
			zch <- r
		}(zone)
	}
	var ips []string
	for i := 0; i < len(g.zones); i++ {
		select {
		case ip := <-zch:
			ips = append(ips, ip...)
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if len(ips) == 0 {
		log.Debugf("cloudenv.gceMigEnv: no peers found in %d zones", len(g.zones))
	} else {
		log.Debugf("cloudenv.gceMigEnv: %d peers found in %d zones", len(ips), len(g.zones))
	}
	return ips, nil
}

// Get returns a CloudEnv object for the current running process, or an error.
func Get(ctx context.Context, logName string, dev, verbose bool) (CloudEnv, error) {
	log.SetFormatter(&log.TextFormatter{}) // default, overwritten below
	if verbose {
		log.SetLevel(log.TraceLevel)
	} else if dev {
		log.SetLevel(log.WarnLevel)
	} else {
		log.SetLevel(log.DebugLevel)
	}

	// Kubernetes
	_, ink8s := os.LookupEnv("KUBERNETES_SERVICE_HOST")
	if ink8s {
		// TODO: for now we require the pod name to be passed as an environment
		// variable as part of the "downward API".
		hostname, ok := os.LookupEnv("KUBERNETES_POD_NAME")
		if !ok {
			return nil, fmt.Errorf("'KUBERNETES_POD_NAME' must be set on kubernetes")
		}
		config, err := rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
		k8sclient, err := corev1client.NewForConfig(config)
		if err != nil {
			return nil, err
		}
		// https://stackoverflow.com/a/46046153/608382
		namespaceb, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
		if err != nil {
			return nil, err
		}
		namespace := strings.TrimSpace(string(namespaceb))

		// Logging
		if !dev {
			log.SetFormatter(&log.JSONFormatter{})
		}

		log.WithFields(log.Fields{
			"hostname":  hostname,
			"namespace": namespace,
		}).Debug("cloudenv: detected Kubernetes")

		return &k8sEnv{
			hostname:  hostname,
			namespace: namespace,
			k8sclient: k8sclient,
		}, nil
	}

	// GCE Managed Instance Group (MIG)
	if metadata.OnGCE() {
		hostname, err := metadata.Hostname()
		if err != nil {
			return nil, err
		}
		projectName, err := metadata.ProjectID()
		if err != nil {
			return nil, fmt.Errorf("cloudenv: getting project ID: %w", err)
		}

		createdBy, err := metadata.Get("instance/attributes/created-by")
		if err != nil {
			return &gceMigEnv{
				hostname: hostname,
				project:  projectName,
				inMig:    false,
			}, nil
		}
		// Format:
		//      Zonal: "projects/123456789012/zones/us-central1-f/instanceGroupManagers/igm-metadata"
		//   Regional: "projects/123456789012/regions/us-central1/instanceGroupManagers/igm-metadata"
		//
		// https://cloud.google.com/compute/docs/instance-groups/getting-info-about-migs#checking_if_a_vm_instance_is_part_of_a_mig
		parts := strings.Split(createdBy, "/")
		if len(parts) != 6 {
			return nil, fmt.Errorf("cloudenv: bad GCP created-by response: %s", createdBy)
		}

		computeSvc, err := compute.NewService(ctx)
		if err != nil {
			return nil, fmt.Errorf("cloudenv: creating Google compute service client: %w", err)
		}

		var zones []string
		switch parts[2] {
		case "regions":
			region := parts[3]
			filter := fmt.Sprintf("region=\"https://www.googleapis.com/compute/v1/projects/fabula-8589/regions/%s\"", region)
			req := computeSvc.Zones.List(projectName).Filter(filter)
			if err := req.Pages(ctx, func(page *compute.ZoneList) error {
				for _, zone := range page.Items {
					zones = append(zones, zone.Name)
				}
				return nil
			}); err != nil {
				return nil, fmt.Errorf("cloudenv: getting zones: %w", err)
			}
		case "zones":
			zones = []string{parts[3]}
		default:
			return nil, fmt.Errorf("cloudenv: createdBy[2] must be in {regions,zones}: %s", createdBy)
		}

		// Logging
		if !dev {
			hook, err := newStackDriverHook(logName, projectName)
			if err != nil {
				return nil, fmt.Errorf("cloudenv: creating Stack Driver hook: %w", err)
			}
			log.AddHook(hook)
		}

		log.WithFields(log.Fields{
			"hostname":   hostname,
			"created-by": createdBy,
			"zones":      zones,
		}).Debug("cloudenv: detected GCE MIG")

		return &gceMigEnv{
			hostname:   hostname,
			project:    projectName,
			createdBy:  createdBy,
			zones:      zones,
			inMig:      true,
			computeSvc: computeSvc,
		}, nil
	}

	// Local
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"hostname": hostname,
	}).Debug("cloudenv: detected local")

	return &localEnv{hostname: hostname}, nil
}
