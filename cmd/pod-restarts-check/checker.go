package main

import (
	"context"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Checker represents a long running pod restart checker.
type Checker struct {
	// Namespace is the namespace to inspect.
	Namespace string
	// MaxFailuresAllowed is the restart threshold.
	MaxFailuresAllowed int32
	// BadPods contains pods exceeding the threshold.
	BadPods map[string]string
	// client is the Kubernetes client.
	client *kubernetes.Clientset
	// checkTimeout is the runtime timeout for the check.
	checkTimeout time.Duration
}

// NewChecker creates a new pod restart checker for a specific namespace.
func NewChecker(cfg *CheckConfig, client *kubernetes.Clientset) *Checker {
	// Build the checker instance.
	return &Checker{
		Namespace:          cfg.Namespace,
		MaxFailuresAllowed: cfg.MaxFailuresAllowed,
		BadPods:            make(map[string]string),
		client:             client,
		checkTimeout:       cfg.CheckTimeout,
	}
}

// Run starts the check and reports results.
func (prc *Checker) Run() error {
	// TODO: refactor function to receive context on exported function in next breaking change.
	ctx := context.TODO()

	// Log the start of the check.
	log.Infoln("Running Pod Restarts checker")
	doneChan := make(chan error)

	// Run the check in the background and notify on completion.
	go prc.runChecksAsync(ctx, doneChan)

	// Wait for either a timeout or job completion.
	select {
	case <-time.After(prc.checkTimeout):
		// The check has timed out after its specified timeout period.
		errorMessage := "Failed to complete Pod Restart check in time! Timeout was reached."
		return reportFailure([]string{errorMessage})
	case err := <-doneChan:
		if len(prc.BadPods) != 0 || err != nil {
			var errorMessages []string
			if err != nil {
				log.Error(err)
				errorMessages = append(errorMessages, err.Error())
			}
			for _, msg := range prc.BadPods {
				errorMessages = append(errorMessages, msg)
			}
			return reportFailure(errorMessages)
		}
		return reportSuccess()
	}
}

// runChecksAsync executes checks and sends the result to the channel.
func (prc *Checker) runChecksAsync(ctx context.Context, doneChan chan error) {
	// Run the checks and send the outcome.
	err := prc.doChecks(ctx)
	doneChan <- err
}

// doChecks scans events and records pods with excessive restarts.
func (prc *Checker) doChecks(ctx context.Context) error {
	// Look for BackOff events.
	log.Infoln("Checking for pod BackOff events for all pods in the namespace:", prc.Namespace)

	podWarningEvents, err := prc.client.CoreV1().Events(prc.Namespace).List(ctx, metav1.ListOptions{FieldSelector: "type=Warning"})
	if err != nil {
		return err
	}

	if len(podWarningEvents.Items) != 0 {
		log.Infoln("Found `Warning` events in the namespace:", prc.Namespace)

		for _, event := range podWarningEvents.Items {
			// Check for pods with BackOff events greater than the threshold.
			if event.InvolvedObject.Kind == "Pod" && event.Reason == "BackOff" && event.Count > prc.MaxFailuresAllowed {
				errorMessage := "Found: " + strconv.FormatInt(int64(event.Count), 10) + " `BackOff` events for pod: " + event.InvolvedObject.Name + " in namespace: " + event.Namespace

				log.Infoln(errorMessage)

				// We could be checking for pods in all namespaces so prefix the namespace.
				prc.BadPods[event.InvolvedObject.Namespace+"/"+event.InvolvedObject.Name] = errorMessage
			}
		}
	}

	for pod := range prc.BadPods {
		err = prc.verifyBadPodRestartExists(ctx, pod)
		if err != nil {
			return err
		}
	}

	return err
}

// verifyBadPodRestartExists removes bad pods that no longer exist.
func (prc *Checker) verifyBadPodRestartExists(ctx context.Context, pod string) error {
	// Pod is in the form namespace/pod_name.
	parts := strings.Split(pod, "/")
	namespace := parts[0]
	podName := parts[1]

	_, err := prc.client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) || strings.Contains(err.Error(), "not found") {
			log.Infoln("Bad Pod:", podName, "no longer exists. Removing from bad pods map")
			delete(prc.BadPods, podName)
			return nil
		}
		log.Infoln("Error getting bad pod:", podName, err)
		return err
	}

	return nil
}
