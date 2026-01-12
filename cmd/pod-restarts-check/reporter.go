package main

import (
	"os"

	"github.com/kuberhealthy/kuberhealthy/v3/pkg/checkclient"
	log "github.com/sirupsen/logrus"
)

// reportSuccess reports success to Kuberhealthy.
func reportSuccess() error {
	// Report success to Kuberhealthy.
	err := checkclient.ReportSuccess()
	if err != nil {
		log.Println("Error reporting success to Kuberhealthy servers:", err)
		return err
	}
	log.Println("Successfully reported success to Kuberhealthy servers")
	return nil
}

// reportFailure reports failure to Kuberhealthy.
func reportFailure(errorMessages []string) error {
	// Report failure to Kuberhealthy.
	err := checkclient.ReportFailure(errorMessages)
	if err != nil {
		log.Println("Error reporting failure to Kuberhealthy servers:", err)
		return err
	}
	log.Println("Successfully reported failure to Kuberhealthy servers")
	return nil
}

// reportFailureAndExit reports failure to Kuberhealthy and exits.
func reportFailureAndExit(err error) {
	// Report failure to Kuberhealthy.
	reportErr := checkclient.ReportFailure([]string{err.Error()})
	if reportErr != nil {
		log.Fatalln("error when reporting to kuberhealthy with error:", reportErr)
	}
	log.Infoln("Successfully reported error to kuberhealthy")
	os.Exit(0)
}
