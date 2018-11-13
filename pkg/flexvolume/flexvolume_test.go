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
	"bytes"
	"testing"

	"go.uber.org/zap/zaptest"
)

const defaultTestOps = `{"kubernetes.io/fsType":"ext4","kubernetes.io/readwrite":"rw"}`

func TestInit(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	bak := out
	out = new(bytes.Buffer)
	defer func() { out = bak }()

	code := 0
	osexit := exit
	exit = func(c int) { code = c }
	defer func() { exit = osexit }()

	ExecDriver(logger, mockFlexvolumeDriver{}, []string{"oci", "init"})

	if out.(*bytes.Buffer).String() != `{"status":"Success"}`+"\n" {
		t.Fatalf(`Expected '{"status":"Success"}'; got %q`, out.(*bytes.Buffer).String())
	}

	if code != 0 {
		t.Fatalf("Expected 'exit 0'; got 'exit %d'", code)
	}
}

// TestVolumeName tests that the getvolumename call-out results in
// StatusNotSupported as the call-out is broken as of the latest stable Kube
// release (1.6.4).
func TestGetVolumeName(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	bak := out
	out = new(bytes.Buffer)
	defer func() { out = bak }()

	code := 0
	osexit := exit
	exit = func(c int) { code = c }
	defer func() { exit = osexit }()

	ExecDriver(logger, mockFlexvolumeDriver{}, []string{"oci", "getvolumename", defaultTestOps})

	if out.(*bytes.Buffer).String() != `{"status":"Not supported","message":"getvolumename is broken as of kube 1.6.4"}`+"\n" {
		t.Fatalf(`Expected '{"status":"Not supported","message":"getvolumename is broken as of kube 1.6.4"}}'; got %s`, out.(*bytes.Buffer).String())
	}

	if code != 0 {
		t.Fatalf("Expected 'exit 0'; got 'exit %d'", code)
	}
}

func TestAttachUnsuported(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	bak := out
	out = new(bytes.Buffer)
	defer func() { out = bak }()

	code := 0
	osexit := exit
	exit = func(c int) { code = c }
	defer func() { exit = osexit }()

	ExecDriver(logger, mockFlexvolumeDriver{}, []string{"oci", "attach", defaultTestOps, "nodeName"})

	if out.(*bytes.Buffer).String() != `{"status":"Not supported"}`+"\n" {
		t.Fatalf(`Expected '{"status":"Not supported""}'; got %s`, out.(*bytes.Buffer).String())
	}

	if code != 0 {
		t.Fatalf("Expected 'exit 0'; got 'exit %d'", code)
	}
}
