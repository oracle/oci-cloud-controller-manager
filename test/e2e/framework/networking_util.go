// Copyright 2020 Oracle and/or its affiliates. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package framework

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"

	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	EndpointHttpPort           = 8080
	EndpointHttpsPort          = 443
	EndpointUdpPort            = 8081
	TestContainerHttpPort      = 8080
	ClusterHttpPort            = 80
	ClusterUdpPort             = 90
	testPodName                = "test-container-pod"
	hostTestPodName            = "host-test-container-pod"
	nodePortServiceName        = "node-port-service"
	sessionAffinityServiceName = "session-affinity-service"
	// wait time between poll attempts of a Service vip and/or nodePort.
	// coupled with testTries to produce a net timeout value.
	hitEndpointRetryDelay = 2 * time.Second
	// Number of retries to hit a given set of endpoints. Needs to be high
	// because we verify iptables statistical rr loadbalancing.
	testTries = 30
	// Maximum number of pods in a test, to make test work in large clusters.
	maxNetProxyPodsCount = 10
	// Number of checks to hit a given set of endpoints when enable session affinity.
	SessionAffinityChecks = 10
)

func getServiceSelector() map[string]string {
	By("creating a selector")
	selectorName := "selector-" + string(uuid.NewUUID())
	serviceSelector := map[string]string{
		selectorName: "true",
	}
	return serviceSelector
}

// Does an HTTP GET, but does not reuse TCP connections
// This masks problems where the iptables rule has changed, but we don't see it
// This is intended for relatively quick requests (status checks), so we set a short (5 seconds) timeout
func httpGetNoConnectionPool(url string) (*http.Response, error) {
	return httpGetNoConnectionPoolTimeout(url, 5*time.Second)
}

func httpGetNoConnectionPoolTimeout(url string, timeout time.Duration) (*http.Response, error) {
	tr := utilnet.SetTransportDefaults(&http.Transport{
		DisableKeepAlives: true,
	})
	client := &http.Client{
		Transport: tr,
		Timeout:   timeout,
	}

	return client.Get(url)
}

func TestReachableHTTP(secure bool, ip string, port int, request string, expect string) (bool, error) {
	return TestReachableHTTPWithContent(secure, ip, port, request, expect, nil)
}

func TestReachableHTTPWithRetriableErrorCodes(secure bool, ip string, port int, request string, expect string, retriableErrCodes []int) (bool, error) {
	return TestReachableHTTPWithContentTimeoutWithRetriableErrorCodes(secure, ip, port, request, expect, nil, retriableErrCodes, time.Second*5)
}

func TestReachableHTTPWithContent(secure bool, ip string, port int, request string, expect string, content *bytes.Buffer) (bool, error) {
	return TestReachableHTTPWithContentTimeout(secure, ip, port, request, expect, content, 5*time.Second)
}

func TestReachableHTTPWithContentTimeout(secure bool, ip string, port int, request string, expect string, content *bytes.Buffer, timeout time.Duration) (bool, error) {
	return TestReachableHTTPWithContentTimeoutWithRetriableErrorCodes(secure, ip, port, request, expect, content, []int{}, timeout)
}

func TestReachableHTTPWithContentTimeoutWithRetriableErrorCodes(secure bool, ip string, port int, request string, expect string, content *bytes.Buffer, retriableErrCodes []int, timeout time.Duration) (bool, error) {

	ipPort := net.JoinHostPort(ip, strconv.Itoa(port))
	url := fmt.Sprintf("http://%s%s", ipPort, request)
	if secure {
		url = fmt.Sprintf("https://%s%s", ipPort, request)
	}
	if ip == "" {
		Failf("Got empty IP for reachability check (%s)", url)
		return false, nil
	}
	if port == 0 {
		Failf("Got port==0 for reachability check (%s)", url)
		return false, nil
	}

	Logf("Testing HTTP reachability of %v", url)

	resp, err := httpGetNoConnectionPoolTimeout(url, timeout)
	if err != nil {
		Logf("Got error testing for reachability of %s: %v", url, err)
		return false, nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Logf("Got error reading response from %s: %v", url, err)
		return false, nil
	}
	if resp.StatusCode != 200 {
		for _, code := range retriableErrCodes {
			if resp.StatusCode == code {
				Logf("Got non-success status %q when trying to access %s, but the error code is retriable", resp.Status, url)
				return false, nil
			}
		}
		return false, fmt.Errorf("received non-success return status %q trying to access %s; got body: %s",
			resp.Status, url, string(body))
	}
	if !strings.Contains(string(body), expect) {
		return false, fmt.Errorf("received response body without expected substring %q: %s", expect, string(body))
	}
	if content != nil {
		content.Write(body)
	}
	return true, nil
}

func TestNotReachableHTTP(ip string, port int) (bool, error) {
	return TestNotReachableHTTPTimeout(ip, port, 5*time.Second)
}

