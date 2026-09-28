package client

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/oracle/oci-go-sdk/v65/core"
	"go.uber.org/zap"
	"k8s.io/client-go/util/flowcontrol"
)

const concurrencyTestTimeout = 5 * time.Second

func TestAttachVolumeSerializesDeviceSelectionForSameInstance(t *testing.T) {
	compute := newConcurrentAttachmentComputeClient("instance-a")
	t.Cleanup(compute.releaseFirstAttachment)
	client := newVolumeAttachmentTestClient(compute)

	firstResult := make(chan error, 1)
	go func() {
		_, err := client.AttachVolume(context.Background(), "instance-a", "volume-a", false)
		firstResult <- err
	}()

	<-compute.firstAttachStarted

	secondResult := make(chan error, 1)
	go func() {
		_, err := client.AttachParavirtualizedVolume(context.Background(), "instance-a", "volume-b", false, false)
		secondResult <- err
	}()

	// Two lock references prove that the second call has reached the
	// same-instance lock. The unchanged list count proves it cannot select a
	// device while the first attachment still holds that lock.
	waitForInstanceAttachmentReferences(t, client, "instance-a", 2)
	if got := compute.listCallCount("instance-a"); got != 1 {
		t.Fatalf("ListInstanceDevices calls while first attachment is blocked = %d, want 1", got)
	}

	compute.releaseFirstAttachment()

	if err := <-firstResult; err != nil {
		t.Fatalf("first attachment failed: %v", err)
	}
	if err := <-secondResult; err != nil {
		t.Fatalf("second attachment failed: %v", err)
	}

	if got, want := compute.attachedDevices("instance-a"), []string{"/dev/oracleoci/oraclevdac", "/dev/oracleoci/oraclevdad"}; !sameStrings(got, want) {
		t.Fatalf("attached devices = %v, want %v", got, want)
	}

	// Both callers have released their references, so the ephemeral instance
	// must no longer retain a lock-map entry.
	if references, exists := instanceAttachmentLockState(client, "instance-a"); exists || references != 0 {
		t.Fatalf("instance attachment lock remains after final release: references = %d, exists = %t", references, exists)
	}
}

func TestAttachVolumeDoesNotSerializeDifferentInstances(t *testing.T) {
	compute := newConcurrentAttachmentComputeClient("instance-a")
	t.Cleanup(compute.releaseFirstAttachment)
	client := newVolumeAttachmentTestClient(compute)

	firstResult := make(chan error, 1)
	go func() {
		_, err := client.AttachVolume(context.Background(), "instance-a", "volume-a", false)
		firstResult <- err
	}()

	<-compute.firstAttachStarted

	secondResult := make(chan error, 1)
	go func() {
		_, err := client.AttachVolume(context.Background(), "instance-b", "volume-b", false)
		secondResult <- err
	}()

	// The mock closes this channel only when instance B enters AttachVolume.
	// Receiving it while instance A is still blocked proves locks are keyed by
	// instance rather than globally serializing all attachments.
	select {
	case <-compute.attachStartedByInstance["instance-b"]:
	case <-time.After(concurrencyTestTimeout):
		compute.releaseFirstAttachment()
		<-firstResult
		<-secondResult
		t.Fatal("attachment for different instance did not reach AttachVolume")
	}

	if err := <-secondResult; err != nil {
		t.Fatalf("attachment for different instance failed: %v", err)
	}

	compute.releaseFirstAttachment()
	if err := <-firstResult; err != nil {
		t.Fatalf("first attachment failed: %v", err)
	}
}

func TestAttachVolumeStopsWaitingWhenContextIsCancelled(t *testing.T) {
	compute := newConcurrentAttachmentComputeClient("instance-a")
	t.Cleanup(compute.releaseFirstAttachment)
	client := newVolumeAttachmentTestClient(compute)

	firstResult := make(chan error, 1)
	go func() {
		_, err := client.AttachVolume(context.Background(), "instance-a", "volume-a", false)
		firstResult <- err
	}()

	<-compute.firstAttachStarted

	ctx, cancel := context.WithCancel(context.Background())
	secondResult := make(chan error, 1)
	go func() {
		_, err := client.AttachVolume(ctx, "instance-a", "volume-b", false)
		secondResult <- err
	}()

	// Waiting for two lock references proves the second call has reached the
	// same-instance lock and is queued behind the blocked first attachment.
	waitForInstanceAttachmentReferences(t, client, "instance-a", 2)
	cancel()

	select {
	case err := <-secondResult:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("cancelled attachment error = %v, want %v", err, context.Canceled)
		}
	case <-time.After(concurrencyTestTimeout):
		// Release the first request so the old non-cancellable implementation
		// cannot leave test goroutines blocked after this regression test fails.
		compute.releaseFirstAttachment()
		<-firstResult
		<-secondResult
		t.Fatal("cancelled attachment remained blocked waiting for the instance lock")
	}

	// Cancellation while queued must not reach either OCI operation.
	if got := compute.listCallCount("instance-a"); got != 1 {
		t.Fatalf("ListInstanceDevices calls = %d, want 1", got)
	}
	if got := compute.attachCallCount("instance-a"); got != 1 {
		t.Fatalf("AttachVolume calls = %d, want 1", got)
	}

	// The cancelled waiter has removed its reference while the first caller
	// still holds the only remaining reference.
	if references, exists := instanceAttachmentLockState(client, "instance-a"); !exists || references != 1 {
		t.Fatalf("instance attachment lock after cancellation: references = %d, exists = %t, want references = 1, exists = true", references, exists)
	}

	compute.releaseFirstAttachment()
	if err := <-firstResult; err != nil {
		t.Fatalf("first attachment failed: %v", err)
	}
	if references, exists := instanceAttachmentLockState(client, "instance-a"); exists || references != 0 {
		t.Fatalf("instance attachment lock remains after final release: references = %d, exists = %t", references, exists)
	}
}

