set -e

echo "Runing Go Format on the following files:"
go fmt github.com/intel/cndp_device_plugin/...

echo "Running unit tests:"
go test -v github.com/intel/cndp_device_plugin/...

echo "Building Device Plugin"
go build -o ./bin/cndp-dp ./cmd/cndp-dp

echo "Building CNI"
go build -o ./bin/cndp ./cmd/cndp-cni

echo "Build complete!"
