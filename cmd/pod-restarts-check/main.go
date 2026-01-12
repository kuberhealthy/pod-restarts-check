package main

import (
	"os"

	log "github.com/sirupsen/logrus"
)

// main loads configuration and executes the pod restarts check.
func main() {
	// Parse configuration from environment variables.
	cfg, err := parseConfig()
	if err != nil {
		reportFailureAndExit(err)
		return
	}

	// Create a Kubernetes client.
	client, err := createKubeClient(cfg.KubeConfigFile)
	if err != nil {
		log.Fatalln("Unable to create kubernetes client", err)
	}

	// Create a new pod restarts checker.
	prc := NewChecker(cfg, client)

	// Run the check.
	err = prc.Run()
	if err != nil {
		log.Errorln("Error running Pod Restarts check:", err)
		os.Exit(2)
	}
	log.Infoln("Done running Pod Restarts check")
	os.Exit(0)
}
