// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type RequestorTestSuite struct {
	suite.Suite
	client    *Client
	transport *http.Transport
}

func (s *RequestorTestSuite) SetupTest() {
	var e error
	s.client, e = getTestClient()
	if e != nil {
		panic(e)
	}

	s.transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
}

func buildTestURL(urlTemplate string, region string, resource resourceName, query url.Values, ids ...interface{}) string {
	return "url!"
}

func newTestAPIRequestor(authInfo *authenticationInfo, nco *NewClientOptions) (r *apiRequestor) {
	return &apiRequestor{
		httpClient: &http.Client{
			Transport: nco.Transport,
		},
		authInfo:   authInfo,
		urlBuilder: buildTestURL,
		userAgent:  nco.UserAgent,
		region:     nco.Region,
	}
}

func TestRunRequestorTests(t *testing.T) {
	suite.Run(t, new(RequestorTestSuite))
}

func (s *RequestorTestSuite) TestDeleteRequest() {
	userAgent := "test-user-agent"
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.Require().Equal(r.Header.Get("User-Agent"), userAgent)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		w.Write([]byte("Deleted group 1"))
		return

	}))

	api := newTestAPIRequestor(s.client.authInfo, &NewClientOptions{Transport: s.transport, UserAgent: userAgent})

	url, _ := url.Parse(fmt.Sprintf("%s/%s/%s", ts.URL, resourceGroups, "123"))

	reqOpts := &mockRequestOptions{}
	reqOpts.On("marshalURL", mock.MatchedBy(matchBuilderFns(buildTestURL))).Return(url.String())
	reqOpts.On("marshalHeader").Return(http.Header{})

	e := api.deleteRequest(reqOpts)
	s.Nil(e)
}

func (s *RequestorTestSuite) TestUnsuccessfulDeleteRequest() {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("opc-request-id", "1234567890")

		w.WriteHeader(http.StatusNotFound)

		error := Error{
			Code:    "42",
			Message: "ultimate answer",
		}

		var buffer bytes.Buffer
		encoder := json.NewEncoder(&buffer)
		encoder.Encode(error)

		w.Write(buffer.Bytes())

		return

	}))

	api := newTestAPIRequestor(s.client.authInfo, &NewClientOptions{Transport: s.transport})

	url, _ := url.Parse(fmt.Sprintf("%s/%s/%s", ts.URL, resourceGroups, "123"))

	reqOpts := &mockRequestOptions{}
	reqOpts.On("marshalURL", mock.MatchedBy(matchBuilderFns(buildTestURL))).Return(url.String())
	reqOpts.On("marshalHeader").Return(http.Header{})

	e := api.deleteRequest(reqOpts)
	s.NotNil(e)
	s.Equal("Status: 404; Code: 42; OPC Request ID: 1234567890; Message: ultimate answer", e.Error())
}

func (s *RequestorTestSuite) TestGetRequest() {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		u := []User{
			{
				ID:            "0123456789",
				CompartmentID: r.URL.Query().Get("queryCompartmentID"),
				Name:          "Bob",
				Description:   "Bob's name",
				TimeCreated:   time.Now(),
				State:         ResourceCreated,
			},
		}

		buff, _ := json.Marshal(u)

		w.Write(buff)
		return

	}))

	api := newTestAPIRequestor(s.client.authInfo, &NewClientOptions{Transport: s.transport})

	url := ts.URL + "/users?" + fmt.Sprintf("%s=%s", "queryCompartmentID", s.client.authInfo.tenancyOCID)

	reqOpts := &mockRequestOptions{}
	reqOpts.On("marshalURL", mock.MatchedBy(matchBuilderFns(buildTestURL))).Return(url)
	reqOpts.On("marshalHeader").Return(http.Header{})

	resp, e := api.getRequest(reqOpts)
	s.Nil(e)

	buffer := bytes.NewBuffer(resp.body)
	decoder := json.NewDecoder(buffer)
	var users []User
	e = decoder.Decode(&users)
	s.Nil(e)

	s.Equal(len(users), 1)
	s.Equal(users[0].ID, "0123456789")

}

func (s *RequestorTestSuite) TestUnsuccessfulGetRequest() {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("opc-request-id", "1234567890")

		w.WriteHeader(http.StatusForbidden)
		err := Error{
			Message: "foo",
			Code:    "bar",
		}

		buff, _ := json.Marshal(err)

		w.Write(buff)
		return

	}))

	api := newTestAPIRequestor(s.client.authInfo, &NewClientOptions{Transport: s.transport})

	url := ts.URL + "/users?compartmentId=1"

	reqOpts := &mockRequestOptions{}
	reqOpts.On("marshalURL", mock.MatchedBy(matchBuilderFns(buildTestURL))).Return(url)
	reqOpts.On("marshalHeader").Return(http.Header{})

	resp, e := api.getRequest(reqOpts)
	s.NotNil(e)
	s.Nil(resp)
	s.Equal(e.Error(), "Status: 403; Code: bar; OPC Request ID: 1234567890; Message: foo")

}

