package autotest

import (
	"fmt"
	"os"
	"testing"
)

var testClient *OCITestClient
var testConfig *TestConfiguration

var testingServiceEnabled = true

func TestMain(m *testing.M) {

	if _, ok := os.LookupEnv("AUTOTEST_DISABLE_SERVICE"); ok {
		fmt.Println("Running auto tests without testing service")
		testingServiceEnabled = false
	}

	if !testingServiceEnabled {
		os.Exit(m.Run())
	}

	var err error
	testConfig, err = newTestConfiguration()
	if err != nil {
		panic(err)
	}

	testClient = NewOCITestClient()
	err = testClient.startSession()
	if err != nil {
		panic(err)
	}

	runResult := m.Run()

	err = testClient.endSession()
	if err != nil {
		panic(err)
	}
	os.Exit(runResult)
}
