set -e

go build -o ./bin/cndp-dp ./cmd/cndp-dp
go build -o ./bin/cndp ./cmd/cndp-cni
