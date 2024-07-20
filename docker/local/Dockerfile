FROM golang:1.18-alpine AS builder

WORKDIR /nexus

ADD go.mod go.sum ./
RUN go mod download

RUN apk add git

COPY . .

RUN go build -o nexus -tags netgo -ldflags="\
  -s -w \
  -X 'github.com/apex-fusion/nexus/versioning.Version=$(git describe --tags --abbrev=0)' \
  -X 'github.com/apex-fusion/nexus/versioning.Commit=$(git rev-parse HEAD)' \
  -X 'github.com/apex-fusion/nexus/versioning.Branch=$(git rev-parse --abbrev-ref HEAD | tr -d '\040\011\012\015\n')' \
  -X 'github.com/apex-fusion/nexus/versioning.BuildTime=$(date)'" \
  main.go

FROM alpine:latest AS runner

RUN apk --no-cache add ca-certificates jq

WORKDIR /nexus

COPY --from=builder /nexus/nexus ./
COPY ./docker/local/nexus.sh ./

# Expose json-rpc, libp2p and grpc ports
EXPOSE 8545 9632 1478 5001

ENTRYPOINT ["./entrypoint.sh"]
