// Copyright (c) 2017-2019, Oracle and/or its affiliates. All rights reserved.

package ociauthz

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"oracle.com/oci/httpsigner"
)

type requestSigner struct {
	keySupplier httpsigner.KeySupplier
	algorithm   httpsigner.Algorithm
}

func TestNewClientValidSigner(t *testing.T) {
	var testSigner httpsigner.RequestSigner = &MockRequestSigner{}
	t.Run(
		`should return a client with non-nil signer`,
		func(t *testing.T) {
			c := NewSigningClient(testSigner, testKeyID)
			assert.NotNil(t, c)
			assert.Equal(t, testSigner, c.(*SigningClient).Signer())
			assert.Equal(t, testKeyID, c.(*SigningClient).KeyID())
		})
}

func TestNewClientNilSigner(t *testing.T) {
	t.Run(
		`should panic when passed a nil signer`,
		func(t *testing.T) {
			assert.Panics(t, func() {
				NewSigningClient(nil, "")
			})
		})
}

func TestNewCustomClient(t *testing.T) {
	var testSigner httpsigner.RequestSigner = &MockRequestSigner{}
	t.Run(
		`should return a client with non-nil signer and client`,
		func(t *testing.T) {
			options := &ClientOptions{HTTPClient: &http.Client{}}
			c := NewCustomSigningClient(testSigner, testKeyID, options)
			assert.NotNil(t, c)
			assert.Equal(t, testSigner, c.(*SigningClient).Signer())
			assert.Equal(t, testKeyID, c.(*SigningClient).KeyID())
		})
}

func TestNewCustomClientNilSigner(t *testing.T) {
	t.Run(
		`should panic when passed a nil signer`,
		func(t *testing.T) {
			options := &ClientOptions{}
			assert.Panics(t, func() {
				NewCustomSigningClient(nil, testKeyID, options)
			})
		})
}

func TestNewCustomClientNilClient(t *testing.T) {
	var testSigner httpsigner.RequestSigner = &MockRequestSigner{}
	t.Run(
		`should panic when passed a nil client`,
		func(t *testing.T) {
			assert.Panics(t, func() {
				options := &ClientOptions{}
				NewCustomSigningClient(testSigner, testKeyID, options)
			})
		})
}

