package framework

import (
	"crypto/rand"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/oracle/oci-go-sdk/v50/common"
)

func nowStamp() string {
	return time.Now().Format(time.StampMilli)
}

func log(level string, format string, args ...interface{}) {
	fmt.Fprintf(GinkgoWriter, nowStamp()+": "+level+": "+format+"\n", args...)
}

func Logf(format string, args ...interface{}) {
	log("INFO", format, args...)
}

func Failf(format string, args ...interface{}) {
	FailfWithOffset(1, format, args...)
}

// FailfWithOffset calls "Fail" and logs the error at "offset" levels above its caller
// (for example, for call chain f -> g -> FailfWithOffset(1, ...) error would be logged for "f").
func FailfWithOffset(offset int, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log("INFO", msg)
	Fail(nowStamp()+": "+msg, 1+offset)
}

func ExpectNoError(err error, explain ...interface{}) {
	ExpectNoErrorWithOffset(1, err, explain...)
}

// ExpectNoErrorWithOffset checks if "err" is set, and if so, fails assertion while logging the error at "offset" levels above its caller
// (for example, for call chain f -> g -> ExpectNoErrorWithOffset(1, ...) error would be logged for "f").
func ExpectNoErrorWithOffset(offset int, err error, explain ...interface{}) {
	if err != nil {
		Logf("Unexpected error occurred: %v", err)
	}
	ExpectWithOffset(1+offset, err).NotTo(HaveOccurred(), explain...)
}

// UniqueID returns a unique UUID-like identifier for use in generating
// resources for integration tests.
func UniqueID() string {
	uuid := make([]byte, 8)
	io.ReadFull(rand.Reader, uuid)
	return fmt.Sprintf("%x", uuid)
}

func checkForExpectedError(err error, expectedError common.ServiceError) {
	serviceError, isServiceError := common.IsServiceError(err)
	Expect(isServiceError).To(Equal(true))
	Expect(serviceError.GetHTTPStatusCode()).To(Equal(expectedError.GetHTTPStatusCode()))
	Expect(serviceError.GetMessage()).To(Equal(expectedError.GetMessage()))
	Expect(serviceError.GetCode()).To(Equal(expectedError.GetCode()))
}

func compareVersions(v1 string, v2 string) (ret int) {
	v1Arr := strings.Split(v1, ".")
	v2Arr := strings.Split(v2, ".")
	num := len(v2Arr)
	if len(v1Arr) > len(v2Arr) {
		num = len(v1Arr)
	}
	for i := 0; i < num; i++ {
		var x, y string
		if len(v1Arr) > i {
			x = v1Arr[i]
		}
		if len(v2Arr) > i {
			y = v2Arr[i]
		}
		if x == y {
			ret = 0
		} else {
			xi, _ := strconv.Atoi(x)
			yi, _ := strconv.Atoi(y)
			if xi > yi {
				ret = 1
			} else if xi < yi {
				ret = -1
			}
		}
		if ret != 0 {
			break
		}
	}
	return
}
