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
	"fmt"
	"strings"
)

var decodeKubeSecret = base64.StdEncoding.DecodeString

// DecodeKubeSecrets takes the options passed to the driver and decodes any
// secrets.
func DecodeKubeSecrets(opts Options) (Options, error) {
	for k, opt := range opts {
		if strings.HasPrefix(k, OptionKeySecret) {
			secret, err := decodeKubeSecret(opt)
			if err != nil {
				return nil, fmt.Errorf("unable to decode secret %q: %v", k, err)
			}
			opts[k] = string(secret)
		}
	}
	return opts, nil
}
