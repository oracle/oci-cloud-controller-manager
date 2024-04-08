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

package driver

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	authv1 "k8s.io/api/authentication/v1"
	kubeAPI "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"

	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	csi_util "github.com/oracle/oci-cloud-controller-manager/pkg/csi-util"
	"github.com/oracle/oci-cloud-controller-manager/pkg/logging"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-cloud-controller-manager/pkg/util"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/oracle/oci-go-sdk/v65/loadbalancer"
)

const (
	testMinimumVolumeSizeInBytes int64 = 50 * client.GiB
	testTimeout                        = 15 * time.Second
	testPollInterval                   = 5 * time.Second
)

var (
	inTransitEncryptionEnabled  = true
	inTransitEncryptionDisabled = false
	errNotFound                 = errors.New("not found")
	instances                   = map[string]*core.Instance{
		"inTransitEnabled": {
			LaunchOptions: &core.LaunchOptions{
				IsPvEncryptionInTransitEnabled: &inTransitEncryptionEnabled,
			},
		},
		"inTransitDisabled": {
			LaunchOptions: &core.LaunchOptions{
				IsPvEncryptionInTransitEnabled: &inTransitEncryptionDisabled,
			},
		},
		"sample-provider-id": {
			LaunchOptions: &core.LaunchOptions{
				IsPvEncryptionInTransitEnabled: &inTransitEncryptionDisabled,
			},
		},
	}

	volumes = map[string]*core.Volume{
		"volume-in-provisioning-state": {
			DisplayName:        common.String("volume-in-provisioning-state"),
			LifecycleState:     core.VolumeLifecycleStateProvisioning,
			SizeInMBs:          common.Int64(50000),
			AvailabilityDomain: common.String("NWuj:PHX-AD-2"),
			Id:                 common.String("volume-in-provisioning-state"),
		},
		"volume-in-available-state": {
			DisplayName:        common.String("volume-in-available-state"),
			LifecycleState:     core.VolumeLifecycleStateAvailable,
			SizeInMBs:          common.Int64(50000),
			AvailabilityDomain: common.String("NWuj:PHX-AD-2"),
			Id:                 common.String("volume-in-available-state"),
		},
		"clone-volume-in-provisioning-state": {
			DisplayName:        common.String("clone-volume-in-provisioning-state"),
			LifecycleState:     core.VolumeLifecycleStateProvisioning,
			SizeInMBs:          common.Int64(50000),
			AvailabilityDomain: common.String("NWuj:PHX-AD-2"),
			Id:                 common.String("clone-volume-in-provisioning-state"),
		},
	}

	create_volume_requests = map[string]*csi.CreateVolumeRequest{
		"volume-stuck-in-provisioning-state": {
			Name: "volume-in-provisioning-state",
			VolumeCapabilities: []*csi.VolumeCapability{{
				AccessType: &csi.VolumeCapability_Mount{
					Mount: &csi.VolumeCapability_MountVolume{},
				},
				AccessMode: &csi.VolumeCapability_AccessMode{
					Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
				},
			}},
			AccessibilityRequirements: &csi.TopologyRequirement{Requisite: []*csi.Topology{
				{
					Segments: map[string]string{kubeAPI.LabelZoneFailureDomain: "ad1"},
				},
			},
			},
		},
		"clone-volume-stuck-in-provisioning-state": {
			Name: "clone-volume-in-provisioning-state",
			VolumeContentSource: &csi.VolumeContentSource{
				Type: &csi.VolumeContentSource_Volume{
					Volume: &csi.VolumeContentSource_VolumeSource{
						VolumeId: "volume-in-available-state",
					},
				},
			},
			CapacityRange: &csi.CapacityRange{
				RequiredBytes: 50000 * client.MiB,
			},
			VolumeCapabilities: []*csi.VolumeCapability{{
				AccessType: &csi.VolumeCapability_Mount{
					Mount: &csi.VolumeCapability_MountVolume{},
				},
				AccessMode: &csi.VolumeCapability_AccessMode{
					Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
				},
			}},
			AccessibilityRequirements: &csi.TopologyRequirement{Requisite: []*csi.Topology{
				{
					Segments: map[string]string{kubeAPI.LabelZoneFailureDomain: "ad1"},
				},
			},
			},
		},
	}

	volume_attachments = map[string]*core.IScsiVolumeAttachment{
		"volume-attachment-stuck-in-detaching-state": {
			DisplayName:        common.String("volume-attachment-stuck-in-detaching-state"),
			LifecycleState:     core.VolumeAttachmentLifecycleStateDetaching,
			AvailabilityDomain: common.String("NWuj:PHX-AD-2"),
			Id:                 common.String("volume-attachment-stuck-in-detaching-state"),
			InstanceId:         common.String("sample-instance-id"),
		},
		"volume-attachment-stuck-in-attaching-state": {
			DisplayName:        common.String("volume-attachment-stuck-in-attaching-state"),
			LifecycleState:     core.VolumeAttachmentLifecycleStateAttaching,
			AvailabilityDomain: common.String("NWuj:PHX-AD-2"),
			Id:                 common.String("volume-attachment-stuck-in-attaching-state"),
			InstanceId:         common.String("sample-provider-id"),
		},
	}
)

type MockOCIClient struct{}

func (MockOCIClient) Compute() client.ComputeInterface {
	return &MockComputeClient{}
}

func (MockOCIClient) LoadBalancer(logger *zap.SugaredLogger, lbType string, tenancy string, token *authv1.TokenRequest) client.GenericLoadBalancerInterface {
	return &MockLoadBalancerClient{}
}

func (MockOCIClient) Networking() client.NetworkingInterface {
	return &MockVirtualNetworkClient{}
}

func (MockOCIClient) BlockStorage() client.BlockStorageInterface {
	return &MockBlockStorageClient{}
}

func (MockOCIClient) FSS() client.FileStorageInterface {
	return &MockFileStorageClient{}
}

func (MockOCIClient) Identity() client.IdentityInterface {
	return &MockIdentityClient{}
}

type MockBlockStorageClient struct {
	bs util.MockOCIBlockStorageClient
}

// AwaitVolumeCloneAvailableOrTimeout implements client.BlockStorageInterface.
func (c *MockBlockStorageClient) AwaitVolumeCloneAvailableOrTimeout(ctx context.Context, id string) (*core.Volume, error) {
	var volClone *core.Volume
	if err := wait.PollImmediateUntil(testPollInterval, func() (bool, error) {
		var err error
		volClone, err = c.GetVolume(ctx, id)
		if err != nil {
			if !client.IsRetryable(err) {
				return false, err
			}
			return false, nil
		}

		switch state := volClone.LifecycleState; state {
		case core.VolumeLifecycleStateAvailable:
			if *volClone.IsHydrated == true {
				return true, nil
			}
			return false, nil
		case core.VolumeLifecycleStateFaulty,
			core.VolumeLifecycleStateTerminated,
			core.VolumeLifecycleStateTerminating:
			return false, errors.Errorf("Clone volume did not become available and hydrated (lifecycleState=%q) (hydrationStatus=%v)", state, *volClone.IsHydrated)
		}
		return false, nil
	}, ctx.Done()); err != nil {
		return nil, err
	}
	return &core.Volume{}, nil
}

