// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

// +build recording

package helpers

import (
	"fmt"
	"os"

	"github.com/dnaeon/go-vcr/recorder"
	"github.com/joho/godotenv"

	bm "github.com/oracle/bmcs-go-sdk"
)

const RUNMODE = RunmodeRecord

const (
	prettyString = "==========================================================="
)

var apiParams = map[string]string{}

func init() {
	if err := godotenv.Load(); err != nil {
		panic("could not load required environment - " + err.Error())
	}

	keys := []string{
		"BAREMETAL_PRIVATE_KEY_PATH",
		"BAREMETAL_TENANCY_OCID",
		"BAREMETAL_USER_OCID",
		"BAREMETAL_FINGERPRINT",
		"BAREMETAL_KEY_PASSWORD",
		"TEST_COMPARTMENT_ID",
	}
	fmt.Println(prettyString)

	for _, key := range keys {
		val := os.Getenv(key)
		fmt.Println(key, "=>", val)
		apiParams[key] = val
	}

	fmt.Println(prettyString)
}

func GetClient(cassetteName string) *TestClient {
	rec, err := recorder.NewAsMode(cassetteName, recorder.ModeRecording, NewTestTransport(nil))

	if err != nil {
		panic(fmt.Sprintf("could not create recorder. error: %s", err))
	}

	client, err := bm.NewClient(
		apiParams["BAREMETAL_USER_OCID"],
		apiParams["BAREMETAL_TENANCY_OCID"],
		apiParams["BAREMETAL_FINGERPRINT"],
		bm.CustomTransport(rec),
		bm.PrivateKeyFilePath(apiParams["BAREMETAL_PRIVATE_KEY_PATH"]),
		bm.PrivateKeyPassword(apiParams["BAREMETAL_KEY_PASSWORD"]),
	)

	if err != nil {
		panic(fmt.Sprintf("could not create client. error: %s", err))
	}

	tc := &TestClient{
		Client:   client,
		recorder: rec,
	}

	return tc

}
