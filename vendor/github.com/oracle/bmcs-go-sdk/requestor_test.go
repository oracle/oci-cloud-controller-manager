// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"math/rand"
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

var numRetries = 0

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

func buildTestURL(urlTemplate string, region string, resource resourceName, query url.Values, ids ...interface{}) (string, error) {
	return "url!", nil
}

func newTestAPIRequestor(authInfo *authenticationInfo, nco *NewClientOptions) (r *apiRequestor) {
	return &apiRequestor{
		httpClient: &http.Client{
			Transport: nco.Transport,
		},
		authInfo:           authInfo,
		urlBuilder:         buildTestURL,
		userAgent:          nco.UserAgent,
		region:             nco.Region,
		shortRetryTime:     nco.ShortRetryTime,
		longRetryTime:      nco.LongRetryTime,
		randGen:            nco.RandGen,
		disableAutoRetries: nco.DisableAutoRetries,
	}
}

func TestRunRequestorTests(t *testing.T) {
	sleep = mockSleep
	suite.Run(t, new(RequestorTestSuite))
	sleep = time.Sleep
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

	api := newTestAPIRequestor(s.client.authInfo, &NewClientOptions{
		Transport: s.transport,
		UserAgent: userAgent,
		RandGen:   rand.New(rand.NewSource(time.Now().UnixNano())),
	})

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

	api := newTestAPIRequestor(s.client.authInfo, &NewClientOptions{
		Transport: s.transport,
		RandGen:   rand.New(rand.NewSource(time.Now().UnixNano())),
	})

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

	api := newTestAPIRequestor(s.client.authInfo, &NewClientOptions{
		Transport: s.transport,
		RandGen:   rand.New(rand.NewSource(time.Now().UnixNano())),
	})

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

	api := newTestAPIRequestor(s.client.authInfo, &NewClientOptions{
		Transport: s.transport,
		RandGen:   rand.New(rand.NewSource(time.Now().UnixNano())),
	})

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

	api := newTestAPIRequestor(s.client.authInfo, &NewClientOptions{
		Transport: s.transport,
		RandGen:   rand.New(rand.NewSource(time.Now().UnixNano())),
	})

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

	api := newTestAPIRequestor(s.client.authInfo, &NewClientOptions{
		Transport: s.transport,
		RandGen:   rand.New(rand.NewSource(time.Now().UnixNano())),
	})

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

	api := newTestAPIRequestor(s.client.authInfo, &NewClientOptions{
		Transport: s.transport,
		RandGen:   rand.New(rand.NewSource(time.Now().UnixNano())),
	})

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

	api := newTestAPIRequestor(s.client.authInfo, &NewClientOptions{
		Transport: s.transport,
		RandGen:   rand.New(rand.NewSource(time.Now().UnixNano())),
	})

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

func testGetMaxRetryTimeForStatus(s *RequestorTestSuite, status string, codes []string, noRetryServices []string, defaultRetryServices []string, longRetryServices []string, methods []string) {
	var testUrl string
	var waitTime time.Duration
	var service string
	var err = Error{}
	api := newTestAPIRequestor(s.client.authInfo, &NewClientOptions{
		Transport:      s.transport,
		ShortRetryTime: shortRetryTime,
		LongRetryTime:  longRetryTime,
		RandGen:        rand.New(rand.NewSource(time.Now().UnixNano())),
	})

	err.Status = status
	for _, code := range codes {
		err.Code = code
		for _, method := range methods {
			for _, service = range noRetryServices {
				testUrl = fmt.Sprintf(baseUrlTemplate, service, us_phoenix_1)
				waitTime = getMaxRetryTimeInSeconds(api, err, testUrl, method, false)
				s.Equal(waitTime, time.Duration(0))
			}
			for _, service = range defaultRetryServices {
				testUrl = fmt.Sprintf(baseUrlTemplate, service, us_phoenix_1)
				waitTime = getMaxRetryTimeInSeconds(api, err, testUrl, method, false)
				s.Equal(waitTime, api.shortRetryTime)
			}
			for _, service = range longRetryServices {
				testUrl = fmt.Sprintf(baseUrlTemplate, service, us_phoenix_1)
				waitTime = getMaxRetryTimeInSeconds(api, err, testUrl, method, false)
				s.Equal(waitTime, api.longRetryTime)
			}
		}
	}
}

func (s *RequestorTestSuite) TestGetMaxRetryTimeInSeconds() {
	var status string
	var codes []string
	var noRetryServices []string
	var defaultRetryServices []string
	var longRetryServices []string
	var methods []string

	allMethods := []string{http.MethodDelete, http.MethodGet, http.MethodPost, http.MethodPut}
	allServices := []string{"AnyService", identityServiceAPI, coreServiceAPI, databaseServiceAPI, objectStorageServiceAPI}

	status = "400"
	codes = []string{"AnyCode", "CannotParseRequest", "InvalidParameter", "LimitExceeded", "MissingParameter", "QuotaExceeded"}
	noRetryServices = allServices
	defaultRetryServices = []string{}
	longRetryServices = []string{}
	testGetMaxRetryTimeForStatus(s, status, codes, noRetryServices, defaultRetryServices, longRetryServices, allMethods)

	status = "401"
	codes = []string{"AnyCode", "NotAuthenticated"}
	noRetryServices = allServices
	defaultRetryServices = []string{}
	longRetryServices = []string{}
	testGetMaxRetryTimeForStatus(s, status, codes, noRetryServices, defaultRetryServices, longRetryServices, allMethods)

	status = "403"
	codes = []string{"AnyCode"}
	noRetryServices = allServices
	defaultRetryServices = []string{}
	longRetryServices = []string{}
	testGetMaxRetryTimeForStatus(s, status, codes, noRetryServices, defaultRetryServices, longRetryServices, allMethods)

	status = "404"
	codes = []string{"AnyCode", "NotAuthorizedOrNotFound"}
	noRetryServices = []string{}
	defaultRetryServices = []string{"AnyService", coreServiceAPI, databaseServiceAPI}
	longRetryServices = []string{identityServiceAPI, objectStorageServiceAPI}
	methods = []string{http.MethodGet, http.MethodPost, http.MethodPut}
	testGetMaxRetryTimeForStatus(s, status, codes, noRetryServices, defaultRetryServices, longRetryServices, methods)

	noRetryServices = allServices
	defaultRetryServices = []string{}
	longRetryServices = []string{}
	methods = []string{http.MethodDelete}
	testGetMaxRetryTimeForStatus(s, status, codes, noRetryServices, defaultRetryServices, longRetryServices, methods)

	status = "409"
	codes = []string{"InvalidatedRetryToken", "CompartmentAlreadyExists"}
	noRetryServices = allServices
	defaultRetryServices = []string{}
	longRetryServices = []string{}
	testGetMaxRetryTimeForStatus(s, status, codes, noRetryServices, defaultRetryServices, longRetryServices, allMethods)

	codes = []string{"NotAuthorizedOrResourceAlreadyExists"}
	noRetryServices = []string{}
	defaultRetryServices = []string{"AnyService", coreServiceAPI, databaseServiceAPI}
	longRetryServices = []string{identityServiceAPI, objectStorageServiceAPI}
	testGetMaxRetryTimeForStatus(s, status, codes, noRetryServices, defaultRetryServices, longRetryServices, allMethods)

	codes = []string{"AnyCode", "IncorrectState"}
	noRetryServices = []string{}
	defaultRetryServices = allServices
	longRetryServices = []string{}
	testGetMaxRetryTimeForStatus(s, status, codes, noRetryServices, defaultRetryServices, longRetryServices, allMethods)

	status = "412"
	codes = []string{"AnyCode", "NoEtagMatch"}
	noRetryServices = allServices
	defaultRetryServices = []string{}
	longRetryServices = []string{}
	testGetMaxRetryTimeForStatus(s, status, codes, noRetryServices, defaultRetryServices, longRetryServices, allMethods)

	status = "429"
	codes = []string{"AnyCode", "TooManyRequests"}
	noRetryServices = []string{}
	defaultRetryServices = []string{}
	longRetryServices = allServices
	testGetMaxRetryTimeForStatus(s, status, codes, noRetryServices, defaultRetryServices, longRetryServices, allMethods)

	status = "500"
	codes = []string{"AnyCode", "InternalServerError"}
	noRetryServices = []string{}
	defaultRetryServices = []string{"AnyService", coreServiceAPI, databaseServiceAPI, identityServiceAPI}
	longRetryServices = []string{objectStorageServiceAPI}
	testGetMaxRetryTimeForStatus(s, status, codes, noRetryServices, defaultRetryServices, longRetryServices, allMethods)

	status = "AnyStatus"
	codes = []string{"AnyCode"}
	noRetryServices = []string{}
	defaultRetryServices = allServices
	longRetryServices = []string{}
	testGetMaxRetryTimeForStatus(s, status, codes, noRetryServices, defaultRetryServices, longRetryServices, allMethods)
}

func (s *RequestorTestSuite) TestRequestServiceCheck() {
	var testUrl string
	testUrl = fmt.Sprintf(baseUrlTemplate, objectStorageServiceAPI, us_phoenix_1)
	s.True(requestServiceCheck(testUrl, objectStorageServiceAPI))
	s.False(requestServiceCheck(testUrl, identityServiceAPI))
	s.False(requestServiceCheck(testUrl, ""))
	testUrl = fmt.Sprintf("https://something.%s.%s.oraclebmc.com", objectStorageServiceAPI, us_phoenix_1)
	s.False(requestServiceCheck(testUrl, objectStorageServiceAPI))
}

func mockSleep(_ time.Duration) {
	numRetries++
}

func (s *RequestorTestSuite) TestPolynomialBackoffSleep() {
	var secondsSlept time.Duration
	secondsSlept = polynomialBackoffSleep(1, 0)
	s.Equal(secondsSlept, time.Duration(0))
	secondsSlept = polynomialBackoffSleep(1, time.Duration(4)*time.Second)
	s.Equal(secondsSlept, time.Duration(1)*time.Second)
	secondsSlept = polynomialBackoffSleep(2, time.Duration(4)*time.Second)
	s.Equal(secondsSlept, time.Duration(4)*time.Second)
}

func (s *RequestorTestSuite) TestNotRetriableRequest() {
	numRetries = 0
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

	api := newTestAPIRequestor(s.client.authInfo, &NewClientOptions{
		Transport:      s.transport,
		ShortRetryTime: shortRetryTime,
		LongRetryTime:  longRetryTime,
		RandGen:        rand.New(rand.NewSource(time.Now().UnixNano())),
	})

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
	s.Equal(numRetries, 0)
}

func (s *RequestorTestSuite) TestRetriableRequestTimeout() {
	numRetries = 0
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("opc-request-id", "1234567890")
		// anything other than 200 is an error
		w.WriteHeader(http.StatusNotFound)
		err := Error{
			Message: "foo",
			Code:    "bar",
		}

		buff, _ := json.Marshal(err)

		w.Write(buff)
		return

	}))

	api := newTestAPIRequestor(s.client.authInfo, &NewClientOptions{
		Transport:      s.transport,
		ShortRetryTime: shortRetryTime,
		LongRetryTime:  longRetryTime,
		RandGen:        rand.New(rand.NewSource(time.Now().UnixNano())),
	})

	body := struct{}{}
	var marshaled []byte
	marshaled, _ = json.Marshal(body)

	reqOpts := &mockRequestOptions{}
	reqOpts.On("marshalBody").Return(marshaled)
	reqOpts.On("marshalURL", mock.MatchedBy(matchBuilderFns(buildTestURL))).Return(ts.URL)
	reqOpts.On("marshalHeader").Return(http.Header{})

	_, e := api.request(http.MethodPost, reqOpts)

	s.NotNil(e)
	s.Equal(e.Error(), "Status: 404; Code: bar; OPC Request ID: 1234567890; Message: foo")

	expectedRetryNum := 0
	timeWaited := 0
	for timeWaited < int(shortRetryTime.Seconds()) {
		expectedRetryNum++
		timeWaited += expectedRetryNum * expectedRetryNum
	}
	s.Equal(numRetries, expectedRetryNum)
}

