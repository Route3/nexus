#!/bin/bash

set -e

# Parse flags
while getopts "ng" flag; do
  case "$flag" in
  n) option_n=true ;;
  g) option_g=true ;;
  *) echo "Invalid option" ;;
  esac
done

set -ex

if [ $option_g ]; then
  echo "Resetting geth state"

  # Reset geth state
  sudo rm -rf ./shared/geth ./shared/geth-genesis.json

  # Stop and remove geth
  docker compose -f docker-compose.single.yaml stop
  docker compose -f docker-compose.single.yaml rm -f
fi

if [ $option_n ]; then
  echo "Resetting nexus state"

  # Reset nexus state
  sudo rm -rf ./shared/nexus-genesis.json ./shared/nexus ./shared/jwt.hex

  # Generate nexus secrets
  go run main.go secrets init --data-dir ./shared/nexus

  # Generate jwt.hex
  openssl rand -hex 32 | tr -d '\n' >'./shared/jwt.hex'

  # Generate Nexus Genesis
  go run main.go secrets output --data-dir ./shared/nexus --json | jq -j .node_id >./shared/nexus/node_id
  go run main.go secrets output --data-dir ./shared/nexus --json | jq -j .address >./shared/nexus/validator_key
  go run main.go genesis --consensus ibft --ibft-validator-type ecdsa --ibft-validator $(cat ./shared/nexus/validator_key) --bootnode /ip4/127.0.0.1/tcp/1478/p2p/$(cat ./shared/nexus/node_id)/ --dir ./shared/nexus-genesis.json
fi

# Run geth
docker compose -f docker-compose.single.yaml up -d || true

# Run nexus
go run main.go server --log-level DEBUG --config nexus-config.yaml