func (c *MockBlockStorageClient) AwaitVolumeBackupAvailableOrTimeout(ctx context.Context, id string) (*core.VolumeBackup, error) {
	return &core.VolumeBackup{}, nil
}

func (c *MockBlockStorageClient) CreateVolumeBackup(ctx context.Context, details core.CreateVolumeBackupDetails) (*core.VolumeBackup, error) {
	id := "oc1.volumebackup1.xxxx"
	return &core.VolumeBackup{
		Id: &id,
	}, nil
}

func (c *MockBlockStorageClient) DeleteVolumeBackup(ctx context.Context, id string) error {
	return nil
}

func (c *MockBlockStorageClient) GetVolumeBackup(ctx context.Context, id string) (*core.VolumeBackup, error) {
	return &core.VolumeBackup{
		Id: &id,
	}, nil
}

func (c *MockBlockStorageClient) GetVolumeBackupsByName(ctx context.Context, snapshotName, compartmentID string) ([]core.VolumeBackup, error) {
	return []core.VolumeBackup{}, nil
}

type MockProvisionerClient struct {
	Storage *MockBlockStorageClient
}

func (c *MockBlockStorageClient) AwaitVolumeAvailableORTimeout(ctx context.Context, id string) (*core.Volume, error) {
	volume := volumes[id]
	if volume == nil {
		return &core.Volume{}, nil
	}
	if err := wait.PollImmediateUntil(testPollInterval, func() (bool, error) {
		if volume.LifecycleState != core.VolumeLifecycleStateAvailable {
			return false, nil
		}
		return true, nil
	}, ctx.Done()); err != nil {
		return nil, err
	}
	return &core.Volume{}, nil
}

func (c *MockBlockStorageClient) GetVolume(ctx context.Context, id string) (*core.Volume, error) {
	if id == "invalid_volume_id" {
		return nil, fmt.Errorf("failed to find existence of volume")
	} else if id == "valid_volume_id" {
		ad := "zkJl:US-ASHBURN-AD-1"
		var oldSizeInBytes = int64(csi_util.MaximumVolumeSizeInBytes)
		oldSizeInGB := csi_util.RoundUpSize(oldSizeInBytes, 1*client.GiB)
		return &core.Volume{
			Id:                 &id,
			AvailabilityDomain: &ad,
			SizeInGBs:          &oldSizeInGB,
		}, nil
	} else if id == "valid_volume_id_valid_old_size_fail" {
		ad := "zkJl:US-ASHBURN-AD-1"
		vpuspergb := int64(10)
		var oldSizeInBytes int64 = 2147483648
		oldSizeInGB := csi_util.RoundUpSize(oldSizeInBytes, 1*client.GiB)
		return &core.Volume{
			Id:                 &id,
			AvailabilityDomain: &ad,
			SizeInGBs:          &oldSizeInGB,
			VpusPerGB:          &vpuspergb,
		}, nil
	} else if id == "uhp_volume_id" {
		ad := "zkJl:US-ASHBURN-AD-1"
		vpuspergb := int64(40)
		var oldSizeInBytes int64 = 2147483648
		oldSizeInGB := csi_util.RoundUpSize(oldSizeInBytes, 1*client.GiB)
		return &core.Volume{
			Id:                 &id,
			AvailabilityDomain: &ad,
			SizeInGBs:          &oldSizeInGB,
			VpusPerGB:          &vpuspergb,
		}, nil
	} else {
		return volumes[id], nil
	}
}

func (c *MockBlockStorageClient) GetVolumesByName(ctx context.Context, volumeName, compartmentID string) ([]core.Volume, error) {
	if volumeName == "get-volumes-by-name-timeout-volume" {
		var page *string
		var requestMetadata common.RequestMetadata
		volumeList := make([]core.Volume, 0)
		for {
			listVolumeResponse, err := c.bs.ListVolumes(ctx,
				core.ListVolumesRequest{
					CompartmentId:   &compartmentID,
					Page:            page,
					DisplayName:     &volumeName,
					RequestMetadata: requestMetadata,
				})

			if err != nil {
				return nil, errors.WithStack(err)
			}

			for _, volume := range listVolumeResponse.Items {
				volumeState := volume.LifecycleState
				if volumeState == core.VolumeLifecycleStateAvailable ||
					volumeState == core.VolumeLifecycleStateProvisioning {
					volumeList = append(volumeList, volume)
				}
			}

			if page = listVolumeResponse.OpcNextPage; page == nil {
				break
			}
		}
	}
	return []core.Volume{}, nil
}

// CreateVolume mocks the BlockStorage CreateVolume implementation
func (c *MockBlockStorageClient) CreateVolume(ctx context.Context, details core.CreateVolumeDetails) (*core.Volume, error) {
	volume := volumes[*details.DisplayName]
	if volume != nil {
		return volume, nil
	}

	id := "oc1.volume1.xxxx"
	ad := "zkJl:US-ASHBURN-AD-1"
	return &core.Volume{
		Id:                 &id,
		AvailabilityDomain: &ad,
	}, nil
}

func (c *MockBlockStorageClient) UpdateVolume(ctx context.Context, volumeId string, details core.UpdateVolumeDetails) (*core.Volume, error) {
	if volumeId == "valid_volume_id_valid_old_size_fail" {
		return nil, fmt.Errorf("Update volume failed")
	} else {
		ad := "zkJl:US-ASHBURN-AD-1"
		return &core.Volume{
			Id:                 &volumeId,
			AvailabilityDomain: &ad,
		}, nil
	}
}

// DeleteVolume mocks the BlockStorage DeleteVolume implementation
func (c *MockBlockStorageClient) DeleteVolume(ctx context.Context, id string) error {
	return nil
}

func (c *MockBlockStorageClient) AwaitVolumeAvailable(ctx context.Context, id string) (*core.Volume, error) {
	return nil, nil
}

// BlockStorage mocks client BlockStorage implementation
func (p *MockProvisionerClient) BlockStorage() client.BlockStorageInterface {
	return p.Storage
}

// MockVirtualNetworkClient mocks VirtualNetwork client implementation
type MockVirtualNetworkClient struct {
}

func (c *MockVirtualNetworkClient) CreateNetworkSecurityGroup(ctx context.Context, compartmentId, vcnId, displayName, lbID string) (*core.NetworkSecurityGroup, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) UpdateNetworkSecurityGroup(ctx context.Context, id, etag string, freeformTags map[string]string) (*core.NetworkSecurityGroup, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) GetNetworkSecurityGroup(ctx context.Context, id string) (*core.NetworkSecurityGroup, *string, error) {
	return nil, nil, nil
}

func (c *MockVirtualNetworkClient) ListNetworkSecurityGroups(ctx context.Context, displayName, compartmentId, vcnId string) ([]core.NetworkSecurityGroup, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) DeleteNetworkSecurityGroup(ctx context.Context, id, etag string) (*string, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) AddNetworkSecurityGroupSecurityRules(ctx context.Context, id string, details core.AddNetworkSecurityGroupSecurityRulesDetails) (*core.AddNetworkSecurityGroupSecurityRulesResponse, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) RemoveNetworkSecurityGroupSecurityRules(ctx context.Context, id string, details core.RemoveNetworkSecurityGroupSecurityRulesDetails) (*core.RemoveNetworkSecurityGroupSecurityRulesResponse, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) ListNetworkSecurityGroupSecurityRules(ctx context.Context, id string, direction core.ListNetworkSecurityGroupSecurityRulesDirectionEnum) ([]core.SecurityRule, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) UpdateNetworkSecurityGroupSecurityRules(ctx context.Context, id string, details core.UpdateNetworkSecurityGroupSecurityRulesDetails) (*core.UpdateNetworkSecurityGroupSecurityRulesResponse, error) {
	return nil, nil
}