func TestNotReachableHTTPTimeout(ip string, port int, timeout time.Duration) (bool, error) {
	ipPort := net.JoinHostPort(ip, strconv.Itoa(port))
	url := fmt.Sprintf("http://%s", ipPort)
	if ip == "" {
		Failf("Got empty IP for non-reachability check (%s)", url)
		return false, nil
	}
	if port == 0 {
		Failf("Got port==0 for non-reachability check (%s)", url)
		return false, nil
	}

	Logf("Testing HTTP non-reachability of %v", url)

	resp, err := httpGetNoConnectionPoolTimeout(url, timeout)
	if err != nil {
		Logf("Confirmed that %s is not reachable", url)
		return true, nil
	}
	resp.Body.Close()
	return false, nil
}

func TestReachableUDP(ip string, port int, request string, expect string) (bool, error) {
	ipPort := net.JoinHostPort(ip, strconv.Itoa(port))
	uri := fmt.Sprintf("udp://%s", ipPort)
	if ip == "" {
		Failf("Got empty IP for reachability check (%s)", uri)
		return false, nil
	}
	if port == 0 {
		Failf("Got port==0 for reachability check (%s)", uri)
		return false, nil
	}

	Logf("Testing UDP reachability of %v", uri)

	con, err := net.Dial("udp", ipPort)
	if err != nil {
		return false, fmt.Errorf("Failed to dial %s: %v", ipPort, err)
	}
	defer con.Close()

	_, err = con.Write([]byte(fmt.Sprintf("%s\n", request)))
	if err != nil {
		return false, fmt.Errorf("Failed to send request: %v", err)
	}

	var buf []byte = make([]byte, len(expect)+1)

	err = con.SetDeadline(time.Now().Add(3 * time.Second))
	if err != nil {
		return false, fmt.Errorf("Failed to set deadline: %v", err)
	}

	_, err = con.Read(buf)
	if err != nil {
		return false, nil
	}

	if !strings.Contains(string(buf), expect) {
		return false, fmt.Errorf("Failed to retrieve %q, got %q", expect, string(buf))
	}

	Logf("Successfully reached %v", uri)
	return true, nil
}

func TestNotReachableUDP(ip string, port int, request string) (bool, error) {
	ipPort := net.JoinHostPort(ip, strconv.Itoa(port))
	uri := fmt.Sprintf("udp://%s", ipPort)
	if ip == "" {
		Failf("Got empty IP for reachability check (%s)", uri)
		return false, nil
	}
	if port == 0 {
		Failf("Got port==0 for reachability check (%s)", uri)
		return false, nil
	}

	Logf("Testing UDP non-reachability of %v", uri)

	con, err := net.Dial("udp", ipPort)
	if err != nil {
		Logf("Confirmed that %s is not reachable", uri)
		return true, nil
	}
	defer con.Close()

	_, err = con.Write([]byte(fmt.Sprintf("%s\n", request)))
	if err != nil {
		Logf("Confirmed that %s is not reachable", uri)
		return true, nil
	}

	var buf []byte = make([]byte, 1)

	err = con.SetDeadline(time.Now().Add(3 * time.Second))
	if err != nil {
		return false, fmt.Errorf("Failed to set deadline: %v", err)
	}

	_, err = con.Read(buf)
	if err != nil {
		Logf("Confirmed that %s is not reachable", uri)
		return true, nil
	}

	return false, nil
}

func TestHitNodesFromOutside(externalIP string, httpPort int32, timeout time.Duration, expectedHosts sets.String) error {
	return TestHitNodesFromOutsideWithCount(externalIP, httpPort, timeout, expectedHosts, 1)
}

func TestHitNodesFromOutsideWithCount(externalIP string, httpPort int32, timeout time.Duration, expectedHosts sets.String,
	countToSucceed int) error {
	Logf("Waiting up to %v for satisfying expectedHosts for %v times", timeout, countToSucceed)
	hittedHosts := sets.NewString()
	count := 0
	condition := func() (bool, error) {
		var respBody bytes.Buffer
		reached, err := TestReachableHTTPWithContentTimeout(false, externalIP, int(httpPort), "/hostname", "", &respBody,
			1*time.Second)
		if err != nil || !reached {
			return false, nil
		}
		hittedHost := strings.TrimSpace(respBody.String())
		if !expectedHosts.Has(hittedHost) {
			Logf("Error hitting unexpected host: %v, reset counter: %v", hittedHost, count)
			count = 0
			return false, nil
		}
		if !hittedHosts.Has(hittedHost) {
			hittedHosts.Insert(hittedHost)
			Logf("Missing %+v, got %+v", expectedHosts.Difference(hittedHosts), hittedHosts)
		}
		if hittedHosts.Equal(expectedHosts) {
			count++
			if count >= countToSucceed {
				return true, nil
			}
		}
		return false, nil
	}

	if err := wait.Poll(time.Second, timeout, condition); err != nil {
		return fmt.Errorf("error waiting for expectedHosts: %v, hittedHosts: %v, count: %v, expected count: %v",
			expectedHosts, hittedHosts, count, countToSucceed)
	}
	return nil
}
