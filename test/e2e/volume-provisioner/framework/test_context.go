// Copyright 2018 Oracle and/or its affiliates. All rights reserved.
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

package framework

import (
	"flag"
)

// TestContextType represents the framework flags
type TestContextType struct {
	// RepoRoot is the root directory of the repository.
	RepoRoot string

	// KubeConfig is the path to the kubeconfig file.
	KubeConfig string

	// OCIConfig is the path to the ociconfig file
	OCIConfig string

	// AD used to specify an availability domain to create the volumes in.
	AD string

	// MntTargetOCID used mount a volume to the specific mount id
	MntTargetOCID string

	// Image is the docker image to which we are building
	Image string

	// Namespace (if provided) is the namespace of an existing namespace to
	// use for test execution rather than creating a new namespace.
	Namespace string
	// DeleteNamespace controls whether or not to delete test namespaces
	DeleteNamespace bool
	// DeleteNamespaceOnFailure controls whether or not to delete test
	// namespaces when the test fails.
	DeleteNamespaceOnFailure bool
}

// TestContext holds the context of the the test run.
var TestContext TestContextType

// RegisterFlags registers the test framework flags and populates TestContext.
func RegisterFlags() {
	flag.StringVar(&TestContext.RepoRoot, "repo-root", "../../", "Root directory of kubernetes repository, for finding test files.")
	flag.StringVar(&TestContext.KubeConfig, "kubeconfig", "", "Path to Kubeconfig file with authorization and master location information.")
	flag.StringVar(&TestContext.OCIConfig, "ociconfig", "", "Path to OCIconfig file with cloud provider config.")
	flag.StringVar(&TestContext.AD, "ad", "AD-2", "The availability domain is specified to identify in which to create the volumes.")
	flag.StringVar(&TestContext.MntTargetOCID, "mnt-target-id", "", "Mount Target ID is specified to identify the mount target to be attached to the volumes")
	flag.StringVar(&TestContext.Image, "image", "", "Image name and version to build images")
	flag.StringVar(&TestContext.Namespace, "namespace", "", "Name of an existing Namespace to run tests in.")
	flag.BoolVar(&TestContext.DeleteNamespace, "delete-namespace", true, "If true tests will delete namespace after completion. It is only designed to make debugging easier, DO NOT turn it off by default.")
	flag.BoolVar(&TestContext.DeleteNamespaceOnFailure, "delete-namespace-on-failure", true, "If true tests will delete their associated namespace upon completion whether or not the test has failed.")
}

func init() {
	RegisterFlags()
}
