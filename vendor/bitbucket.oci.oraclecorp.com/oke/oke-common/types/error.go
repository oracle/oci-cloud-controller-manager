package types

import (
	"errors"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

var (
	ErrClusterUpFail = errors.New("CRITICAL: cluster creation failure: ")
)

// ErrorV1 an API friendly form of Error
type ErrorV1 struct {
	Message string `json:"message"`
}

func (src *Error) ToV1() (dst ErrorV1) {
	if src == nil {
		return
	}

	dst.Message = src.Message
	return dst
}

// ErrorBMC contains the parsed message contained within a
// bmc-go-sdk error
type ErrorBMC struct {
	Status    int    `json:"status",omitempty`
	Code      string `json:"code",omitempty`
	RequestID string `json:"request_id",omitempty`
	Message   string `json:"message",omitempty`
	ParseErr  error  `json:"parse_error",omitempty`
}

// Fields writes the parsed contents of a BMC error to log.Fields
func (e ErrorBMC) Fields() log.Fields {
	f := log.Fields{}
	f["bmc_status"] = e.Status
	f["bmc_code"] = e.Code
	f["bmc_request_id"] = e.RequestID
	f["bmc_message"] = e.Message
	if e.ParseErr != nil {
		f["bmc_parse_err"] = e.ParseErr.Error()
	}
	return f
}

// Error stringifies the error message
func (e ErrorBMC) Error() string {
	var output string
	if e.Status > 0 {
		output += fmt.Sprintf("Status: %d;", e.Status)
	}
	if len(e.Code) > 0 {
		output += fmt.Sprintf("Code: %s;", e.Code)
	}
	if len(e.RequestID) > 0 {
		output += fmt.Sprintf("OPC Request ID: %s; ", e.RequestID)
	}
	if len(e.Message) > 0 {
		output += fmt.Sprintf("Message: %s; ", e.Message)
	}
	if e.ParseErr != nil {
		output += fmt.Sprintf("unable to parse error: %s", e.ParseErr)
	}
	return output
}

// ErrorHTTP implements an error with attached status code
type ErrStatus interface {
	error
	Status() int
}

// NewErrorStatus creates an ErrStatus with the provided text and status
func NewErrStatus(message string, status int) ErrStatus {
	return &errorStatus{m: message, s: status}
}

type errorStatus struct {
	m string
	s int
}

func (e *errorStatus) Error() string {
	return e.m
}

func (e *errorStatus) Status() int {
	return e.s
}

func IsBMCTooManyRequestsError(err error) bool {
	e, ok := err.(ErrorBMC)
	return ok && e.Status == http.StatusTooManyRequests
}