func TestClientSignRequest(t *testing.T) {
	testIO := []struct {
		tc             string
		signingError   error
		keyRefreshed   bool
		reqMethod      string
		reqContentType string
		reqBody        []byte
		setGetBody     bool
	}{
		{tc: `should pass args through to signer and return non-error result`,
			signingError:   nil,
			keyRefreshed:   false,
			reqMethod:      "GET",
			reqContentType: "",
			reqBody:        nil,
			setGetBody:     false,
		},
		{tc: `should pass args through to signer and return non-error result for a request with nil GetBody`,
			signingError:   nil,
			keyRefreshed:   false,
			reqMethod:      "GET",
			reqContentType: "",
			reqBody:        nil,
			setGetBody:     false,
		},
		{tc: `should pass args through to signer and return error`,
			signingError:   errors.New("test error"),
			keyRefreshed:   false,
			reqMethod:      "POST",
			reqContentType: "",
			reqBody:        nil,
			setGetBody:     true,
		},
		{tc: `should have new keyid if KeyExpirationError is raised`,
			signingError:   httpsigner.NewKeyRotationError("new-key", testKeyID),
			keyRefreshed:   true,
			reqMethod:      "GET",
			reqContentType: "",
			reqBody:        nil,
			setGetBody:     false,
		},
		{tc: `should pass args through to signer and return non-error result with json body post request`,
			signingError:   nil,
			keyRefreshed:   false,
			reqMethod:      "POST",
			reqContentType: "",
			reqBody:        []byte(`{"title":"ociauthz"}`),
			setGetBody:     true,
		},
		{tc: `should pass args through to signer and return non-error result with json body put request`,
			signingError:   nil,
			keyRefreshed:   false,
			reqMethod:      "PUT",
			reqContentType: "",
			reqBody:        []byte(`{"title":"ociauthz"}`),
			setGetBody:     true,
		},
		{tc: `should pass args through to signer and return non-error result with json body patch request`,
			signingError:   nil,
			keyRefreshed:   false,
			reqMethod:      "PATCH",
			reqContentType: "",
			reqBody:        []byte(`{"title":"ociauthz"}`),
			setGetBody:     true,
		},
		{tc: `should pass args through to signer and return non-error result with json body post request and content-type header`,
			signingError:   nil,
			keyRefreshed:   false,
			reqMethod:      "POST",
			reqContentType: "application/json",
			reqBody:        []byte(`{"title":"ociauthz"}`),
			setGetBody:     true,
		},
		{tc: `should pass args through to signer and return non-error result with octet-stream body put and content-type header`,
			signingError:   nil,
			keyRefreshed:   false,
			reqMethod:      "PUT",
			reqContentType: "application/octet-stream",
			reqBody:        []byte(`{"title":"ociauthz"}`),
			setGetBody:     true,
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			// NOTE: we're using http.NewRequest rather than httptest.NewRequest here.  This is because http.request.GetBody does not get set
			// from httptest.NewRequest. SigningClient.SignRequest uses GetBody() to read the request body so it doesn't mutate the state
			// of the request.
			req, e := http.NewRequest(test.reqMethod, "http://localhost", bytes.NewBuffer(test.reqBody))
			if !test.setGetBody {
				req.GetBody = nil
			}

			if test.reqContentType != "" {
				req.Header.Set(HdrContentType, test.reqContentType)
			}

			assert.Nil(t, e)

			// setup
			ms := &MockRequestSigner{signingError: test.signingError}
			client := NewSigningClient(ms, testKeyID).(httpsigner.SigningClient)

			// go
			sreq, err := client.SignRequest(req)

			// verify results
			if test.signingError == nil {
				assert.Equal(t, req, sreq)
				assert.Nil(t, err)
			} else {
				assert.Nil(t, sreq)
				assert.Equal(t, test.signingError, err)
			}

			// verify mock
			assert.True(t, ms.SignRequestCalled)
			assert.Equal(t, req, ms.ProfferedRequest)

			if test.reqMethod == http.MethodPost || test.reqMethod == http.MethodPut || test.reqMethod == http.MethodPatch {
				headersToSign := make([]string, len(DefaultWriteHeadersToSign))
				copy(headersToSign, DefaultWriteHeadersToSign)
				assert.Equal(t, headersToSign, ms.ProfferedHeaders)
				assert.NotEqual(t, "", req.Header.Get(httpsigner.HdrDate))

				if test.reqContentType == "" {
					assert.Equal(t, "application/json", req.Header.Get(HdrContentType))
				} else {
					assert.Equal(t, test.reqContentType, req.Header.Get(HdrContentType))
				}

				assert.NotEqual(t, "", req.Header.Get(HdrContentLength))
				assert.NotEqual(t, "", req.Header.Get(HdrXContentSha256))
			} else {
				assert.Nil(t, req.GetBody)
				assert.Equal(t, httpsigner.DefaultHeadersToSign, ms.ProfferedHeaders)
				assert.NotEqual(t, "", req.Header.Get(httpsigner.HdrDate))
			}

			if test.keyRefreshed {
				assert.Equal(t, test.signingError.(*httpsigner.KeyRotationError).ReplacementKeyID, ms.ProfferedKey)
				assert.Equal(t, test.signingError.(*httpsigner.KeyRotationError).OldKeyID, testKeyID)
			} else {
				assert.Equal(t, testKeyID, ms.ProfferedKey)
			}

			// Verify request state is not mutated
			if test.reqBody != nil {
				reader, getBodyErr := req.GetBody()
				assert.Nil(t, getBodyErr)

				body, readErr := ioutil.ReadAll(reader)
				assert.Nil(t, readErr)

				assert.Equal(t, body, test.reqBody)
			}
		})
	}
}

