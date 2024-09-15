#!/bin/bash

set -e

# Stop and remove geth
docker compose rm -f

# Reset state
sudo rm -rf ./shared
mkdir shared

# Generate nexus secrets
go run main.go secrets init --data-dir ./shared/nexus

# Generate jwt.hex
openssl rand -hex 32 | tr -d '\n' >'./shared/jwt.hex'

# Generate Nexus Genesis
go run main.go secrets output --data-dir ./shared/nexus --json | jq -j .node_id >./shared/nexus/node_id
go run main.go secrets output --data-dir ./shared/nexus --json | jq -j .address >./shared/nexus/validator_key
go run main.go genesis --consensus ibft --ibft-validator-type ecdsa --ibft-validator $(cat ./shared/nexus/validator_key) --bootnode /ip4/127.0.0.1/tcp/1478/p2p/$(cat ./shared/nexus/node_id)/ --dir ./shared/nexus-genesis.json

# Run geth
docker compose up -d

# Run nexus
go run main.go server --log-level DEBUG --config nexus-config.yaml