func (s *RequestorTestSuite) TestRetriableRequestTimeoutWithChangeInStatus() {
	numRetries = 0
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if numRetries <= 3 {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("opc-request-id", "1234567890")
			// anything other than 200 is an error
			w.WriteHeader(http.StatusInternalServerError)
			err := Error{
				Message: "foo",
				Code:    "bar",
			}

			buff, _ := json.Marshal(err)

			w.Write(buff)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("opc-request-id", "1234567890")
			// anything other than 200 is an error
			w.WriteHeader(http.StatusTooManyRequests)
			err := Error{
				Message: "foo",
				Code:    "bar",
			}

			buff, _ := json.Marshal(err)

			w.Write(buff)
		}
		return

	}))

	api := newTestAPIRequestor(s.client.authInfo, &NewClientOptions{
		Transport:      s.transport,
		ShortRetryTime: shortRetryTime,
		LongRetryTime:  longRetryTime,
		RandGen:        rand.New(rand.NewSource(time.Now().UnixNano())),
	})

	body := struct{}{}
	var marshaled []byte
	marshaled, _ = json.Marshal(body)

	reqOpts := &mockRequestOptions{}
	reqOpts.On("marshalBody").Return(marshaled)
	reqOpts.On("marshalURL", mock.MatchedBy(matchBuilderFns(buildTestURL))).Return(ts.URL)
	reqOpts.On("marshalHeader").Return(http.Header{})

	_, e := api.request(http.MethodPost, reqOpts)

	s.NotNil(e)
	s.Equal(e.Error(), "Status: 429; Code: bar; OPC Request ID: 1234567890; Message: foo")

	expectedRetryNum := 0
	timeWaited := 0
	for timeWaited < int(longRetryTime.Seconds()) {
		expectedRetryNum++
		timeWaited += expectedRetryNum * expectedRetryNum
	}
	s.Equal(numRetries, expectedRetryNum)
}