func TestAttachVolumeConsumesWriterPermitAfterLockAcquisition(t *testing.T) {
	compute := newConcurrentAttachmentComputeClient("instance-a")
	t.Cleanup(compute.releaseFirstAttachment)
	client := newVolumeAttachmentTestClient(compute)
	limiter := &countingRateLimiter{RateLimiter: flowcontrol.NewFakeAlwaysRateLimiter()}
	client.rateLimiter.Writer = limiter

	firstResult := make(chan error, 1)
	go func() {
		_, err := client.AttachVolume(context.Background(), "instance-a", "volume-a", false)
		firstResult <- err
	}()

	<-compute.firstAttachStarted

	secondResult := make(chan error, 1)
	go func() {
		_, err := client.AttachVolume(context.Background(), "instance-a", "volume-b", false)
		secondResult <- err
	}()

	// Two lock references prove the second call is queued. At that point only
	// the lock holder should have consumed a writer permit.
	waitForInstanceAttachmentReferences(t, client, "instance-a", 2)
	permitCallsWhileQueued := limiter.callCount()

	compute.releaseFirstAttachment()
	if err := <-firstResult; err != nil {
		t.Fatalf("first attachment failed: %v", err)
	}
	if err := <-secondResult; err != nil {
		t.Fatalf("second attachment failed: %v", err)
	}

	if permitCallsWhileQueued != 1 {
		t.Fatalf("writer permit calls while second attachment was queued = %d, want 1", permitCallsWhileQueued)
	}
	if got := limiter.callCount(); got != 2 {
		t.Fatalf("writer permit calls after both attachments completed = %d, want 2", got)
	}
}

func newVolumeAttachmentTestClient(compute computeClient) *client {
	return &client{
		compute: compute,
		logger:  zap.S(),
		rateLimiter: RateLimiter{
			Reader: flowcontrol.NewFakeAlwaysRateLimiter(),
			Writer: flowcontrol.NewFakeAlwaysRateLimiter(),
		},
	}
}

type countingRateLimiter struct {
	flowcontrol.RateLimiter

	mu    sync.Mutex
	calls int
}

func (l *countingRateLimiter) TryAccept() bool {
	l.mu.Lock()
	l.calls++
	l.mu.Unlock()
	return l.RateLimiter.TryAccept()
}

func (l *countingRateLimiter) callCount() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.calls
}

type concurrentAttachmentComputeClient struct {
	*mockComputeClient

	mu                      sync.Mutex
	availableDevices        map[string][]string
	attachedByInstance      map[string][]string
	listCalls               map[string]int
	attachCalls             map[string]int
	attachStarted           map[string]bool
	attachStartedByInstance map[string]chan struct{}
	blockInstance           string
	firstAttach             bool
	firstAttachStarted      chan struct{}
	releaseFirstAttach      chan struct{}
	releaseFirstAttachOnce  sync.Once
}

func (c *concurrentAttachmentComputeClient) releaseFirstAttachment() {
	c.releaseFirstAttachOnce.Do(func() {
		close(c.releaseFirstAttach)
	})
}

func newConcurrentAttachmentComputeClient(blockInstance string) *concurrentAttachmentComputeClient {
	return &concurrentAttachmentComputeClient{
		mockComputeClient: &mockComputeClient{},
		availableDevices: map[string][]string{
			"instance-a": {"/dev/oracleoci/oraclevdac", "/dev/oracleoci/oraclevdad"},
			"instance-b": {"/dev/oracleoci/oraclevdac", "/dev/oracleoci/oraclevdad"},
		},
		attachedByInstance: map[string][]string{},
		listCalls:          map[string]int{},
		attachCalls:        map[string]int{},
		attachStarted:      map[string]bool{},
		attachStartedByInstance: map[string]chan struct{}{
			"instance-a": make(chan struct{}),
			"instance-b": make(chan struct{}),
		},
		blockInstance:      blockInstance,
		firstAttachStarted: make(chan struct{}),
		releaseFirstAttach: make(chan struct{}),
	}
}

