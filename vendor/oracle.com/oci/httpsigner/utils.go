// Copyright (c) 2018, Oracle and/or its affiliates. All rights reserved.

package httpsigner

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
)

// SetRequestGetBody sets GetBody function on the given http.Request if it is missing.  GetBody allows
// reading from the request body without mutating the state of the request.
func SetRequestGetBody(r *http.Request) error {
	if r.Body == nil || r.GetBody != nil {
		return nil
	}

	buffer, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	original := ioutil.NopCloser(bytes.NewBuffer(buffer))
	r.Body = original

	// GetBody does not mutate the state of the request body
	snapshot := *bytes.NewBuffer(buffer)
	r.GetBody = func() (io.ReadCloser, error) {
		r := snapshot
		return ioutil.NopCloser(&r), nil
	}

	return nil
}