func TestClientDo(t *testing.T) {
	testError := errors.New("test error")
	testIO := []struct {
		tc            string
		signer        *MockRequestSigner
		expectedError error
	}{
		{
			tc:            `should pass signed request to underlying client on valid request`,
			signer:        &MockRequestSigner{},
			expectedError: nil,
		},
		{
			tc:            `should return signing error without calling underlying client`,
			signer:        &MockRequestSigner{signingError: testError},
			expectedError: testError,
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			// setup
			client := NewSigningClient(test.signer, testKeyID)
			req := httptest.NewRequest("GET", "http://localhost", nil)

			// Go
			_, err := client.Do(req)

			// verify
			assert.True(t, test.signer.SignRequestCalled)
			if test.expectedError != nil {
				assert.Equal(t, test.expectedError, err)
			}
		})
	}
}

// Ensure SigningClient.Do calls given signRequestFn
func TestClientDoCallsOverridedRequestPrepare(t *testing.T) {
	signer := &MockRequestSigner{}
	testError := errors.New("error")
	prepareRequestFn := func(*http.Request) (*http.Request, []string, error) {
		return nil, nil, testError
	}
	defaultClient := httpsigner.NewClient(signer, testKeyID, &http.Client{}, prepareRequestFn)
	client := &SigningClient{DefaultClientOptions(), defaultClient.(*httpsigner.DefaultSigningClient)}
	req := httptest.NewRequest("GET", "http://localhost", nil)

	_, e := client.Do(req)

	assert.Equal(t, e, testError)
}

type mockReadCloser struct {
	readReturn int
	readError  error
	closeError error
}

func (m mockReadCloser) Read([]byte) (int, error) {
	return m.readReturn, m.readError
}

func (m mockReadCloser) Close() error {
	return m.closeError
}

func TestGetRequestBodySha256Errors(t *testing.T) {
	testError := errors.New("error")
	testIO := []struct {
		tc            string
		getBodyFn     func() (io.ReadCloser, error)
		expectedError error
	}{
		{
			tc:            `should return the test error if reqest.GetBody returns an error`,
			getBodyFn:     func() (io.ReadCloser, error) { return nil, testError },
			expectedError: testError,
		},
		{
			tc:            `should return the test error if ioutil.ReadAll returns an error`,
			getBodyFn:     func() (io.ReadCloser, error) { return mockReadCloser{readError: testError}, nil },
			expectedError: testError,
		},
	}
	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			// setup
			req := httptest.NewRequest("GET", "http://localhost", nil)
			req.GetBody = test.getBodyFn

			_, e := GetRequestBodySha256(req)
			assert.Equal(t, e, test.expectedError)
		})
	}
}

func TestGetRequestBodySha256UnknownContentLength(t *testing.T) {
	body := "Test Request Body"
	// Hide the body behind an io.Reader so the length is unknown
	bodyReader := ioutil.NopCloser(strings.NewReader(body))

	req := httptest.NewRequest("GET", "http://localhost", bodyReader)
	req.GetBody = func() (io.ReadCloser, error) {
		return ioutil.NopCloser(strings.NewReader(body)), nil
	}

	_, e := GetRequestBodySha256(req)
	assert.NoError(t, e)
}

