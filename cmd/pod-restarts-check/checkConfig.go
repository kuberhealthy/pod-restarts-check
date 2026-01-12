package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/kuberhealthy/kuberhealthy/v3/pkg/checkclient"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

const (
	// defaultMaxFailuresAllowed is the default restart threshold.
	defaultMaxFailuresAllowed = 10
	// defaultCheckTimeout is the default runtime timeout.
	defaultCheckTimeout = 10 * time.Minute
)

// CheckConfig stores configuration for the pod restarts check.
type CheckConfig struct {
	// Namespace is the namespace to scan for events.
	Namespace string
	// CheckTimeout is the runtime timeout for the check.
	CheckTimeout time.Duration
	// MaxFailuresAllowed is the restart threshold.
	MaxFailuresAllowed int32
	// KubeConfigFile is the optional kubeconfig path.
	KubeConfigFile string
}

// parseConfig reads environment variables and builds a CheckConfig.
func parseConfig() (*CheckConfig, error) {
	// Parse namespace selection.
	namespace := os.Getenv("POD_NAMESPACE")
	if len(namespace) == 0 {
		log.Infoln("Looking for pods across all namespaces, this requires a cluster role")
		namespace = v1.NamespaceAll
	}
	if len(namespace) != 0 && namespace != v1.NamespaceAll {
		log.Infoln("Looking for pods in namespace:", namespace)
	}

	// Set check time limit to default.
	checkTimeout := defaultCheckTimeout

	// Override using the Kuberhealthy deadline when available.
	deadline, err := checkclient.GetDeadline()
	if err != nil {
		log.Infoln("There was an issue getting the check deadline:", err.Error())
	}
	checkTimeout = deadline.Sub(time.Now().Add(time.Second * 5))
	log.Infoln("Check time limit set to:", checkTimeout)

	// Parse max failures allowed.
	maxFailuresAllowed := int32(defaultMaxFailuresAllowed)
	maxFailuresEnv := os.Getenv("MAX_FAILURES_ALLOWED")
	if len(maxFailuresEnv) != 0 {
		conversion, parseErr := strconv.ParseInt(maxFailuresEnv, 10, 32)
		if parseErr != nil {
			return nil, fmt.Errorf("error converting MAX_FAILURES_ALLOWED: %w", parseErr)
		}
		maxFailuresAllowed = int32(conversion)
	}

	// Assemble configuration.
	cfg := &CheckConfig{}
	cfg.Namespace = namespace
	cfg.CheckTimeout = checkTimeout
	cfg.MaxFailuresAllowed = maxFailuresAllowed
	cfg.KubeConfigFile = os.Getenv("KUBECONFIG")

	return cfg, nil
}