// GetPrivateIp mocks the VirtualNetwork GetPrivateIp implementation
func (c *MockVirtualNetworkClient) GetPrivateIp(ctx context.Context, id string) (*core.PrivateIp, error) {
	privateIpAddress := "10.0.20.1"
	return &core.PrivateIp{
		IpAddress: &privateIpAddress,
	}, nil
}

func (c *MockVirtualNetworkClient) GetSubnet(ctx context.Context, id string) (*core.Subnet, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) GetSubnetFromCacheByIP(ip string) (*core.Subnet, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) GetVcn(ctx context.Context, id string) (*core.Vcn, error) {
	return &core.Vcn{}, nil
}

func (c *MockVirtualNetworkClient) GetSecurityList(ctx context.Context, id string) (core.GetSecurityListResponse, error) {
	return core.GetSecurityListResponse{}, nil
}

func (c *MockVirtualNetworkClient) UpdateSecurityList(ctx context.Context, id string, etag string, ingressRules []core.IngressSecurityRule, egressRules []core.EgressSecurityRule) (core.UpdateSecurityListResponse, error) {
	return core.UpdateSecurityListResponse{}, nil
}

func (c *MockVirtualNetworkClient) IsRegionalSubnet(ctx context.Context, id string) (bool, error) {
	return true, nil
}

func (c *MockVirtualNetworkClient) GetPublicIpByIpAddress(ctx context.Context, id string) (*core.PublicIp, error) {
	return nil, nil
}

// Networking mocks client VirtualNetwork implementation.
func (p *MockProvisionerClient) Networking() client.NetworkingInterface {
	return &MockVirtualNetworkClient{}
}

type MockLoadBalancerClient struct{}

func (c *MockLoadBalancerClient) ListWorkRequests(ctx context.Context, compartmentId, lbId string) ([]*client.GenericWorkRequest, error) {
	return nil, nil
}

func (c *MockLoadBalancerClient) CreateLoadBalancer(ctx context.Context, details *client.GenericCreateLoadBalancerDetails) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) GetLoadBalancer(ctx context.Context, id string) (*client.GenericLoadBalancer, error) {
	return nil, nil
}

func (c *MockLoadBalancerClient) GetLoadBalancerByName(ctx context.Context, compartmentID string, name string) (*client.GenericLoadBalancer, error) {
	return nil, nil
}

func (c *MockLoadBalancerClient) DeleteLoadBalancer(ctx context.Context, id string) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) GetCertificateByName(ctx context.Context, lbID string, name string) (*client.GenericCertificate, error) {
	return nil, nil
}

func (c *MockLoadBalancerClient) CreateCertificate(ctx context.Context, lbID string, cert *client.GenericCertificate) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) CreateBackendSet(ctx context.Context, lbID string, name string, details *client.GenericBackendSetDetails) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) UpdateBackendSet(ctx context.Context, lbID string, name string, details *client.GenericBackendSetDetails) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) DeleteBackendSet(ctx context.Context, lbID, name string) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) UpdateListener(ctx context.Context, lbID string, name string, details *client.GenericListener) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) CreateListener(ctx context.Context, lbID string, name string, details *client.GenericListener) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) DeleteListener(ctx context.Context, lbID, name string) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) UpdateLoadBalancerShape(context.Context, string, *client.GenericUpdateLoadBalancerShapeDetails) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) CreateBackend(ctx context.Context, lbID, bsName string, details loadbalancer.BackendDetails) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) DeleteBackend(ctx context.Context, lbID, bsName, name string) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) AwaitWorkRequest(ctx context.Context, id string) (*client.GenericWorkRequest, error) {
	return nil, nil
}

func (c *MockLoadBalancerClient) UpdateNetworkSecurityGroups(context.Context, string, []string) (string, error) {
	return "", nil
}

// Networking mocks client VirtualNetwork implementation.
func (p *MockProvisionerClient) LoadBalancer(*zap.SugaredLogger, string, string, *authv1.TokenRequest) client.GenericLoadBalancerInterface {
	return &MockLoadBalancerClient{}
}

type MockComputeClient struct {
	compute util.MockOCIComputeClient
}

// GetInstance gets information about the specified instance.
func (c *MockComputeClient) GetInstance(ctx context.Context, id string) (*core.Instance, error) {
	if instance, ok := instances[id]; ok {
		return instance, nil
	}
	return nil, errors.New("instance not found")
}

// GetInstanceByNodeName gets the OCI instance corresponding to the given
// Kubernetes node name.
func (c *MockComputeClient) GetInstanceByNodeName(ctx context.Context, compartmentID, vcnID, nodeName string) (*core.Instance, error) {
	return nil, nil
}

func (c *MockComputeClient) GetPrimaryVNICForInstance(ctx context.Context, compartmentID, instanceID string) (*core.Vnic, error) {
	return nil, nil
}

func (c *MockComputeClient) GetSecondaryVNICForInstance(ctx context.Context, compartmentID, instanceID string) (*core.Vnic, error) {
	return nil, nil
}

func (c *MockComputeClient) FindVolumeAttachment(ctx context.Context, compartmentID, volumeID string) (core.VolumeAttachment, error) {
	var page *string
	var requestMetadata common.RequestMetadata
	for {
		resp, err := c.compute.ListVolumeAttachments(ctx, core.ListVolumeAttachmentsRequest{
			CompartmentId:   &compartmentID,
			VolumeId:        &volumeID,
			Page:            page,
			RequestMetadata: requestMetadata,
		})

		if err != nil {
			return nil, errors.WithStack(err)
		}

		for _, attachment := range resp.Items {
			state := attachment.GetLifecycleState()
			if state == core.VolumeAttachmentLifecycleStateAttaching ||
				state == core.VolumeAttachmentLifecycleStateAttached {
				return attachment, nil
			}
			if state == core.VolumeAttachmentLifecycleStateDetaching {
				return attachment, errors.WithStack(errNotFound)
			}
		}

		if page = resp.OpcNextPage; page == nil {
			break
		}
	}
	if volume_attachments[volumeID] != nil {
		return volume_attachments[volumeID], nil
	}
	return nil, nil
}

func (c *MockComputeClient) FindActiveVolumeAttachment(ctx context.Context, compartmentID, volumeID string) (core.VolumeAttachment, error) {
	if volumeID == "find-active-volume-attachment-timeout-volume" {
		var page *string
		var requestMetadata common.RequestMetadata
		for {
			resp, err := c.compute.ListVolumeAttachments(ctx, core.ListVolumeAttachmentsRequest{
				CompartmentId:   &compartmentID,
				VolumeId:        &volumeID,
				Page:            page,
				RequestMetadata: requestMetadata,
			})

			if err != nil {
				return nil, errors.WithStack(err)
			}

			for _, attachment := range resp.Items {
				state := attachment.GetLifecycleState()
				if state == core.VolumeAttachmentLifecycleStateAttaching ||
					state == core.VolumeAttachmentLifecycleStateAttached ||
					state == core.VolumeAttachmentLifecycleStateDetaching {
					return attachment, nil
				}
			}

			if page = resp.OpcNextPage; page == nil {
				break
			}
		}
	}
	if volume_attachments[volumeID] != nil {
		return volume_attachments[volumeID], nil
	}
	return nil, nil
}

