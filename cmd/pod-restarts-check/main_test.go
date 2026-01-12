package main

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// pod builds a test pod with a single container status.
func pod(podName string, containerName string, restartCount int32) *v1.Pod {
	// Assemble the pod metadata and status.
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test-namespace",
			Name:      podName,
		},
		Spec: v1.PodSpec{},
		Status: v1.PodStatus{
			Reason: "Ready",
			ContainerStatuses: []v1.ContainerStatus{
				{
					Name:         containerName,
					Ready:        true,
					RestartCount: restartCount,
				},
			},
		},
	}

	return &pod
}

// makeRestartObservationsMap builds a nested map of restart counts.
func makeRestartObservationsMap(podName string, containerName string, restartCount int32) map[string]map[string]int32 {
	// Create the inner map for the container.
	firstMap := make(map[string]int32)
	firstMap[containerName] = restartCount

	// Create the outer map for the pod.
	restartObservationsMap := make(map[string]map[string]int32)
	restartObservationsMap[podName] = firstMap

	return restartObservationsMap
}

// TestMakeRestartObservationsMap validates the helper map builder.
func TestMakeRestartObservationsMap(t *testing.T) {
	// Build a test map.
	result := makeRestartObservationsMap("pod-a", "container-a", 3)

	// Assert the pod key exists.
	containerMap, ok := result["pod-a"]
	if !ok {
		t.Fatalf("expected pod key to exist")
	}

	// Assert the restart count matches.
	restartCount, ok := containerMap["container-a"]
	if !ok {
		t.Fatalf("expected container key to exist")
	}
	if restartCount != 3 {
		t.Fatalf("expected restart count 3, got %d", restartCount)
	}
}
