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

  # ================================================ node 0

  # Reset geth state
  sudo rm -rf ./multi-validator-shared/v-0/geth ./multi-validator-shared/v-0/geth-genesis.json

  # ================================================ node 1

  # Reset geth state
  sudo rm -rf ./multi-validator-shared/v-1/geth ./multi-validator-shared/v-1/geth-genesis.json

  # ================================================ node 2

  # Reset geth state
  sudo rm -rf ./multi-validator-shared/v-2/geth ./multi-validator-shared/v-2/geth-genesis.json


  # Stop and remove geth
  docker compose -f docker-compose.multi.yaml stop
  docker compose -f docker-compose.multi.yaml rm -f
fi

if [ $option_n ]; then
  echo "Resetting nexus state"

  # ================================================ node 0

  # Reset nexus state
  sudo rm -rf ./multi-validator-shared/v-0/nexus-genesis.json ./multi-validator-shared/v-0/nexus ./multi-validator-shared/v-0/jwt.hex

  # Generate nexus secrets
  go run main.go secrets init --data-dir ./multi-validator-shared/v-0/nexus

  # Generate jwt.hex
  openssl rand -hex 32 | tr -d '\n' >'./multi-validator-shared/v-0/jwt.hex'

  cp v-0-nodekey ./multi-validator-shared/v-0/

  # Generate Nexus Genesis info
  go run main.go secrets output --data-dir ./multi-validator-shared/v-0/nexus --json | jq -j .node_id >./multi-validator-shared/v-0/nexus/node_id
  go run main.go secrets output --data-dir ./multi-validator-shared/v-0/nexus --json | jq -j .address >./multi-validator-shared/v-0/nexus/validator_key

  # ================================================ node 1

  # Reset nexus state
  sudo rm -rf ./multi-validator-shared/v-1/nexus-genesis.json ./multi-validator-shared/v-1/nexus ./multi-validator-shared/v-1/jwt.hex

  # Generate nexus secrets
  go run main.go secrets init --data-dir ./multi-validator-shared/v-1/nexus

  # Generate jwt.hex
  openssl rand -hex 32 | tr -d '\n' >'./multi-validator-shared/v-1/jwt.hex'

  # Generate Nexus Genesis info
  go run main.go secrets output --data-dir ./multi-validator-shared/v-1/nexus --json | jq -j .node_id >./multi-validator-shared/v-1/nexus/node_id
  go run main.go secrets output --data-dir ./multi-validator-shared/v-1/nexus --json | jq -j .address >./multi-validator-shared/v-1/nexus/validator_key

  # ================================================ node 2

  # Reset nexus state
  sudo rm -rf ./multi-validator-shared/v-2/nexus-genesis.json ./multi-validator-shared/v-2/nexus ./multi-validator-shared/v-2/jwt.hex

  # Generate nexus secrets
  go run main.go secrets init --data-dir ./multi-validator-shared/v-2/nexus

  # Generate jwt.hex
  openssl rand -hex 32 | tr -d '\n' >'./multi-validator-shared/v-2/jwt.hex'

  # Generate Nexus Genesis info
  go run main.go secrets output --data-dir ./multi-validator-shared/v-2/nexus --json | jq -j .node_id >./multi-validator-shared/v-2/nexus/node_id
  go run main.go secrets output --data-dir ./multi-validator-shared/v-2/nexus --json | jq -j .address >./multi-validator-shared/v-2/nexus/validator_key

  # ================================================ nodes all

  # Generate Nexus Genesis
  go run main.go genesis --consensus ibft --ibft-validator-type ecdsa \
    --ibft-validator $(cat ./multi-validator-shared/v-0/nexus/validator_key) \
    --ibft-validator $(cat ./multi-validator-shared/v-1/nexus/validator_key) \
    --ibft-validator $(cat ./multi-validator-shared/v-2/nexus/validator_key) \
    --bootnode /ip4/127.0.0.1/tcp/10001/p2p/$(cat ./multi-validator-shared/v-0/nexus/node_id) \
    --dir ./multi-validator-shared/v-0/nexus-genesis.json 

fi

# Run geth
docker compose -f docker-compose.multi.yaml up -d || true

# Run nexus
go run main.go server --log-level DEBUG --config multi-validator-config/nexus-config-0.yaml &
sleep 1
go run main.go server --log-level DEBUG --config multi-validator-config/nexus-config-1.yaml &
sleep 1
go run main.go server --log-level DEBUG --config multi-validator-config/nexus-config-2.yaml
