// Copyright 2017 Oracle and/or its affiliates. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package flexvolume

import (
	"encoding/base64"
	"testing"
)

const (
	testSecretOption = OptionKeySecret + "/test"
)

func makeTestOpts() Options {
	return Options{
		testSecretOption: base64.StdEncoding.EncodeToString([]byte("hello")),
		OptionFSType:     "ext4",
	}
}

func TestDecodeKubeSecretsDecodesSecret(t *testing.T) {

	opts, err := DecodeKubeSecrets(makeTestOpts())

	if err != nil {
		t.Fatalf("Got unexpected error %s", err)
	}

	if opts[testSecretOption] != "hello" {
		t.Fatalf("Expected 'hello'; got '%s'", opts[testSecretOption])
	}
}

func TestDecodeKubeSecretsDoesntEffectNonSecrets(t *testing.T) {
	opts, err := DecodeKubeSecrets(makeTestOpts())

	if err != nil {
		t.Fatalf("Got unexpected error %s", err)
	}

	if opts[OptionFSType] != "ext4" {
		t.Fatalf("Expected 'ext4'; got '%s'", opts[OptionFSType])
	}
}
