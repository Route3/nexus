package tests

import (
	"os"
)

func LoadSingleTestConfig() TestConfig {

	return TestConfig{
		masterAccountPrivateKey: loadPrivateKey(),
		rpcUrls: []string{
			"http://127.0.0.1:8545",
		},
	}
}

func LoadMultiTestConfig() TestConfig {

	return TestConfig{
		masterAccountPrivateKey: loadPrivateKey() ,
		rpcUrls: []string{
			"http://127.0.0.1:8545",
			"http://127.0.0.1:18545",
			"http://127.0.0.1:28545",
			"http://127.0.0.1:38545",
		},
	}
}

type TestConfig struct {
	masterAccountPrivateKey string
	rpcUrls                 []string
}

func loadPrivateKey() string {
	privateKey, privateKeySet := os.LookupEnv("MASTER_ACCOUNT_PRIVATE_KEY")
	if !privateKeySet {
		panic("MASTER_ACCOUNT_PRIVATE_KEY not set")
	}
	return privateKey
}