func (c *MockComputeClient) AttachParavirtualizedVolume(ctx context.Context, instanceID, volumeID string, isPvEncryptionInTransitEnabled bool) (core.VolumeAttachment, error) {
	return nil, nil
}

func (c *MockComputeClient) AttachVolume(ctx context.Context, instanceID, volumeID string) (core.VolumeAttachment, error) {
	return nil, nil
}

func (c *MockComputeClient) WaitForVolumeAttached(ctx context.Context, attachmentID string) (core.VolumeAttachment, error) {
	var va core.VolumeAttachment
	if err := wait.PollImmediateUntil(testPollInterval, func() (done bool, err error) {
		if va, err = c.GetVolumeAttachment(ctx, attachmentID); err != nil {
			if client.IsRetryable(err) {
				return false, nil
			}
			return true, errors.WithStack(err)
		}
		switch state := va.GetLifecycleState(); state {
		case core.VolumeAttachmentLifecycleStateAttached:
			return true, nil
		case core.VolumeAttachmentLifecycleStateDetaching, core.VolumeAttachmentLifecycleStateDetached:
			return false, errors.Errorf("attachment %q in lifecycle state %q", *(va.GetId()), state)
		}
		return false, nil
	}, ctx.Done()); err != nil {
		return nil, errors.WithStack(err)
	}

	return va, nil
}

func (c *MockComputeClient) GetVolumeAttachment(ctx context.Context, id string) (core.VolumeAttachment, error) {
	var requestMetadata common.RequestMetadata
	resp, err := c.compute.GetVolumeAttachment(ctx, core.GetVolumeAttachmentRequest{
		VolumeAttachmentId: &id,
		RequestMetadata:    requestMetadata,
	})

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return resp.VolumeAttachment, nil
}

func (c *MockComputeClient) DetachVolume(ctx context.Context, id string) error { return nil }

