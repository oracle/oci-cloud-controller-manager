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

package metadata

import "errors"

type mockMetadataGetter struct {
	metadata *InstanceMetadata
}

type mockMetadataErrorGetter struct{}

// NewMock returns a new mock OCI instance metadata getter.
func NewMock(metadata *InstanceMetadata) Interface {
	return &mockMetadataGetter{metadata: metadata}
}

// NewErrorMock returns a new mock OCI instance metadata getter that returns an error on Get().
func NewErrorMock() Interface {
	return &mockMetadataErrorGetter{}
}

func (m *mockMetadataGetter) Get() (*InstanceMetadata, error) {
	return m.metadata, nil
}

func (m *mockMetadataErrorGetter) Get() (*InstanceMetadata, error) {
	return nil, errors.New("uh oh")
}