// Test additional test cases for when we have POST/PUT method but no GetBody function on the request
func TestPrepareOCIRequestNoGetBody(t *testing.T) {
	testIO := []struct {
		tc            string
		method        string
		body          []byte
		expectGetBody bool
	}{
		{
			tc:            `should return signed request with create-update headers and GetBody attached to the request for a POST request with Body`,
			method:        "POST",
			body:          []byte(`body`),
			expectGetBody: true,
		},
		{
			tc:            `should return signed request with create-update headers and GetBody attached to the request for a PUT request with Body`,
			method:        "PUT",
			body:          []byte(`body`),
			expectGetBody: true,
		},
		{
			tc:            `should return signed request with create-update headers and GetBody attached to the request for a PATCH request with Body`,
			method:        "PATCH",
			body:          []byte(`body`),
			expectGetBody: true,
		},
		{
			tc:            `should return signed request with read headers and no GetBody attached to the request for a POST request with no Body`,
			method:        "POST",
			body:          nil,
			expectGetBody: false,
		},
		{
			tc:            `should return signed request with read headers and no GetBody attached to the request for a GET request with no Body`,
			method:        "GET",
			body:          nil,
			expectGetBody: false,
		},
	}
	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			// setup
			// Note here that httptest.NewRequest does not set GetBody
			req := httptest.NewRequest(test.method, "http://localhost", nil)

			// manually set Body to default it to nil
			req.Body = nil
			if test.body != nil {
				req.Body = ioutil.NopCloser(bytes.NewBuffer(test.body))
			}

			options := DefaultClientOptions()
			prepareOCIRequest := GeneratePrepareOCIRequestFn(options)

			// Go
			prepared, headers, err := prepareOCIRequest(req)

			// Check return values
			assert.Nil(t, err)

			// Check read operation headers exist
			assert.NotEqual(t, "", prepared.Header.Get(httpsigner.HdrDate))

			if test.expectGetBody {
				// Check write operation headers exist in the returned request
				assert.Equal(t, headers, DefaultWriteHeadersToSign)
				assert.NotNil(t, req.Body)
				assert.NotNil(t, req.GetBody)
				for _, hdr := range RequiredCreateUpdateHeaders {
					assert.NotEqual(t, "", prepared.Header.Get(hdr))
				}
			} else {
				assert.Equal(t, headers, httpsigner.DefaultHeadersToSign)
				assert.Nil(t, req.Body)
				assert.Nil(t, req.GetBody)
				for _, hdr := range RequiredCreateUpdateHeaders {
					assert.Equal(t, "", prepared.Header.Get(hdr))
				}
			}
		})
	}
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// Test prepareOCIRequest when GetRequestBodySha256 returns an error
func TestPrepareOCIRequestGetReqestBodySha256Err(t *testing.T) {
	testError := errors.New("error")
	testIO := []struct {
		tc     string
		method string
	}{
		{tc: `should handle error when GetRequestBodySha256 returns an error of a POST request`,
			method: "POST"},
		{tc: `should handle error when GetRequestBodySha256 returns an error of a PUT request`,
			method: "PUT"},
		{tc: `should handle error when GetRequestBodySha256 returns an error of a PATCH request`,
			method: "PATCH"},
	}
	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			// setup
			req := httptest.NewRequest(test.method, "http://localhost", nil)

			// Triggering GetBody error will simulate error from GetRequestBodySha256
			req.GetBody = func() (io.ReadCloser, error) { return mockReadCloser{readError: testError}, nil }

			options := DefaultClientOptions()
			prepareOCIRequest := GeneratePrepareOCIRequestFn(options)

			// Go
			prepared, headers, err := prepareOCIRequest(req)

			// Check return values
			assert.Equal(t, err, testError)
			assert.Nil(t, prepared)
			assert.Nil(t, headers)
		})
	}
}

func TestPrepareOCIRequestHeadersToSign(t *testing.T) {
	testIO := []struct {
		tc              string
		method          string
		clientOptions   *ClientOptions
		expectedHeaders []string
	}{
		{tc: `should result in empty headersToSign given empty read/write headers to sign in client options`,
			method: "POST", clientOptions: &ClientOptions{&http.Client{}, []string{}, []string{}}, expectedHeaders: []string{}},
		{tc: `should result in headersToSign with items from read headers to sign in client options for get requests`,
			method: "GET", clientOptions: &ClientOptions{&http.Client{}, []string{"a"}, []string{"b"}}, expectedHeaders: []string{"a"}},
		{tc: `should result in headersToSign with  write headers to sign in client options for post requests`,
			method: "POST", clientOptions: &ClientOptions{&http.Client{}, []string{"a"}, []string{"b"}}, expectedHeaders: []string{"b"}},
	}
	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			var req *http.Request
			if methodCreateOrUpdate(test.method) {
				var e error
				body := []byte(`{}`)
				req, e = http.NewRequest(test.method, "http://localhost", bytes.NewBuffer(body))
				assert.Nil(t, e)
			} else {
				var e error
				req, e = http.NewRequest(test.method, "http://localhost", nil)
				assert.Nil(t, e)
			}

			prepareRequestFn := GeneratePrepareOCIRequestFn(test.clientOptions)
			prepared, headers, err := prepareRequestFn(req)

			// Check return values
			assert.NotNil(t, prepared)
			assert.Nil(t, err, nil)
			sort.Strings(headers)
			sort.Strings(test.expectedHeaders)
			assert.Equal(t, headers, test.expectedHeaders)
		})
	}
}

