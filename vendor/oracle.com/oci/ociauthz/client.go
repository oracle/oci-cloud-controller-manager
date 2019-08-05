// Copyright (c) 2017-2019, Oracle and/or its affiliates. All rights reserved.

package ociauthz

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"oracle.com/oci/tagging"

	"oracle.com/oci/httpsigner"
)

// Constants for headers we're going to sign for POST requests
const (
	HdrXContentSha256 = "x-content-sha256"
	HdrContentLength  = "Content-Length"
	HdrContentType    = "Content-Type"
	HdrXDate          = "x-date"
)

// Default list of headers in the request the signing client will sign.
var (
	DefaultReadHeadersToSign  = []string{httpsigner.HdrDate, httpsigner.HdrRequestTarget}
	DefaultWriteHeadersToSign = []string{httpsigner.HdrDate, httpsigner.HdrRequestTarget, HdrXContentSha256, HdrContentLength, HdrContentType}
)

// These variables describe required headers CheckRequiredHeaders uses for checking
var (
	RequiredHeaders             = []string{httpsigner.HdrRequestTarget}
	RequiredCreateUpdateHeaders = []string{HdrContentType, HdrXContentSha256, HdrContentLength}
)

// ClientOptions is used for configuring ociauthz.SigningClient
type ClientOptions struct {
	HTTPClient         *http.Client
	ReadHeadersToSign  []string
	WriteHeadersToSign []string
}

// DefaultClientOptions returns default client options to use for the OCI signing client.
func DefaultClientOptions() *ClientOptions {
	return &ClientOptions{
		HTTPClient:         &http.Client{Timeout: defaultTimeout},
		ReadHeadersToSign:  DefaultReadHeadersToSign,
		WriteHeadersToSign: DefaultWriteHeadersToSign,
	}
}

// SigningClient is an implementation of httpsigner.SigningClient.  Given a KeySupplier, it will handle signing of OCI requests.
type SigningClient struct {
	ClientOptions *ClientOptions
	*httpsigner.DefaultSigningClient
}

// NewSigningClient returns a new instance of SigningClient. A nil signer will result in a panic.
func NewSigningClient(signer httpsigner.RequestSigner, keyID string) httpsigner.Client {
	defaultClientOptions := DefaultClientOptions()
	prepareOCIRequestFn := GeneratePrepareOCIRequestFn(defaultClientOptions)
	defaultClient := httpsigner.NewClient(signer, keyID, defaultClientOptions.HTTPClient, prepareOCIRequestFn)
	return &SigningClient{defaultClientOptions, defaultClient.(*httpsigner.DefaultSigningClient)}
}

// NewCustomSigningClient returns a new instance of SigningClient with the provided http.client. A nil signer or http.client will
// result in a panic.
func NewCustomSigningClient(signer httpsigner.RequestSigner, keyID string, options *ClientOptions) httpsigner.Client {
	prepareOCIRequestFn := GeneratePrepareOCIRequestFn(options)
	client := httpsigner.NewClient(signer, keyID, options.HTTPClient, prepareOCIRequestFn)
	return &SigningClient{options, client.(*httpsigner.DefaultSigningClient)}
}

// GeneratePrepareOCIRequestFn creates a function which modifies the given request to add required headers and return a list of
// headers the default client should sign. For server side request object that is missing request.GetBody will have one added
// for write requests.
func GeneratePrepareOCIRequestFn(options *ClientOptions) httpsigner.PrepareRequestFn {
	prepareOCIRequest := func(request *http.Request) (*http.Request, []string, error) {
		var headersToSign []string
		request.Header.Set(httpsigner.HdrDate, time.Now().UTC().Format(http.TimeFormat))

		// Set GetBody to the request if it does not have one set.  Note that if the request does not have Body
		// for some reason, GetBody could still be nil.
		if methodCreateOrUpdate(request.Method) && request.GetBody == nil {
			if err := httpsigner.SetRequestGetBody(request); err != nil {
				return nil, nil, err
			}
		}

		// POST, PUT, and PATCH require additional headers
		if methodCreateOrUpdate(request.Method) && request.GetBody != nil {
			// Get the checksum of the request body
			body, err := GetRequestBodySha256(request)
			if err != nil {
				return nil, nil, err
			}

			// set headers
			request.Header.Set(HdrXContentSha256, body)
			request.Header.Set(HdrContentLength, strconv.FormatInt(request.ContentLength, 10))

			// default Content-Type to application/json
			if c := request.Header.Get(HdrContentType); c == "" {
				request.Header.Set(HdrContentType, "application/json")
			}

			headersToSign = make([]string, len(options.WriteHeadersToSign))
			copy(headersToSign, options.WriteHeadersToSign)

		} else {
			headersToSign = make([]string, len(options.ReadHeadersToSign))
			copy(headersToSign, options.ReadHeadersToSign)
		}

		return request, headersToSign, nil
	}
	return prepareOCIRequest
}

