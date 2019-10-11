// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package shared

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

const (
	prettyString = "==========================================================="
	ERR          = 1
)

func PrintTestHeader(funcUnderTest string) {
	fmt.Println("===", funcUnderTest, "===")
	fmt.Println(prettyString)
}

func PrintTestFooter() {
	fmt.Println(prettyString)
}

func PrintResults(v interface{}) {
	fmt.Printf("%+v\n", v)
	PrintTestFooter()
}

var TestVals = map[string]string{}

func init() {

	if err := godotenv.Load(); err != nil {
		fmt.Println(err)
		os.Exit(ERR)
	}

	keys := []string{
		"BAREMETAL_PRIVATE_KEY_PATH",
		"BAREMETAL_TENANCY_OCID",
		"BAREMETAL_USER_OCID",
		"BAREMETAL_FINGERPRINT",
		"BAREMETAL_KEY_PASSWORD",
		"TEST_COMPARTMENT_ID",
	}

	fmt.Println("=== TEST ENVIRONMENT ===")

	fmt.Println(prettyString)

	for _, key := range keys {
		val := os.Getenv(key)
		fmt.Println(key, "=>", val)
		TestVals[key] = val
	}

	fmt.Println(prettyString)

}
