IMAGE := "kuberhealthy/pod-restarts-check"
TAG := "latest"

# Build the pod restarts check container locally.
build:
	podman build -f Containerfile -t {{IMAGE}}:{{TAG}} .

# Run the unit tests for the pod restarts check.
test:
	go test ./...

# Build the pod restarts check binary locally.
binary:
	go build -o bin/pod-restarts-check ./cmd/pod-restarts-check