// GetRequestBodySha256 will return the base64 encoded SHA256 of the request body
func GetRequestBodySha256(request *http.Request) (body string, err error) {
	// We use GetBody() here instead of the Body member directly so that we preserve the original state of the request Body.
	// Body buffer will be exhausted once it is read.
	reader, err := request.GetBody()
	if err != nil {
		return
	}

	rawBody, err := ioutil.ReadAll(reader)
	if err != nil {
		return
	}

	hash := sha256.Sum256(rawBody)
	body = base64.StdEncoding.EncodeToString(hash[:])
	return
}

// CheckRequiredHeaders will respond with an error if the given request does not contain the required headers
func CheckRequiredHeaders(request *http.Request) error {
	// Special case - we allow date or x-date header, so only fail if both are missing
	date := request.Header.Get(httpsigner.HdrDate)
	xDate := request.Header.Get(HdrXDate)
	if date == "" && xDate == "" {
		return ErrRequiredHeaderMissing
	}

	// Headers for read operations except (request-target)
	for _, required := range RequiredHeaders {
		if h := request.Header.Get(required); h == "" && required != httpsigner.HdrRequestTarget {
			return ErrRequiredHeaderMissing
		}
	}

	// Headers for write operations
	if methodCreateOrUpdate(request.Method) {
		for _, required := range RequiredCreateUpdateHeaders {
			if h := request.Header.Get(required); h == "" {
				return ErrRequiredHeaderMissing
			}
		}
	}

	return nil
}

// CheckRequiredHeadersArray will respond with an error if the given array of headers does not contain the required headers.
// different headers are checked based on the method argument which should correspond to a http verb.  The caller may
// pass any additional headers to require in extraRequiredHeaders.
func CheckRequiredHeadersArray(method string, headers, extraRequiredHeaders []string) error {
	// make a map of the headers
	hdrMap := make(map[string]string, len(headers))
	for _, hdr := range headers {
		hdr = strings.ToLower(hdr)
		hdrMap[hdr] = hdr
	}

	// check date and x-date
	_, ok := hdrMap[`date`]
	if !ok {
		_, ok = hdrMap[`x-date`]
		if !ok {
			return ErrRequiredHeaderMissing
		}
	}

	// get headers
	for _, hdr := range RequiredHeaders {
		hdr = strings.ToLower(hdr)
		if _, ok := hdrMap[hdr]; !ok {
			return ErrRequiredHeaderMissing
		}
	}

	// put/post/patch headers
	if methodCreateOrUpdate(method) {
		for _, hdr := range RequiredCreateUpdateHeaders {
			hdr = strings.ToLower(hdr)
			if _, ok := hdrMap[hdr]; !ok {
				return ErrRequiredHeaderMissing
			}
		}
	}

	// caller provided requirements
	for _, hdr := range extraRequiredHeaders {
		hdr = strings.ToLower(hdr)
		if _, ok := hdrMap[hdr]; !ok {
			return ErrRequiredHeaderMissing
		}
	}

	return nil
}

// methodCreateOrUpdate returns true if the given http method is one of POST, PUT, or PATCH
func methodCreateOrUpdate(method string) bool {
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
		return true
	}
	return false
}

// DefinedTagsKey is the name of the json field which contains defined tag data.
const DefinedTagsKey = "definedTags"

// IsDefinedTagsOnlyPutRequest reports if an HTTP request body contains only defined tags. This is used to support "delete tag key" functionality (PLEX-301).
func IsDefinedTagsOnlyPutRequest(request *http.Request) (bool, error) {
	if request == nil {
		return false, ErrInvalidArg
	}

	// Only PUT is supported
	if request.Method != http.MethodPut {
		return false, nil
	}

	// Set request.GetBody() if not set so the body can be read multiple times, thereby avoiding an EOF error.
	if request.GetBody == nil {
		err := httpsigner.SetRequestGetBody(request)
		if err != nil {
			return false, err
		}
	}

	reader, err := request.GetBody()
	if err != nil {
		return false, err
	}

	rawBody, err := ioutil.ReadAll(reader)
	if err != nil {
		return false, err
	}

	// Unmarshal json to verify there is only a single key
	var data interface{}
	err = json.Unmarshal(rawBody, &data)
	if err != nil {
		return false, err
	}

	m, ok := data.(map[string]interface{})
	if !ok {
		return false, nil
	}

	if len(m) != 1 {
		return false, nil
	}

	// The key must be for defined tags only
	tags, exists := m[DefinedTagsKey]
	if !exists {
		return false, nil
	}

	tagJSON, err := json.Marshal(tags)
	if err != nil {
		return false, err
	}

	// Verify the tag data conforms to a defined tag set
	definedTags := tagging.DefinedTagSet{}
	err = json.Unmarshal(tagJSON, &definedTags)
	if err != nil {
		return false, err
	}

	return true, nil
}
