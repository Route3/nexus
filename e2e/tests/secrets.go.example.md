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
    	masterAccountPrivateKey: "8fa8b35d390d5049f1ebc794d7631a00526e35421c92224a2e4a4cb58ea46919",
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
