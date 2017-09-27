// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package helpers

import (
	"errors"
	"log"
	"math/rand"
	"net/http"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/dnaeon/go-vcr/recorder"

	bm "github.com/oracle/bmcs-go-sdk"
)

type Runmode string

const (
	RunmodeRecord Runmode = "RECORD"
	RunmodeReplay Runmode = "REPLAY"
)

// This test used to dynamically select the image and shape from the available options, but some images and sizes are /much/ slower to create than others. These are known to be better than the worst case, but feel free to optimize more!
const (
	OracleLinux73ImageID = "ocid1.image.oc1.phx.aaaaaaaa6uwtn7h3hogd5zlwd35eeqbndurkayshzvrfx5usqn6cwxd5vdqq" // DisplayName: Oracle-Linux-7.3-2017.05.23-0
	FastestImageID       = OracleLinux73ImageID
	SmallestShapeName    = "VM.Standard1.1"
)

type stopper interface {
	Stop() error
}

type TestClient struct {
	*bm.Client
	recorder *recorder.Recorder
}

func (tc *TestClient) Stop() error {
	return tc.recorder.Stop()
}

type testTransport struct {
	requestCount  int
	realTransport http.RoundTripper
	mtx           sync.Mutex
}

func NewTestTransport(t http.RoundTripper) http.RoundTripper {
	tt := &testTransport{
		realTransport: http.DefaultTransport,
	}
	if t != nil {
		tt.realTransport = t
	}
	return tt
}

func (tt *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Request-Number", strconv.Itoa(tt.requestCount))
	tt.mtx.Lock()
	tt.requestCount++
	tt.mtx.Unlock()
	return tt.realTransport.RoundTrip(req)
}

var validNameChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789.-_"

func RandomText(keySize int) string {
	key := make([]byte, keySize)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Read(key)
	for i := range key {
		key[i] = validNameChars[key[i]%byte(len(validNameChars))]
	}
	return string(key)
}

func BoolPtr(v bool) *bool {
	return &v
}

const maxWaitDuration = 120 * time.Minute

type resourceCommandResult struct {
	id  string
	err error
}

var errResourceCommandTimeout = errors.New("timed out waiting to apply command to resource")

// commandFunc takes a channel; when called should push a resourceCreateResponse onto the channel and return whether the finished executing
type commandFunc func(c chan<- resourceCommandResult) (finished bool)

// resourceApply takes a commandFunc; executes it; and returns an id.
// If the command fails to execute, returns an error.
func resourceApply(f commandFunc) (id string, err error) {
	idChan := make(chan resourceCommandResult)
	go func(c chan<- resourceCommandResult) {
		log.Println("[DEBUG] RUNMODE: " + RUNMODE)
		// linear backoff
		waitDuration := 1
		for {
			if finished := f(idChan); finished {
				return
			}
			if RUNMODE == RunmodeRecord {
				log.Printf("[DEBUG] Waiting %d second(s)\n", waitDuration)
				Sleep(time.Duration(waitDuration) * time.Second)
				waitDuration += 1
			}
		}
	}(idChan)
	for {
		select {
		case r := <-idChan:
			return r.id, r.err
		case <-time.After(maxWaitDuration):
			return "", errResourceCommandTimeout
		}
	}
}

func getCreatedState(resource interface{}) string {
	v := reflect.ValueOf(resource).Elem().FieldByName("State")

	if !v.IsValid() {
		panic("Developer error, resource passed to getCreatedState does not have a State field")
	}

	return v.Interface().(string)
}

func getID(resource interface{}) string {
	v := reflect.ValueOf(resource)
	if !v.IsValid() {
		panic("Developer error, resource passed to getID is not valid")
	}
	if v.IsNil() {
		return ""
	}
	id := v.Elem().FieldByName("ID")

	return id.Interface().(string)
}

func isResourceGoneOrTerminated(resource interface{}, err error) (bool, error) {
	if bmError, ok := err.(*bm.Error); ok {
		if bmError.Code == bm.NotAuthorizedOrNotFound {
			return true, nil
		}
	}
	if err != nil {
		return false, err
	}

	v := reflect.ValueOf(resource).Elem().FieldByName("State")

	if !v.IsValid() {
		panic("Developer error, resource passed to isResourceGoneOrTerminated does not have a state field")
	}

	state := v.Interface().(string)

	if state == bm.ResourceTerminated {
		return true, nil
	}
	return false, nil
}

func getCreateFn(err error, resource interface{}, completedState string, resourceFetchFn func() (interface{}, error)) commandFunc {
	id := getID(resource) // This wont be used if err != nil
	return func(c chan<- resourceCommandResult) bool {
		if err != nil {
			c <- resourceCommandResult{"", err}
			return true
		}
		fresh, err := resourceFetchFn()
		if err != nil {
			c <- resourceCommandResult{"", err}
			return true
		}
		if getCreatedState(fresh) == completedState {
			c <- resourceCommandResult{id, nil}
			return true
		}
		return false
	}
}

// getDeleteFn takes an error, an id, and a func to fetch the resource; returns a createFunc.
// the createFunc will fetch the resource and check if it is gone, returning
func getDeleteFn(err error, id string, resourceFetchFn func() (interface{}, error)) commandFunc {
	return func(c chan<- resourceCommandResult) bool {
		if err != nil {
			c <- resourceCommandResult{"", err}
			return true
		}
		res, erra := resourceFetchFn()
		/*if erra != nil {
			c <- resourceCommandResult{"", erra}
			return true
		}*/
		gone, errb := isResourceGoneOrTerminated(res, erra)
		if errb != nil {
			c <- resourceCommandResult{"", errb}
			return true
		}
		if gone {
			c <- resourceCommandResult{id, nil}
			return true
		}
		return false
	}
}
