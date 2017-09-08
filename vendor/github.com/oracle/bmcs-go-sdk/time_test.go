// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTimeJSONMarshalling(t *testing.T) {
	buffer := []byte(`"2016-08-12T19:20:48Z"`)

	var tyme Time
	err := json.Unmarshal(buffer, &tyme)

	assert.Nil(t, err)

	actual, err := tyme.MarshalJSON()

	assert.Nil(t, err)
	assert.Equal(t, string(buffer), string(actual))

}