func TestCheckRequiredHeaders(t *testing.T) {
	testIO := []struct {
		tc            string
		method        string
		headers       []string
		expectedError error
	}{

		// Missing required headers
		{tc: `should return ErrRequiredHeaderMissing if GET request has no headers`,
			method: "GET", headers: []string{}, expectedError: ErrRequiredHeaderMissing},
		{tc: `should return ErrRequiredHeaderMissing if POST request has no headers`,
			method: "POST", headers: []string{}, expectedError: ErrRequiredHeaderMissing},
		{tc: `should return ErrRequiredHeaderMissing if PUT request has no headers`,
			method: "PUT", headers: []string{}, expectedError: ErrRequiredHeaderMissing},
		{tc: `should return ErrRequiredHeaderMissing if PATCH request has no headers`,
			method: "PATCH", headers: []string{}, expectedError: ErrRequiredHeaderMissing},

		// Missing required headers
		{tc: `should return ErrRequiredHeaderMissing if GET request is missing required headers`,
			method: "GET", headers: []string{httpsigner.HdrRequestTarget}, expectedError: ErrRequiredHeaderMissing},
		{tc: `should return ErrRequiredHeaderMissing if POST request is missing required headers`,
			method: "POST", headers: []string{httpsigner.HdrDate, httpsigner.HdrRequestTarget}, expectedError: ErrRequiredHeaderMissing},
		{tc: `should return ErrRequiredHeaderMissing if PUT request is missing required headers`,
			method: "PUT", headers: []string{httpsigner.HdrDate, httpsigner.HdrRequestTarget}, expectedError: ErrRequiredHeaderMissing},
		{tc: `should return ErrRequiredHeaderMissing if PATCH request is missing required headers`,
			method: "PATCH", headers: []string{httpsigner.HdrDate, httpsigner.HdrRequestTarget}, expectedError: ErrRequiredHeaderMissing},

		{tc: `should return ErrRequiredHeaderMissing if POST request is missing required CreateUpdate headers`,
			method: "POST", headers: []string{httpsigner.HdrDate, httpsigner.HdrRequestTarget}, expectedError: ErrRequiredHeaderMissing},
		{tc: `should return ErrRequiredHeaderMissing if PUT request is missing required CreateUpdate headers`,
			method: "PUT", headers: []string{httpsigner.HdrDate, httpsigner.HdrRequestTarget}, expectedError: ErrRequiredHeaderMissing},
		{tc: `should return ErrRequiredHeaderMissing if PATCH request is missing required CreateUpdate headers`,
			method: "PATCH", headers: []string{httpsigner.HdrDate, httpsigner.HdrRequestTarget}, expectedError: ErrRequiredHeaderMissing},

		// Success
		{tc: `should return no error if GET request contains required headers`,
			method: "GET", headers: []string{httpsigner.HdrDate, httpsigner.HdrRequestTarget}, expectedError: nil},
		{tc: `should return no error if POST request contains required headers`,
			method: "POST", headers: []string{httpsigner.HdrDate, httpsigner.HdrRequestTarget, HdrContentType, HdrContentLength, HdrXContentSha256}, expectedError: nil},
		{tc: `should return no error if PUT request contains required headers`,
			method: "PUT", headers: []string{httpsigner.HdrDate, httpsigner.HdrRequestTarget, HdrContentType, HdrContentLength, HdrXContentSha256}, expectedError: nil},
		{tc: `should return no error if PATCH request contains required headers`,
			method: "PATCH", headers: []string{httpsigner.HdrDate, httpsigner.HdrRequestTarget, HdrContentType, HdrContentLength, HdrXContentSha256}, expectedError: nil},

		// Success with x-date instead of date
		{tc: `should return no error if GET request contains required headers and uses x-date instead of date`,
			method: "GET", headers: []string{HdrXDate, httpsigner.HdrRequestTarget}, expectedError: nil},
		{tc: `should return no error if POST request contains required headers and uses x-date instead of date`,
			method: "POST", headers: []string{HdrXDate, httpsigner.HdrRequestTarget, HdrContentType, HdrContentLength, HdrXContentSha256}, expectedError: nil},
		{tc: `should return no error if PUT request contains required headers and uses x-date instead of date`,
			method: "PUT", headers: []string{HdrXDate, httpsigner.HdrRequestTarget, HdrContentType, HdrContentLength, HdrXContentSha256}, expectedError: nil},
		{tc: `should return no error if PATCH request contains required headers and uses x-date instead of date`,
			method: "PATCH", headers: []string{HdrXDate, httpsigner.HdrRequestTarget, HdrContentType, HdrContentLength, HdrXContentSha256}, expectedError: nil},

		// Success with extra headers
		{tc: `should return no error if GET request contains required headers + extra`,
			method: "GET", headers: []string{httpsigner.HdrDate, httpsigner.HdrRequestTarget, "accept-type"}, expectedError: nil},
		{tc: `should return no error if POST request contains required headers + extra`,
			method: "POST", headers: []string{httpsigner.HdrDate, httpsigner.HdrRequestTarget, HdrContentType, HdrContentLength, HdrXContentSha256, "accept-type"}, expectedError: nil},
		{tc: `should return no error if PUT request contains required headers + extra`,
			method: "PUT", headers: []string{httpsigner.HdrDate, httpsigner.HdrRequestTarget, HdrContentType, HdrContentLength, HdrXContentSha256, "accept-type"}, expectedError: nil},
		{tc: `should return no error if PATCH request contains required headers + extra`,
			method: "PATCH", headers: []string{httpsigner.HdrDate, httpsigner.HdrRequestTarget, HdrContentType, HdrContentLength, HdrXContentSha256, "accept-type"}, expectedError: nil},
	}
	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			// setup
			req := httptest.NewRequest(test.method, "http://localhost", nil)
			for _, header := range test.headers {
				if header != httpsigner.HdrRequestTarget {
					req.Header.Set(header, "value")
				}
			}

			e := CheckRequiredHeaders(req)
			assert.Equal(t, test.expectedError, e)
		})
	}
}

