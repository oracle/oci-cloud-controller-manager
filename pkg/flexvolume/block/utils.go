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

package block

import (
	"os"
	"time"
)

const waitForPathDelay = 1 * time.Second

// waitForPathToExist waits for for a given filesystem path to exist.
func waitForPathToExist(path string, maxRetries int) bool {
	// TODO: Replace with "k8s.io/apimachinery/pkg/util/wait".
	for i := 0; i < maxRetries; i++ {
		var err error
		_, err = os.Stat(path)
		if err == nil {
			return true
		}
		if err != nil && !os.IsNotExist(err) {
			return false
		}
		if i == maxRetries-1 {
			break
		}
		time.Sleep(waitForPathDelay)
	}
	return false
}
