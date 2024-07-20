package server

import (
	"github.com/apex-fusion/nexus/consensus"
	consensusDev "github.com/apex-fusion/nexus/consensus/dev"
	consensusDummy "github.com/apex-fusion/nexus/consensus/dummy"
	consensusIBFT "github.com/apex-fusion/nexus/consensus/ibft"
	"github.com/apex-fusion/nexus/secrets"
	"github.com/apex-fusion/nexus/secrets/awsssm"
	"github.com/apex-fusion/nexus/secrets/gcpssm"
	"github.com/apex-fusion/nexus/secrets/hashicorpvault"
	"github.com/apex-fusion/nexus/secrets/local"
)

type ConsensusType string

const (
	DevConsensus   ConsensusType = "dev"
	IBFTConsensus  ConsensusType = "ibft"
	DummyConsensus ConsensusType = "dummy"
)

var consensusBackends = map[ConsensusType]consensus.Factory{
	DevConsensus:   consensusDev.Factory,
	IBFTConsensus:  consensusIBFT.Factory,
	DummyConsensus: consensusDummy.Factory,
}

// secretsManagerBackends defines the SecretManager factories for different
// secret management solutions
var secretsManagerBackends = map[secrets.SecretsManagerType]secrets.SecretsManagerFactory{
	secrets.Local:          local.SecretsManagerFactory,
	secrets.HashicorpVault: hashicorpvault.SecretsManagerFactory,
	secrets.AWSSSM:         awsssm.SecretsManagerFactory,
	secrets.GCPSSM:         gcpssm.SecretsManagerFactory,
}

func ConsensusSupported(value string) bool {
	_, ok := consensusBackends[ConsensusType(value)]

	return ok
}