func TestCheckRequiredHeadersArray(t *testing.T) {
	oboHdr := `opc-obo-token`
	testIO := []struct {
		tc        string
		method    string
		headers   []string
		extraHdrs []string
		expErr    error
	}{
		// No headers
		{tc: `should return ErrRequiredHeaderMissing if GET request has no headers`,
			method: "GET", headers: []string{}, extraHdrs: []string{}, expErr: ErrRequiredHeaderMissing},
		{tc: `should return ErrRequiredHeaderMissing if POST request has no headers`,
			method: "POST", headers: []string{}, extraHdrs: []string{}, expErr: ErrRequiredHeaderMissing},
		{tc: `should return ErrRequiredHeaderMissing if PUT request has no headers`,
			method: "PUT", headers: []string{}, extraHdrs: []string{}, expErr: ErrRequiredHeaderMissing},
		{tc: `should return ErrRequiredHeaderMissing if PATCH request has no headers`,
			method: "PATCH", headers: []string{}, extraHdrs: []string{}, expErr: ErrRequiredHeaderMissing},

		// Missing required headers
		{tc: `should return ErrRequiredHeaderMissing if GET request is missing required headers`,
			method: "GET", headers: []string{httpsigner.HdrDate}, extraHdrs: []string{}, expErr: ErrRequiredHeaderMissing},
		{tc: `should return ErrRequiredHeaderMissing if POST request is missing required headers`,
			method: "POST", headers: []string{httpsigner.HdrDate}, extraHdrs: []string{}, expErr: ErrRequiredHeaderMissing},
		{tc: `should return ErrRequiredHeaderMissing if PUT request is missing required headers`,
			method: "PUT", headers: []string{httpsigner.HdrDate}, extraHdrs: []string{}, expErr: ErrRequiredHeaderMissing},
		{tc: `should return ErrRequiredHeaderMissing if PATCH request is missing required headers`,
			method: "PATCH", headers: []string{httpsigner.HdrDate}, extraHdrs: []string{}, expErr: ErrRequiredHeaderMissing},

		// Missing CreateUpdate headers
		{tc: `should return ErrRequiredHeaderMissing if POST request is missing CreateUpdate headers`,
			method: "POST", headers: []string{httpsigner.HdrDate, httpsigner.HdrRequestTarget}, extraHdrs: []string{}, expErr: ErrRequiredHeaderMissing},
		{tc: `should return ErrRequiredHeaderMissing if PUT request is missing CreateUpdate headers`,
			method: "PUT", headers: []string{httpsigner.HdrDate, httpsigner.HdrRequestTarget}, extraHdrs: []string{}, expErr: ErrRequiredHeaderMissing},
		{tc: `should return ErrRequiredHeaderMissing if PATCH request is missing CreateUpdate headers`,
			method: "PATCH", headers: []string{httpsigner.HdrDate, httpsigner.HdrRequestTarget}, extraHdrs: []string{}, expErr: ErrRequiredHeaderMissing},

		// Success
		{tc: `should return no error if GET request contains required headers`,
			method: "GET", headers: []string{httpsigner.HdrDate, httpsigner.HdrRequestTarget}, extraHdrs: []string{}, expErr: nil},
		{tc: `should return no error if POST request contains required headers`,
			method: "POST", headers: []string{httpsigner.HdrDate, httpsigner.HdrRequestTarget, HdrContentType, HdrContentLength, HdrXContentSha256}, extraHdrs: []string{}, expErr: nil},
		{tc: `should return no error if PUT request contains required headers`,
			method: "PUT", headers: []string{httpsigner.HdrDate, httpsigner.HdrRequestTarget, HdrContentType, HdrContentLength, HdrXContentSha256}, extraHdrs: []string{}, expErr: nil},
		{tc: `should return no error if PATCH request contains required headers`,
			method: "PATCH", headers: []string{httpsigner.HdrDate, httpsigner.HdrRequestTarget, HdrContentType, HdrContentLength, HdrXContentSha256}, extraHdrs: []string{}, expErr: nil},

		// Success with x-date instead of date
		{tc: `should return no error if GET request contains required headers and uses x-date instead of date`,
			method: "GET", headers: []string{HdrXDate, httpsigner.HdrRequestTarget}, extraHdrs: []string{}, expErr: nil},
		{tc: `should return no error if POST request contains required headers and uses x-date instead of date`,
			method: "POST", headers: []string{HdrXDate, httpsigner.HdrRequestTarget, HdrContentType, HdrContentLength, HdrXContentSha256}, extraHdrs: []string{}, expErr: nil},
		{tc: `should return no error if PUT request contains required headers and uses x-date instead of date`,
			method: "PUT", headers: []string{HdrXDate, httpsigner.HdrRequestTarget, HdrContentType, HdrContentLength, HdrXContentSha256}, extraHdrs: []string{}, expErr: nil},
		{tc: `should return no error if PATCH request contains required headers and uses x-date instead of date`,
			method: "PATCH", headers: []string{HdrXDate, httpsigner.HdrRequestTarget, HdrContentType, HdrContentLength, HdrXContentSha256}, extraHdrs: []string{}, expErr: nil},

		// Success with extra headers
		{tc: `should return no error if GET request contains required headers + extra`,
			method: "GET", headers: []string{httpsigner.HdrDate, httpsigner.HdrRequestTarget, "accept-type"}, extraHdrs: []string{}, expErr: nil},
		{tc: `should return no error if POST request contains required headers + extra`,
			method: "POST", headers: []string{httpsigner.HdrDate, httpsigner.HdrRequestTarget, HdrContentType, HdrContentLength, HdrXContentSha256, "accept-type"}, extraHdrs: []string{}, expErr: nil},
		{tc: `should return no error if PUT request contains required headers + extra`,
			method: "PUT", headers: []string{httpsigner.HdrDate, httpsigner.HdrRequestTarget, HdrContentType, HdrContentLength, HdrXContentSha256, "accept-type"}, extraHdrs: []string{}, expErr: nil},
		{tc: `should return no error if PATCH request contains required headers + extra`,
			method: "PATCH", headers: []string{httpsigner.HdrDate, httpsigner.HdrRequestTarget, HdrContentType, HdrContentLength, HdrXContentSha256, "accept-type"}, extraHdrs: []string{}, expErr: nil},

		// Extra required headers tests
		{tc: `should return ErrRequiredHeaderMissing when request is missing extra required headers`,
			method: "GET", headers: []string{httpsigner.HdrDate, httpsigner.HdrRequestTarget}, extraHdrs: []string{oboHdr}, expErr: ErrRequiredHeaderMissing},
		{tc: `should return no error when request is contains extra required headers`,
			method: "GET", headers: []string{httpsigner.HdrDate, httpsigner.HdrRequestTarget, oboHdr}, extraHdrs: []string{oboHdr}, expErr: nil},
	}
	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			e := CheckRequiredHeadersArray(test.method, test.headers, test.extraHdrs)
			if test.expErr == nil {
				assert.Nil(t, e)
			} else {
				assert.NotNil(t, e)
				assert.Equal(t, test.expErr, e)
			}
		})
	}
}

