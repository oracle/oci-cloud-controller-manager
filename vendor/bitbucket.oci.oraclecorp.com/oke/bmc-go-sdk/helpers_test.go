// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"bytes"
	"encoding/json"
)

func marshalObjectForTest(obj interface{}) []byte {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.Encode(obj)

	return buffer.Bytes()
}