func (s *RequestorTestSuite) TestRetriableRequestWithChangeInStatusWithSuccess() {
	numRetries = 0
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if numRetries <= 3 {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("opc-request-id", "1234567890")
			// anything other than 200 is an error
			w.WriteHeader(http.StatusInternalServerError)
			err := Error{
				Message: "foo",
				Code:    "bar",
			}

			buff, _ := json.Marshal(err)

			w.Write(buff)
		} else if numRetries <= 5 {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("opc-request-id", "1234567890")
			// anything other than 200 is an error
			w.WriteHeader(http.StatusTooManyRequests)
			err := Error{
				Message: "foo",
				Code:    "bar",
			}

			buff, _ := json.Marshal(err)

			w.Write(buff)
		} else {
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
		}
		return

	}))

	api := newTestAPIRequestor(s.client.authInfo, &NewClientOptions{
		Transport:      s.transport,
		ShortRetryTime: shortRetryTime,
		LongRetryTime:  longRetryTime,
		RandGen:        rand.New(rand.NewSource(time.Now().UnixNano())),
	})

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
	s.Equal(numRetries, 6)
}

func (s *RequestorTestSuite) TestRetriableRequestWithSuccess() {
	numRetries = 0
	retryNumForSuccess := 3
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if numRetries == retryNumForSuccess {
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
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("opc-request-id", "1234567890")
			// anything other than 200 is an error
			w.WriteHeader(http.StatusNotFound)
			err := Error{
				Message: "foo",
				Code:    "bar",
			}

			buff, _ := json.Marshal(err)

			w.Write(buff)
		}
		return
	}))

	api := newTestAPIRequestor(s.client.authInfo, &NewClientOptions{
		Transport:      s.transport,
		ShortRetryTime: shortRetryTime,
		LongRetryTime:  longRetryTime,
		RandGen:        rand.New(rand.NewSource(time.Now().UnixNano())),
	})

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
	s.Equal(numRetries, retryNumForSuccess)
}