func (s *RequestorTestSuite) TestUnsuccessfulRequest() {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("opc-request-id", "1234567890")
		// anything other than 200 is an error
		w.WriteHeader(http.StatusBadRequest)
		err := Error{
			Message: "foo",
			Code:    "bar",
		}

		buff, _ := json.Marshal(err)

		w.Write(buff)
		return

	}))

	api := newTestAPIRequestor(s.client.authInfo, &NewClientOptions{Transport: s.transport})

	body := struct{}{}
	var marshaled []byte
	marshaled, _ = json.Marshal(body)

	reqOpts := &mockRequestOptions{}
	reqOpts.On("marshalBody").Return(marshaled)
	reqOpts.On("marshalURL", mock.MatchedBy(matchBuilderFns(buildTestURL))).Return(ts.URL)
	reqOpts.On("marshalHeader").Return(http.Header{})

	_, e := api.request(http.MethodPost, reqOpts)

	s.NotNil(e)
	s.Equal(e.Error(), "Status: 400; Code: bar; OPC Request ID: 1234567890; Message: foo")
}

func (s *RequestorTestSuite) TestPlainTextErrorResponse() {
	const errorMsg string = "<html><head></head><body>Internal Server Error</body>"
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errorMsg))
		return
	}))

	api := newTestAPIRequestor(s.client.authInfo, &NewClientOptions{Transport: s.transport})

	body := struct{}{}
	var marshaled []byte
	marshaled, _ = json.Marshal(body)

	reqOpts := &mockRequestOptions{}
	reqOpts.On("marshalBody").Return(marshaled)
	reqOpts.On("marshalURL", mock.MatchedBy(matchBuilderFns(buildTestURL))).Return(ts.URL)
	reqOpts.On("marshalHeader").Return(http.Header{})

	_, e := api.request(http.MethodPost, reqOpts)

	s.NotNil(e)
	s.Equal(e.Error(), "Status: 500; Code: ; OPC Request ID: ; Message: "+errorMsg)
}

func (s *RequestorTestSuite) TestBadJsonErrorResponse() {
	const errorMsg string = "Invalid JSON body"
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errorMsg))
		return
	}))

	api := newTestAPIRequestor(s.client.authInfo, &NewClientOptions{Transport: s.transport})

	body := struct{}{}
	var marshaled []byte
	marshaled, _ = json.Marshal(body)

	reqOpts := &mockRequestOptions{}
	reqOpts.On("marshalBody").Return(marshaled)
	reqOpts.On("marshalURL", mock.MatchedBy(matchBuilderFns(buildTestURL))).Return(ts.URL)
	reqOpts.On("marshalHeader").Return(http.Header{})

	_, e := api.request(http.MethodPost, reqOpts)

	s.NotNil(e)
	s.Equal(e.Error(), "Status: 400; Code: ; OPC Request ID: ; Message: "+errorMsg)
}

func (s *RequestorTestSuite) TestSuccessfulRequest() {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var details identityCreationRequirement
		decoder := json.NewDecoder(r.Body)

		decoder.Decode(&details)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		u := User{
			ID:            "0123456789",
			CompartmentID: details.CompartmentID,
			Name:          details.Name,
			Description:   details.Description,
			TimeCreated:   time.Now(),
			State:         ResourceCreated,
		}

		buff, _ := json.Marshal(u)

		w.Write(buff)
		return

	}))

	api := newTestAPIRequestor(s.client.authInfo, &NewClientOptions{Transport: s.transport})

	body := identityCreationRequirement{
		CompartmentID: "xyz",
		Description:   "123abc",
		Name:          "Bob",
	}
	var marshaled []byte
	marshaled, _ = json.Marshal(body)

	reqOpts := &mockRequestOptions{}
	reqOpts.On("marshalBody").Return(marshaled)
	reqOpts.On("marshalURL", mock.MatchedBy(matchBuilderFns(buildTestURL))).Return(ts.URL)
	reqOpts.On("marshalHeader").Return(http.Header{})

	response, e := api.request(http.MethodPost, reqOpts)

	if !s.Nil(e) {
		s.T().Log(e.Error())
	}

	var user User
	e = json.Unmarshal(response.body, &user)

	s.Nil(e)
	s.Equal(body.CompartmentID, user.CompartmentID)
	s.Equal(body.Name, user.Name)
	s.Equal(body.Description, user.Description)
}