func Test_IsDefinedTagsOnlyPutRequest(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		body    string
		want    bool
		wantErr bool
	}{
		{
			name:   "returns true for a PUT request body with defined tags only",
			method: "PUT",
			body:   `{"definedTags": {"ns1": {"foo": "bar", "k2": "123"}, "ns2": {"baz": "bat"}}}`,
			want:   true,
		},
		{

			name:   "returns false for a request body containing an array",
			method: "PUT",
			body:   `[{"definedTags": {"ns1": {"foo": "bar", "k2": "123"}, "ns2": {"baz": "bat"}}}]`,
		},
		{
			name:   "returns false for a request body with freeform tags",
			method: "PUT",
			body:   `{"freeformTags": {"foo": "bar"}}`,
		},
		{
			name:   "returns false for a request body with freeform and defined tags",
			method: "PUT",
			body:   `{"freeformTags": {"k": "v"}, "definedTags": {"ns": {"key1": "val1"}}}`,
		},
		{
			name:   "returns false for a request body with definedTags and any other data",
			method: "PUT",
			body:   `{"definedTags": {"some-ns": {"foo": "bar"}}, "some key": "some value"}`,
		},
		{
			name:   "returns false for a POST request body with defined tags only",
			method: "POST",
			body:   `{"definedTags": {"ns1": {"foo": "bar", "k2": "123"}, "ns2": {"baz": "bat"}}}`,
		},
		{
			name:   "returns false for a PATCH request body with defined tags only",
			method: "PATCH",
			body:   `{"definedTags": {"ns1": {"foo": "bar", "k2": "123"}, "ns2": {"baz": "bat"}}}`,
		},
		{
			name:   "returns false for a GET request body with defined tags only",
			method: "GET",
			body:   `{"definedTags": {"ns1": {"foo": "bar", "k2": "123"}, "ns2": {"baz": "bat"}}}`,
		},
		{
			name:   "returns false for a DELETE request body with defined tags only",
			method: "DELETE",
			body:   `{"definedTags": {"ns1": {"foo": "bar", "k2": "123"}, "ns2": {"baz": "bat"}}}`,
		},
		{
			name:    "returns an error for a request body where definedTags is not a map",
			method:  "PUT",
			body:    `{"definedTags": "foo"}`,
			wantErr: true,
		},
		{
			name:    "returns an error for an invalid request body (missing curly braces)",
			method:  "PUT",
			body:    `"freeformTags": {"foo": "bar"}`,
			wantErr: true,
		},
		{
			name:    "returns an error for a request body with definedTags in an invalid format (missing namespace)",
			method:  "PUT",
			body:    `{"definedTags": {"foo": "bar"}}`,
			wantErr: true,
		},
		{
			name:    "returns an error for a request body with definedTags in an invalid format (too much nesting)",
			method:  "PUT",
			body:    `{"definedTags": {"ns": {"k1": {"too": "nested"}}}}`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request, err := http.NewRequest(tt.method, "www.example.com", ioutil.NopCloser(strings.NewReader(tt.body)))
			assert.NoError(t, err)

			got, err := IsDefinedTagsOnlyPutRequest(request)
			if tt.wantErr {
				assert.Error(t, err)
				assert.False(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			// Verify request body can be read multiple times and doesn't throw EOF errors.
			rawBody, err := ioutil.ReadAll(request.Body)
			assert.NoError(t, err)
			assert.NotEmpty(t, rawBody)
		})
	}
}