func (c *MockComputeClient) WaitForVolumeDetached(ctx context.Context, attachmentID string) error {
	if err := wait.PollImmediateUntil(testPollInterval, func() (done bool, err error) {
		va := volume_attachments[attachmentID]
		if va.GetLifecycleState() == core.VolumeAttachmentLifecycleStateDetached {
			return true, nil
		}
		return false, nil
	}, ctx.Done()); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (p *MockProvisionerClient) Compute() client.ComputeInterface {
	return &MockComputeClient{}
}

// MockIdentityClient mocks identity client structure
type MockIdentityClient struct {
	common.BaseClient
}

func (client MockIdentityClient) ListAvailabilityDomains(ctx context.Context, compartmentID string) ([]identity.AvailabilityDomain, error) {
	return nil, nil
}

// ListAvailabilityDomains mocks the client ListAvailabilityDomains implementation
func (client MockIdentityClient) GetAvailabilityDomainByName(ctx context.Context, compartmentID, name string) (*identity.AvailabilityDomain, error) {
	ad1 := "AD1"
	return &identity.AvailabilityDomain{Name: &ad1}, nil
}

// Identity mocks client Identity implementation
func (p *MockProvisionerClient) Identity() client.IdentityInterface {
	return &MockIdentityClient{}
}

func NewClientProvisioner(pcData client.Interface, storageBlock *MockBlockStorageClient, storageFile *MockFileStorageClient) client.Interface {
	if storageFile == nil {
		return &MockProvisionerClient{Storage: storageBlock}
	}
	return &MockFSSProvisionerClient{Storage: storageFile}
}

func TestControllerDriver_CreateVolume(t *testing.T) {
	type fields struct {
		KubeClient kubernetes.Interface
		logger     *zap.SugaredLogger
		config     *providercfg.Config
		client     client.Interface
		util       *csi_util.Util
	}
	type args struct {
		ctx context.Context
		req *csi.CreateVolumeRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *csi.CreateVolumeResponse
		wantErr error
	}{
		{
			name:   "Error for name not provided for creating volume",
			fields: fields{},
			args: args{
				ctx: nil,
				req: &csi.CreateVolumeRequest{Name: ""},
			},
			want:    nil,
			wantErr: errors.New("CreateVolume Name must be provided"),
		},
		{
			name:   "Error for unsupported VolumeCapabilities: MULTI_NODE_MULTI_WRITER provided in CreateVolumeRequest",
			fields: fields{},
			args: args{
				ctx: nil,
				req: &csi.CreateVolumeRequest{
					Name: "ut-volume",
					VolumeCapabilities: []*csi.VolumeCapability{{
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
						},
					}},
				},
			},
			want:    nil,
			wantErr: errors.New("invalid volume capabilities requested. Only SINGLE_NODE_WRITER is supported ('accessModes.ReadWriteOnce' on Kubernetes)"),
		},
		{
			name:   "Error for no VolumeCapabilities provided in CreateVolumeRequest",
			fields: fields{},
			args: args{
				ctx: nil,
				req: &csi.CreateVolumeRequest{
					Name:               "ut-volume",
					VolumeCapabilities: []*csi.VolumeCapability{},
				},
			},
			want:    nil,
			wantErr: errors.New("VolumeCapabilities must be provided in CreateVolumeRequest"),
		},
		{
			name:   "Error for unsupported VolumeCapabilities: MULTI_NODE_READER_ONLY provided in CreateVolumeRequest",
			fields: fields{},
			args: args{
				ctx: nil,
				req: &csi.CreateVolumeRequest{
					Name: "ut-volume",
					VolumeCapabilities: []*csi.VolumeCapability{{
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY,
						},
					}},
				},
			},
			want:    nil,
			wantErr: errors.New("invalid volume capabilities requested. Only SINGLE_NODE_WRITER is supported ('accessModes.ReadWriteOnce' on Kubernetes)"),
		},
		{
			name:   "Error for unsupported VolumeCapabilities: MULTI_NODE_SINGLE_WRITER provided in CreateVolumeRequest",
			fields: fields{},
			args: args{
				ctx: nil,
				req: &csi.CreateVolumeRequest{
					Name: "ut-volume",
					VolumeCapabilities: []*csi.VolumeCapability{{
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_SINGLE_WRITER,
						},
					}},
				},
			},
			want:    nil,
			wantErr: errors.New("invalid volume capabilities requested. Only SINGLE_NODE_WRITER is supported ('accessModes.ReadWriteOnce' on Kubernetes)"),
		},
		{
			name:   "Error for exceeding capacity range",
			fields: fields{},
			args: args{
				ctx: nil,
				req: &csi.CreateVolumeRequest{
					Name: "ut-volume",
					VolumeCapabilities: []*csi.VolumeCapability{{
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
						},
					}},
					CapacityRange: &csi.CapacityRange{
						RequiredBytes: int64(csi_util.MaximumVolumeSizeInBytes) + int64(1024),
						LimitBytes:    csi_util.MinimumVolumeSizeInBytes,
					},
				},
			},
			want:    nil,
			wantErr: errors.New("invalid capacity range"),
		},
		{
			name:   "Error in finding topology requirement",
			fields: fields{},
			args: args{
				ctx: nil,
				req: &csi.CreateVolumeRequest{
					Name: "ut-volume",
					VolumeCapabilities: []*csi.VolumeCapability{
						{
							AccessMode: &csi.VolumeCapability_AccessMode{
								Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
							},
						}},
					CapacityRange: &csi.CapacityRange{
						RequiredBytes: int64(csi_util.MaximumVolumeSizeInBytes),
					},
					AccessibilityRequirements: &csi.TopologyRequirement{Requisite: []*csi.Topology{
						{
							Segments: map[string]string{"x": "ad1"},
						},
					},
					},
				},
			},
			want:    nil,
			wantErr: errors.New("required in PreferredTopologies or allowedTopologies"),
		},
		{
			name:   "Error for unsupported volumeMode Block",
			fields: fields{},
			args: args{
				ctx: nil,
				req: &csi.CreateVolumeRequest{
					Name: "ut-volume",
					VolumeCapabilities: []*csi.VolumeCapability{{
						AccessType: &csi.VolumeCapability_Block{
							Block: &csi.VolumeCapability_BlockVolume{},
						},
					}},
				},
			},
			want:    nil,
			wantErr: errors.New("driver does not support Block volumeMode. Please use Filesystem mode"),
		},
		{
			name:   "Create Volume times out waiting for volume to become available",
			fields: fields{},
			args: args{
				req: create_volume_requests["volume-stuck-in-provisioning-state"],
			},
			want:    nil,
			wantErr: errors.New("Create volume failed with time out timed out waiting for the condition"),
		},
		{
			name:   "GetVolumesByName times out due to multiple pages of ListVolumes results",
			fields: fields{},
			args: args{
				req: &csi.CreateVolumeRequest{
					Name: "get-volumes-by-name-timeout-volume",
					VolumeCapabilities: []*csi.VolumeCapability{{
						AccessType: &csi.VolumeCapability_Mount{
							Mount: &csi.VolumeCapability_MountVolume{},
						},
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
						},
					}},
					AccessibilityRequirements: &csi.TopologyRequirement{Requisite: []*csi.Topology{
						{
							Segments: map[string]string{kubeAPI.LabelZoneFailureDomain: "ad1"},
						},
					},
					},
				},
			},
			want:    nil,
			wantErr: errors.New("failed to check existence of volume context deadline exceeded"),
		},
		{
			name:   "Create Volume times out waiting for cloned volume to become available",
			fields: fields{},
			args: args{
				req: create_volume_requests["clone-volume-stuck-in-provisioning-state"],
			},
			want:    nil,
			wantErr: errors.New("Create volume failed with time out timed out waiting for the condition"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
			defer cancel()
			d := &BlockVolumeControllerDriver{ControllerDriver{
				KubeClient: nil,
				logger:     zap.S(),
				config:     &providercfg.Config{CompartmentID: ""},
				client:     NewClientProvisioner(nil, &MockBlockStorageClient{}, nil),
				util:       &csi_util.Util{Logger: logging.Logger().Sugar()},
			}}
			got, err := d.CreateVolume(ctx, tt.args.req)
			if tt.wantErr == nil && err != nil {
				t.Errorf("got error %q, want none", err)
			}
			if tt.wantErr != nil && err == nil {
				t.Errorf("want error %q, got none", tt.wantErr)
			} else if tt.wantErr != nil && !strings.Contains(err.Error(), tt.wantErr.Error()) {
				t.Errorf("want error %q to include %q", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ControllerDriver.CreateVolume() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestControllerDriver_DeleteVolume(t *testing.T) {
	type fields struct {
		KubeClient kubernetes.Interface
		logger     *zap.SugaredLogger
		config     *providercfg.Config
		client     client.Interface
		util       *csi_util.Util
	}
	type args struct {
		ctx context.Context
		req *csi.DeleteVolumeRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *csi.DeleteVolumeResponse
		wantErr error
	}{
		{
			name:   "Error for volume OCID missing in delete block volume",
			fields: fields{},
			args: args{
				ctx: nil,
				req: &csi.DeleteVolumeRequest{},
			},
			want:    nil,
			wantErr: errors.New("DeleteVolume Volume ID must be provided"),
		},
		{
			name:   "Delete volume and get empty response",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				req: &csi.DeleteVolumeRequest{VolumeId: "oc1.volume1.xxxx"},
			},
			want:    &csi.DeleteVolumeResponse{},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &BlockVolumeControllerDriver{ControllerDriver{
				KubeClient: nil,
				logger:     zap.S(),
				config:     &providercfg.Config{CompartmentID: ""},
				client:     NewClientProvisioner(nil, &MockBlockStorageClient{}, nil),
				util:       &csi_util.Util{},
			}}
			got, err := d.DeleteVolume(tt.args.ctx, tt.args.req)
			if tt.wantErr == nil && err != nil {
				t.Errorf("got error %q, want none", err)
			}
			if tt.wantErr != nil && !strings.Contains(err.Error(), tt.wantErr.Error()) {
				t.Errorf("want error %q to include %q", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ControllerDriver.DeleteVolume() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestControllerDriver_ControllerPublishVolume(t *testing.T) {
	type args struct {
		ctx context.Context
		req *csi.ControllerPublishVolumeRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *csi.ControllerPublishVolumeResponse
		wantErr error
	}{
		{
			name: "FindActiveVolumeAttachment times out",
			args: args{
				req: &csi.ControllerPublishVolumeRequest{
					VolumeId: "find-active-volume-attachment-timeout-volume",
					NodeId:   "sample-node-id",
					VolumeCapability: &csi.VolumeCapability{
						AccessType: &csi.VolumeCapability_Mount{
							Mount: &csi.VolumeCapability_MountVolume{},
						},
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
						},
					},
				},
			},
			want:    nil,
			wantErr: errors.New("context deadline exceeded"),
		},
		{
			name: "WaitForVolumeAttached times out",
			args: args{
				req: &csi.ControllerPublishVolumeRequest{
					VolumeId: "volume-attachment-stuck-in-attaching-state",
					NodeId:   "sample-provider-id",
					VolumeCapability: &csi.VolumeCapability{
						AccessType: &csi.VolumeCapability_Mount{
							Mount: &csi.VolumeCapability_MountVolume{},
						},
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
						},
					},
				},
			},
			want:    nil,
			wantErr: errors.New("Failed to attach volume to the node: timed out waiting for the condition"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
			defer cancel()
			d := &BlockVolumeControllerDriver{ControllerDriver{
				KubeClient: &util.MockKubeClient{
					CoreClient: &util.MockCoreClient{},
				},
				logger: zap.S(),
				config: &providercfg.Config{CompartmentID: ""},
				client: NewClientProvisioner(nil, &MockBlockStorageClient{}, nil),
				util:   &csi_util.Util{Logger: logging.Logger().Sugar()},
			}}
			got, err := d.ControllerPublishVolume(ctx, tt.args.req)
			if tt.wantErr == nil && err != nil {
				t.Errorf("got error %q, want none", err)
			}
			if tt.wantErr != nil && err == nil {
				t.Errorf("want error %q, got none", tt.wantErr)
			} else if tt.wantErr != nil && !strings.Contains(err.Error(), tt.wantErr.Error()) {
				t.Errorf("want error %q to include %q", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ControllerDriver.ControllerPublish() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestControllerDriver_ControllerUnpublishVolume(t *testing.T) {
	type args struct {
		ctx context.Context
		req *csi.ControllerUnpublishVolumeRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *csi.ControllerUnpublishVolumeResponse
		wantErr error
	}{
		{
			name: "Volume stuck in detaching state",
			args: args{
				req: &csi.ControllerUnpublishVolumeRequest{
					VolumeId: "volume-attachment-stuck-in-detaching-state",
					NodeId:   "sample-node-id",
				},
			},
			want:    nil,
			wantErr: errors.New("timed out waiting for volume to be detached"),
		},
		{
			name: "FindVolumeAttachment times out",
			args: args{
				req: &csi.ControllerUnpublishVolumeRequest{
					VolumeId: "find-volume-attachment-timeout",
					NodeId:   "sample-node-id",
				},
			},
			want:    nil,
			wantErr: errors.New("context deadline exceeded"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
			defer cancel()
			d := &BlockVolumeControllerDriver{ControllerDriver{
				KubeClient: &util.MockKubeClient{
					CoreClient: &util.MockCoreClient{},
				},
				logger: zap.S(),
				config: &providercfg.Config{CompartmentID: ""},
				client: NewClientProvisioner(nil, &MockBlockStorageClient{}, nil),
				util:   &csi_util.Util{Logger: logging.Logger().Sugar()},
			}}
			got, err := d.ControllerUnpublishVolume(ctx, tt.args.req)
			if tt.wantErr == nil && err != nil {
				t.Errorf("got error %q, want none", err)
			}
			if tt.wantErr != nil && err == nil {
				t.Errorf("want error %q, got none", tt.wantErr)
			} else if tt.wantErr != nil && !strings.Contains(err.Error(), tt.wantErr.Error()) {
				t.Errorf("want error %q to include %q", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ControllerDriver.ControllerUnpublish() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestControllerDriver_ControllerExpandVolume(t *testing.T) {
	type fields struct {
		KubeClient kubernetes.Interface
		logger     *zap.SugaredLogger
		config     *providercfg.Config
		client     client.Interface
		util       *csi_util.Util
	}
	type args struct {
		ctx context.Context
		req *csi.ControllerExpandVolumeRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *csi.ControllerExpandVolumeResponse
		wantErr error
	}{
		{
			name:   "Error for volume OCID missing in controller expand volume",
			fields: fields{},
			args: args{
				ctx: nil,
				req: &csi.ControllerExpandVolumeRequest{
					VolumeId: "",
				},
			},
			want:    nil,
			wantErr: errors.New("UpdateVolume volumeId must be provided"),
		},
		{
			name:   "Error for invalid capacity range in ControllerExpandVolume",
			fields: fields{},
			args: args{
				ctx: nil,
				req: &csi.ControllerExpandVolumeRequest{
					VolumeId: "oc1.volume1.xxxx",
					CapacityRange: &csi.CapacityRange{
						RequiredBytes: int64(csi_util.MaximumVolumeSizeInBytes) + int64(1024),
						LimitBytes:    csi_util.MinimumVolumeSizeInBytes,
					},
				},
			},
			want:    nil,
			wantErr: errors.New("invalid capacity range"),
		},
		{
			name:   "Error for invalid Volume ID in ControllerExpandVolume",
			fields: fields{},
			args: args{
				ctx: nil,
				req: &csi.ControllerExpandVolumeRequest{
					VolumeId: "invalid_volume_id",
					CapacityRange: &csi.CapacityRange{
						RequiredBytes: int64(csi_util.MaximumVolumeSizeInBytes),
					},
				},
			},
			want:    nil,
			wantErr: errors.New("failed to check existence of volume"),
		},

		{
			name:   "Error for update Volume fail for ControllerExpandVolume",
			fields: fields{},
			args: args{
				ctx: nil,
				req: &csi.ControllerExpandVolumeRequest{
					VolumeId: "valid_volume_id_valid_old_size_fail",
					CapacityRange: &csi.CapacityRange{
						RequiredBytes: int64(csi_util.MaximumVolumeSizeInBytes),
					},
				},
			},
			want:    nil,
			wantErr: errors.New("Update volume failed"),
		},
		{
			name:   "Uhp volume expand success in ControllerExpandVolume",
			fields: fields{},
			args: args{
				ctx: nil,
				req: &csi.ControllerExpandVolumeRequest{
					VolumeId: "uhp_volume_id",
					CapacityRange: &csi.CapacityRange{
						RequiredBytes: int64(csi_util.MaximumVolumeSizeInBytes),
					},
				},
			},
			want: &csi.ControllerExpandVolumeResponse{
				CapacityBytes:         int64(csi_util.MaximumVolumeSizeInBytes),
				NodeExpansionRequired: true,
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &BlockVolumeControllerDriver{ControllerDriver{
				KubeClient: nil,
				logger:     zap.S(),
				config:     &providercfg.Config{CompartmentID: ""},
				client:     NewClientProvisioner(nil, &MockBlockStorageClient{}, nil),
				util:       &csi_util.Util{},
			}}
			got, err := d.ControllerExpandVolume(tt.args.ctx, tt.args.req)
			if tt.wantErr == nil && err != nil {
				t.Errorf("got error %q, want none", err)
			}
			if tt.wantErr != nil && err == nil {
				t.Errorf("want error %q, got none", tt.wantErr)
			} else if tt.wantErr != nil && !strings.Contains(err.Error(), tt.wantErr.Error()) {
				t.Errorf("want error %q to include %q", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ControllerDriver.ControllerExpandVolume() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_extractStorage(t *testing.T) {
	type args struct {
		capRange *csi.CapacityRange
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name:    "Nil CapacityRange",
			args:    args{capRange: nil},
			want:    testMinimumVolumeSizeInBytes,
			wantErr: false,
		},
		{
			name:    "Empty CapacityRange",
			args:    args{capRange: &csi.CapacityRange{}},
			want:    testMinimumVolumeSizeInBytes,
			wantErr: false,
		},
		{
			name: "Limit bytes is less than required",
			args: args{capRange: &csi.CapacityRange{
				RequiredBytes: 100 * client.GiB,
				LimitBytes:    50 * client.GiB,
			},
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "Required set and limit not set",
			args: args{capRange: &csi.CapacityRange{
				RequiredBytes: 100 * client.GiB,
			},
			},
			want:    100 * client.GiB,
			wantErr: false,
		},
		{
			name: "Required set and limit set",
			args: args{capRange: &csi.CapacityRange{
				RequiredBytes: 70 * client.GiB,
				LimitBytes:    100 * client.GiB,
			},
			},
			want:    100 * client.GiB,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := csi_util.ExtractStorage(tt.args.capRange)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractStorage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("extractStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractVolumeParameters(t *testing.T) {
	tests := map[string]struct {
		storageParameters map[string]string
		volumeParameters  VolumeParameters
		wantErr           bool
	}{
		"Wrong Attachment Type": {
			storageParameters: map[string]string{
				attachmentType: "foo",
			},
			volumeParameters: VolumeParameters{
				diskEncryptionKey:   "",
				attachmentParameter: make(map[string]string),
				vpusPerGB:           10,
			},
			wantErr: true,
		},
		"StorageClass Parameters are empty": {
			storageParameters: map[string]string{},
			volumeParameters: VolumeParameters{
				diskEncryptionKey:   "",
				attachmentParameter: make(map[string]string),
				vpusPerGB:           10,
			},
			wantErr: false,
		},
		"StorageClass with CMEK and attachment type paravirtualized": {
			storageParameters: map[string]string{
				attachmentType: attachmentTypeParavirtualized,
				kmsKey:         "foo",
			},
			volumeParameters: VolumeParameters{
				diskEncryptionKey: "foo",
				attachmentParameter: map[string]string{
					attachmentType: attachmentTypeParavirtualized,
				},
				vpusPerGB: 10,
			},
			wantErr: false,
		},
		"StorageClass with CMEK and attachment type iscsi": {
			storageParameters: map[string]string{
				attachmentType: attachmentTypeISCSI,
				kmsKey:         "bar",
			},
			volumeParameters: VolumeParameters{
				diskEncryptionKey: "bar",
				attachmentParameter: map[string]string{
					attachmentType: attachmentTypeISCSI,
				},
				vpusPerGB: 10,
			},
			wantErr: false,
		},
		"StorageClass with CMEK and attachment type IScsi(string casing is different)": {
			storageParameters: map[string]string{
				attachmentType: "IScsi",
				kmsKey:         "bar",
			},
			volumeParameters: VolumeParameters{
				diskEncryptionKey: "bar",
				attachmentParameter: map[string]string{
					attachmentType: attachmentTypeISCSI,
				},
				vpusPerGB: 10,
			},
			wantErr: false,
		},
		"StorageClass with CMEK and attachment type ParaVirtualized(string casing is different)": {
			storageParameters: map[string]string{
				attachmentType: "ParaVirtualized",
				kmsKey:         "foo",
			},
			volumeParameters: VolumeParameters{
				diskEncryptionKey: "foo",
				attachmentParameter: map[string]string{
					attachmentType: attachmentTypeParavirtualized,
				},
				vpusPerGB: 10,
			},
			wantErr: false,
		},
		"Invalid defined tags": {
			storageParameters: map[string]string{
				initialDefinedTagsOverride: "foo",
			},
			volumeParameters: VolumeParameters{
				diskEncryptionKey:   "",
				attachmentParameter: make(map[string]string),
				vpusPerGB:           10,
			},
			wantErr: true,
		},
		"Invalid freeform tags": {
			storageParameters: map[string]string{
				initialFreeformTagsOverride: "foo",
			},
			volumeParameters: VolumeParameters{
				diskEncryptionKey:   "",
				attachmentParameter: make(map[string]string),
				vpusPerGB:           10,
			},
			wantErr: true,
		},
		"With freeform tags": {
			storageParameters: map[string]string{
				initialFreeformTagsOverride: `{"foo":"bar"}`,
			},
			volumeParameters: VolumeParameters{
				diskEncryptionKey:   "",
				attachmentParameter: make(map[string]string),
				freeformTags:        map[string]string{"foo": "bar"},
				vpusPerGB:           10,
			},
			wantErr: false,
		},
		"With defined tags": {
			storageParameters: map[string]string{
				initialDefinedTagsOverride: `{"ns":{"foo":"bar"}}`,
			},
			volumeParameters: VolumeParameters{
				diskEncryptionKey:   "",
				attachmentParameter: make(map[string]string),
				definedTags:         map[string]map[string]interface{}{"ns": {"foo": "bar"}},
				vpusPerGB:           10,
			},
			wantErr: false,
		},
		"With freeform+defined tags": {
			storageParameters: map[string]string{
				initialFreeformTagsOverride: `{"foo":"bar"}`,
				initialDefinedTagsOverride:  `{"ns":{"foo":"bar"}}`,
			},
			volumeParameters: VolumeParameters{
				diskEncryptionKey:   "",
				attachmentParameter: make(map[string]string),
				freeformTags:        map[string]string{"foo": "bar"},
				definedTags:         map[string]map[string]interface{}{"ns": {"foo": "bar"}},
				vpusPerGB:           10,
			},
			wantErr: false,
		},
		"if low performance level then vpusPerGB should be 0": {
			storageParameters: map[string]string{
				csi_util.VpusPerGB: "0",
			},
			volumeParameters: VolumeParameters{
				diskEncryptionKey:   "",
				attachmentParameter: make(map[string]string),
				vpusPerGB:           0,
			},
			wantErr: false,
		},
		"if balanced performance level then vpusPerGB should be 10": {
			storageParameters: map[string]string{
				csi_util.VpusPerGB: "10",
			},
			volumeParameters: VolumeParameters{
				diskEncryptionKey:   "",
				attachmentParameter: make(map[string]string),
				vpusPerGB:           10,
			},
			wantErr: false,
		},
		"if high performance level then vpusPerGB should be 20": {
			storageParameters: map[string]string{
				csi_util.VpusPerGB: "20",
			},
			volumeParameters: VolumeParameters{
				diskEncryptionKey:   "",
				attachmentParameter: make(map[string]string),
				vpusPerGB:           20,
			},
			wantErr: false,
		},
		"if no parameters for performance level then default should be 10": {
			storageParameters: map[string]string{},
			volumeParameters: VolumeParameters{
				diskEncryptionKey:   "",
				attachmentParameter: make(map[string]string),
				vpusPerGB:           10,
			},
			wantErr: false,
		},
		"if out of range parameter for performance level then return error": {
			storageParameters: map[string]string{
				csi_util.VpusPerGB: "40",
			},
			volumeParameters: VolumeParameters{
				diskEncryptionKey:   "",
				attachmentParameter: make(map[string]string),
				vpusPerGB:           40,
			},
			wantErr: false,
		},
		"if invalid parameter for performance level then return error": {
			storageParameters: map[string]string{
				csi_util.VpusPerGB: "abc",
			},
			volumeParameters: VolumeParameters{
				diskEncryptionKey:   "",
				attachmentParameter: make(map[string]string),
				vpusPerGB:           10,
			},
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			volumeParameters, err := extractVolumeParameters(zap.S(), tt.storageParameters)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractVolumeParameters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(volumeParameters, tt.volumeParameters) {
				t.Errorf("extractStorage() = %+v, want %+v", volumeParameters, tt.volumeParameters)
			}
		})
	}
}

func TestExtractSnapshotParameters(t *testing.T) {
	tests := map[string]struct {
		inputParameters    map[string]string
		snapshotParameters SnapshotParameters
		wantErr            bool
	}{
		"Wrong Backup Type": {
			inputParameters: map[string]string{
				backupType: "foo",
			},
			snapshotParameters: SnapshotParameters{
				backupType: core.CreateVolumeBackupDetailsTypeIncremental,
			},
			wantErr: true,
		},
		"Incremental Backup Type": {
			inputParameters: map[string]string{
				backupType: "Incremental",
			},
			snapshotParameters: SnapshotParameters{
				backupType: core.CreateVolumeBackupDetailsTypeIncremental,
			},
			wantErr: false,
		},
		"Full Backup Type": {
			inputParameters: map[string]string{
				backupType: "Full",
			},
			snapshotParameters: SnapshotParameters{
				backupType: core.CreateVolumeBackupDetailsTypeFull,
			},
			wantErr: false,
		},
		"Invalid defined tags": {
			inputParameters: map[string]string{
				backupDefinedTags: "foo",
			},
			snapshotParameters: SnapshotParameters{
				backupType: core.CreateVolumeBackupDetailsTypeIncremental,
			},
			wantErr: true,
		},
		"Invalid freeform tags": {
			inputParameters: map[string]string{
				backupFreeformTags: "foo",
			},
			snapshotParameters: SnapshotParameters{
				backupType: core.CreateVolumeBackupDetailsTypeIncremental,
			},
			wantErr: true,
		},
		"With freeform tags": {
			inputParameters: map[string]string{
				backupFreeformTags: `{"foo":"bar"}`,
			},
			snapshotParameters: SnapshotParameters{
				backupType:   core.CreateVolumeBackupDetailsTypeIncremental,
				freeformTags: map[string]string{"foo": "bar"},
			},
			wantErr: false,
		},
		"With defined tags": {
			inputParameters: map[string]string{
				backupDefinedTags: `{"ns":{"foo":"bar"}}`,
			},
			snapshotParameters: SnapshotParameters{
				backupType:  core.CreateVolumeBackupDetailsTypeIncremental,
				definedTags: map[string]map[string]interface{}{"ns": {"foo": "bar"}},
			},
			wantErr: false,
		},
		"With freeform+defined tags": {
			inputParameters: map[string]string{
				backupFreeformTags: `{"foo":"bar"}`,
				backupDefinedTags:  `{"ns":{"foo":"bar"}}`,
			},
			snapshotParameters: SnapshotParameters{
				backupType:   core.CreateVolumeBackupDetailsTypeIncremental,
				freeformTags: map[string]string{"foo": "bar"},
				definedTags:  map[string]map[string]interface{}{"ns": {"foo": "bar"}},
			},
			wantErr: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			snapshotParameters, err := extractSnapshotParameters(tt.inputParameters)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractSnapshotParameters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(snapshotParameters, tt.snapshotParameters) {
				t.Errorf("extractSnapshotParameters() = %+v, want %+v", snapshotParameters, tt.snapshotParameters)
			}
		})
	}
}

func TestCreateSnapshot(t *testing.T) {
	type args struct {
		ctx context.Context
		req *csi.CreateSnapshotRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *csi.CreateSnapshotResponse
		wantErr error
	}{
		{
			name: "Error for name not provided for creating snapshot",
			args: args{
				ctx: nil,
				req: &csi.CreateSnapshotRequest{Name: ""},
			},
			want:    nil,
			wantErr: errors.New("Volume snapshot name must be provided"),
		},
		{
			name: "Error for volume snapshot source ID not provided for creating snapshot",
			args: args{
				ctx: nil,
				req: &csi.CreateSnapshotRequest{
					Name:           "demo",
					SourceVolumeId: "",
				},
			},
			want:    nil,
			wantErr: errors.New("Volume snapshot source ID must be provided"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &BlockVolumeControllerDriver{ControllerDriver{
				KubeClient: nil,
				logger:     zap.S(),
				config:     &providercfg.Config{CompartmentID: ""},
				client:     NewClientProvisioner(nil, &MockBlockStorageClient{}, nil),
				util:       &csi_util.Util{},
			}}
			got, err := d.CreateSnapshot(tt.args.ctx, tt.args.req)
			if tt.wantErr == nil && err != nil {
				t.Errorf("got error %q, want none", err)
			}
			if tt.wantErr != nil && !strings.Contains(err.Error(), tt.wantErr.Error()) {
				t.Errorf("want error %q to include %q", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ControllerDriver.CreateSnapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestControllerDriver_DeleteSnapshot(t *testing.T) {
	type args struct {
		ctx context.Context
		req *csi.DeleteSnapshotRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *csi.DeleteSnapshotResponse
		wantErr error
	}{
		{
			name: "Error for snapshot OCID missing in delete block volume",
			args: args{
				ctx: nil,
				req: &csi.DeleteSnapshotRequest{},
			},
			want:    nil,
			wantErr: errors.New("SnapshotId must be provided"),
		},
		{
			name: "Delete volume and get empty response",
			args: args{
				ctx: context.Background(),
				req: &csi.DeleteSnapshotRequest{SnapshotId: "oc1.volumebackup1.xxxx"},
			},
			want:    &csi.DeleteSnapshotResponse{},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &BlockVolumeControllerDriver{ControllerDriver{
				KubeClient: nil,
				logger:     zap.S(),
				config:     &providercfg.Config{CompartmentID: ""},
				client:     NewClientProvisioner(nil, &MockBlockStorageClient{}, nil),
				util:       &csi_util.Util{},
			}}
			got, err := d.DeleteSnapshot(tt.args.ctx, tt.args.req)
			if tt.wantErr == nil && err != nil {
				t.Errorf("got error %q, want none", err)
			}
			if tt.wantErr != nil && !strings.Contains(err.Error(), tt.wantErr.Error()) {
				t.Errorf("want error %q to include %q", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ControllerDriver.DeleteVolume() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetAttachmentOptions(t *testing.T) {
	tests := map[string]struct {
		attachmentType         string
		instanceID             string
		volumeAttachmentOption VolumeAttachmentOption
		wantErr                bool
	}{
		"PV attachment with instance in-transit encryption enabled": {
			attachmentType: attachmentTypeParavirtualized,
			instanceID:     "inTransitEnabled",
			volumeAttachmentOption: VolumeAttachmentOption{
				enableInTransitEncryption:    true,
				useParavirtualizedAttachment: true,
			},
			wantErr: false,
		},

		"PV attachment with instance in-transit encryption disabled": {
			attachmentType: attachmentTypeParavirtualized,
			instanceID:     "inTransitDisabled",
			volumeAttachmentOption: VolumeAttachmentOption{
				enableInTransitEncryption:    false,
				useParavirtualizedAttachment: true,
			},
			wantErr: false,
		},
		"ISCSI attachment with instance in-transit encryption enabled": {
			attachmentType: attachmentTypeISCSI,
			instanceID:     "inTransitEnabled",
			volumeAttachmentOption: VolumeAttachmentOption{
				enableInTransitEncryption:    true,
				useParavirtualizedAttachment: false,
			},
			wantErr: false,
		},
		"ISCSI attachment with instance in-transit encryption disabled": {
			attachmentType: attachmentTypeISCSI,
			instanceID:     "inTransitDisabled",
			volumeAttachmentOption: VolumeAttachmentOption{
				enableInTransitEncryption:    false,
				useParavirtualizedAttachment: false,
			},
			wantErr: false,
		},
		"API error": {
			attachmentType:         attachmentTypeISCSI,
			instanceID:             "foo",
			volumeAttachmentOption: VolumeAttachmentOption{},
			wantErr:                true,
		},
	}

	computeClient := MockOCIClient{}.Compute()

	for name, tt := range tests {

		t.Run(name, func(t *testing.T) {
			volumeAttachmentOption, err := getAttachmentOptions(context.Background(), computeClient, tt.attachmentType, tt.instanceID)
			if (err != nil) != tt.wantErr {
				t.Errorf("getAttachmentOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(volumeAttachmentOption, tt.volumeAttachmentOption) {
				t.Errorf("getAttachmentOptions() = %v, want %v", volumeAttachmentOption, tt.volumeAttachmentOption)
			}
		})
	}
}