func (c *concurrentAttachmentComputeClient) ListInstanceDevices(_ context.Context, request core.ListInstanceDevicesRequest) (core.ListInstanceDevicesResponse, error) {
	if request.Limit == nil || *request.Limit != 1 {
		return core.ListInstanceDevicesResponse{}, fmt.Errorf("ListInstanceDevices limit = %v, want 1", request.Limit)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.listCalls[*request.InstanceId]++

	devices := c.availableDevices[*request.InstanceId]
	if len(devices) == 0 {
		return core.ListInstanceDevicesResponse{}, nil
	}

	device := devices[0]
	return core.ListInstanceDevicesResponse{Items: []core.Device{{Name: &device}}}, nil
}

func (c *concurrentAttachmentComputeClient) AttachVolume(_ context.Context, request core.AttachVolumeRequest) (core.AttachVolumeResponse, error) {
	instanceID, device, err := attachmentInstanceAndDevice(request.AttachVolumeDetails)
	if err != nil {
		return core.AttachVolumeResponse{}, err
	}

	c.mu.Lock()
	c.attachCalls[instanceID]++
	if started := c.attachStartedByInstance[instanceID]; started != nil && !c.attachStarted[instanceID] {
		close(started)
		c.attachStarted[instanceID] = true
	}
	if instanceID == c.blockInstance && !c.firstAttach {
		c.firstAttach = true
		c.mu.Unlock()
		close(c.firstAttachStarted)
		<-c.releaseFirstAttach
		c.mu.Lock()
	} else if !contains(c.availableDevices[instanceID], device) {
		c.mu.Unlock()
		return core.AttachVolumeResponse{}, fmt.Errorf("device %s is already being attached to %s", device, instanceID)
	}

	if !contains(c.availableDevices[instanceID], device) {
		c.mu.Unlock()
		return core.AttachVolumeResponse{}, fmt.Errorf("device %s is already being attached to %s", device, instanceID)
	}
	c.availableDevices[instanceID] = remove(c.availableDevices[instanceID], device)
	c.attachedByInstance[instanceID] = append(c.attachedByInstance[instanceID], device)
	c.mu.Unlock()

	return core.AttachVolumeResponse{}, nil
}

func attachmentInstanceAndDevice(details core.AttachVolumeDetails) (string, string, error) {
	switch details := details.(type) {
	case core.AttachIScsiVolumeDetails:
		return *details.InstanceId, *details.Device, nil
	case core.AttachParavirtualizedVolumeDetails:
		return *details.InstanceId, *details.Device, nil
	default:
		return "", "", fmt.Errorf("unexpected attachment details %T", details)
	}
}

func (c *concurrentAttachmentComputeClient) attachedDevices(instanceID string) []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]string(nil), c.attachedByInstance[instanceID]...)
}

func (c *concurrentAttachmentComputeClient) listCallCount(instanceID string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.listCalls[instanceID]
}

func (c *concurrentAttachmentComputeClient) attachCallCount(instanceID string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.attachCalls[instanceID]
}

func waitForInstanceAttachmentReferences(t *testing.T, client *client, instanceID string, want int) {
	t.Helper()

	deadline := time.Now().Add(concurrencyTestTimeout)
	for time.Now().Before(deadline) {
		client.instanceAttachmentLocksMutex.Lock()
		lock := client.instanceAttachmentLocks[instanceID]
		got := 0
		if lock != nil {
			got = lock.references
		}
		client.instanceAttachmentLocksMutex.Unlock()

		if got == want {
			return
		}
		time.Sleep(time.Millisecond)
	}

	t.Fatalf("instance attachment lock references did not reach %d", want)
}

func instanceAttachmentLockState(client *client, instanceID string) (int, bool) {
	client.instanceAttachmentLocksMutex.Lock()
	defer client.instanceAttachmentLocksMutex.Unlock()

	lock, exists := client.instanceAttachmentLocks[instanceID]
	if !exists {
		return 0, false
	}
	return lock.references, true
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func remove(values []string, want string) []string {
	for i, value := range values {
		if value == want {
			return append(values[:i], values[i+1:]...)
		}
	}
	return values
}

func sameStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func Test_getDevicePath(t *testing.T) {
	var tests = map[string]struct {
		instanceID string
		want       string
		wantErr    error
	}{
		"getDevicePathNoDeviceAvailable": {
			instanceID: "ocid1.device-path-not-available",
			wantErr:    fmt.Errorf("Max number of volumes are already attached to instance %s. Please schedule workload on different node.", "ocid1.device-path-not-available"),
		},
		"getDevicePathOneDeviceAvailable": {
			instanceID: "ocid1.one-device-path-available",
			want:       "/dev/oracleoci/oraclevdac",
		},
		"getDevicePathReturnsError": {
			instanceID: "ocid1.device-path-returns-error",
			wantErr:    errNotFound,
		},
	}

	vaClient := &client{
		compute: &mockComputeClient{},
		logger:  zap.S(),
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := vaClient.getDevicePath(context.Background(), tc.instanceID)
			if tc.wantErr != nil && !strings.EqualFold(tc.wantErr.Error(), err.Error()) {
				t.Errorf("getDevicePath() = %v, want %v", err.Error(), tc.wantErr.Error())
			}
			if tc.want != "" && !strings.EqualFold(tc.want, *result) {
				t.Errorf("getDevicePath() = %v, want %v", *result, tc.want)
			}
		})
	}
}