func (s *RequestorTestSuite) TestAutomaticallyAddingRetryToken() {
	numRetries = 0
	retryNumForSuccess := 3
	var retryToken string
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if numRetries == retryNumForSuccess {
			var details identityCreationRequirement
			decoder := json.NewDecoder(r.Body)

			decoder.Decode(&details)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			u := User{
				ID:            "0123456789",
				CompartmentID: details.CompartmentID,
				Name:          r.Header.Get(retryTokenKey),
				Description:   details.Description,
				TimeCreated:   time.Now(),
				State:         ResourceCreated,
			}

			buff, _ := json.Marshal(u)

			w.Write(buff)
		} else {
			retryToken = r.Header.Get(retryTokenKey)
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("opc-request-id", "1234567890")
			// anything other than 200 is an error
			w.WriteHeader(http.StatusNotFound)
			err := Error{
				Message: "foo",
				Code:    "bar",
			}

			buff, _ := json.Marshal(err)

			w.Write(buff)
		}
		return
	}))

	api := newTestAPIRequestor(s.client.authInfo, &NewClientOptions{
		Transport:      s.transport,
		ShortRetryTime: shortRetryTime,
		LongRetryTime:  longRetryTime,
		RandGen:        rand.New(rand.NewSource(time.Now().UnixNano())),
	})

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
	s.NotEqual(retryToken, "")
	s.Equal(retryToken, user.Name)
	s.Equal(numRetries, retryNumForSuccess)
}

