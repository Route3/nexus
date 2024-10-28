package tests

func LoadSingleTestConfig() (TestConfig) {

    return TestConfig{
    	masterAccountPrivateKey: "<ENTER PRIVATE KEY>",
    	rpcUrls: []string{
    		"http://127.0.0.1:8545",
    	},
    }

}

func LoadMultiTestConfig() (TestConfig) {

    return TestConfig{
    	masterAccountPrivateKey: "<ENTER PRIVATE KEY>",
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
rpcUrls []string
}
