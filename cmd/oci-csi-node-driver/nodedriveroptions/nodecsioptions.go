// Copyright 2019 Oracle and/or its affiliates. All rights reserved.
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

package nodedriveroptions

//NodeCSIOptions contains details about the flag
type NodeCSIOptions struct {
	Endpoint   string // Used for Block Volume CSI driver
	NodeID     string
	LogLevel   string
	Master     string
	Kubeconfig string

	EnableFssDriver            bool
	FssEndpoint                string
}

type NodeOptions struct {
	Name                   string
	Endpoint               string
	NodeID                 string
	Kubeconfig             string
	Master                 string
	DriverName             string
	DriverVersion          string
	EnableControllerServer bool
}