func (s *RequestorTestSuite) TestKeepingProvidedRetryToken() {
	numRetries = 0
	retryNumForSuccess := 3
	retryToken := "abcdef"
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if numRetries == retryNumForSuccess {
			var details identityCreationRequirement
			decoder := json.NewDecoder(r.Body)

			decoder.Decode(&details)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			u := User{
				ID:            "0123456789",
				CompartmentID: details.CompartmentID,
				Name:          r.Header.Get(retryTokenKey),
				Description:   details.Description,
				TimeCreated:   time.Now(),
				State:         ResourceCreated,
			}

			buff, _ := json.Marshal(u)

			w.Write(buff)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("opc-request-id", "1234567890")
			// anything other than 200 is an error
			w.WriteHeader(http.StatusNotFound)
			err := Error{
				Message: "foo",
				Code:    "bar",
			}

			buff, _ := json.Marshal(err)

			w.Write(buff)
		}
		return
	}))

	api := newTestAPIRequestor(s.client.authInfo, &NewClientOptions{
		Transport:      s.transport,
		ShortRetryTime: shortRetryTime,
		LongRetryTime:  longRetryTime,
		RandGen:        rand.New(rand.NewSource(time.Now().UnixNano())),
	})

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
	header := http.Header{}
	header.Set(retryTokenKey, retryToken)
	reqOpts.On("marshalHeader").Return(header)

	response, e := api.request(http.MethodPost, reqOpts)

	if !s.Nil(e) {
		s.T().Log(e.Error())
	}

	var user User
	e = json.Unmarshal(response.body, &user)

	s.Nil(e)
	s.Equal(retryToken, user.Name)
	s.Equal(numRetries, retryNumForSuccess)
}

func (s *RequestorTestSuite) TestDisablingAutoRetries() {
	numRetries = 0
	retryNumForSuccess := 3
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if numRetries == retryNumForSuccess {
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
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("opc-request-id", "1234567890")
			// anything other than 200 is an error
			w.WriteHeader(http.StatusNotFound)
			err := Error{
				Message: r.Header.Get(retryTokenKey),
				Code:    "bar",
			}

			buff, _ := json.Marshal(err)

			w.Write(buff)
		}
		return
	}))

	api := newTestAPIRequestor(s.client.authInfo, &NewClientOptions{
		Transport:          s.transport,
		ShortRetryTime:     shortRetryTime,
		LongRetryTime:      longRetryTime,
		RandGen:            rand.New(rand.NewSource(time.Now().UnixNano())),
		DisableAutoRetries: true,
	})

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

	_, e := api.request(http.MethodPost, reqOpts)

	s.Equal(numRetries, 0)
	s.NotNil(e)
	s.Equal(e.Error(), "Status: 404; Code: bar; OPC Request ID: 1234567890; Message: ")
}
