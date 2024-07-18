package util

import (
	"context"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/oracle/oci-go-sdk/v65/filestorage"
)

type MockOCIBlockStorageClient struct {
	client core.BlockstorageClient
}

type MockOCIComputeClient struct {
	client core.ComputeClient
}

type MockOCIFileStorageClient struct {
	client filestorage.FileStorageClient
}

func (c MockOCIFileStorageClient) ListMountTargets(ctx context.Context, request filestorage.ListMountTargetsRequest) (response filestorage.ListMountTargetsResponse, err error) {
	if *request.DisplayName == "mount-target-idempotency-check-timeout-volume" {
		select {
		// from retry.go
		case <-ctx.Done():
			return response, ctx.Err()
		default:
			return filestorage.ListMountTargetsResponse{
				Items: []filestorage.MountTargetSummary{
					{
						LifecycleState: filestorage.MountTargetSummaryLifecycleStateActive,
					},
				},
				OpcNextPage: common.String("a"),
			}, nil
		}
	}
	return filestorage.ListMountTargetsResponse{}, nil
}

func (c MockOCIFileStorageClient) ListFileSystems(ctx context.Context, request filestorage.ListFileSystemsRequest) (response filestorage.ListFileSystemsResponse, err error) {
	if *request.DisplayName == "file-system-idempotency-check-timeout-volume" {
		select {
		// from retry.go
		case <-ctx.Done():
			return response, ctx.Err()
		default:
			return filestorage.ListFileSystemsResponse{
				Items: []filestorage.FileSystemSummary{
					{
						LifecycleState: filestorage.FileSystemSummaryLifecycleStateActive,
					},
				},
				OpcNextPage: common.String("a"),
			}, nil
		}
	}
	return filestorage.ListFileSystemsResponse{}, nil
}

func (c MockOCIFileStorageClient) ListExports(ctx context.Context, request filestorage.ListExportsRequest) (response filestorage.ListExportsResponse, err error) {
	if *request.FileSystemId == "export-idempotency-check-timeout" {
		select {
		// from retry.go
		case <-ctx.Done():
			return response, ctx.Err()
		default:
			return filestorage.ListExportsResponse{
				Items:       []filestorage.ExportSummary{},
				OpcNextPage: common.String("a"),
			}, nil
		}
	}
	return filestorage.ListExportsResponse{}, nil
}

func (c MockOCIComputeClient) ListVolumeAttachments(ctx context.Context, request core.ListVolumeAttachmentsRequest) (response core.ListVolumeAttachmentsResponse, err error) {
	if *request.VolumeId == "find-active-volume-attachment-timeout-volume" || *request.VolumeId == "find-volume-attachment-timeout" {
		select {
		// from retry.go
		case <-ctx.Done():
			return response, ctx.Err()
		default:
			return core.ListVolumeAttachmentsResponse{
				Items: []core.VolumeAttachment{
					core.IScsiVolumeAttachment{
						LifecycleState: core.VolumeAttachmentLifecycleStateDetached,
					},
				},
				OpcNextPage: common.String("a"),
			}, nil
		}
	}
	return core.ListVolumeAttachmentsResponse{}, nil
}

func (c MockOCIComputeClient) GetVolumeAttachment(ctx context.Context, request core.GetVolumeAttachmentRequest) (response core.GetVolumeAttachmentResponse, err error) {
	if *request.VolumeAttachmentId == "volume-attachment-stuck-in-attaching-state" {
		select {
		// from retry.go
		case <-ctx.Done():
			return response, ctx.Err()
		default:
			return core.GetVolumeAttachmentResponse{
				VolumeAttachment: core.IScsiVolumeAttachment{
					Id:             request.VolumeAttachmentId,
					LifecycleState: core.VolumeAttachmentLifecycleStateAttaching,
				},
			}, nil
		}
	}
	return core.GetVolumeAttachmentResponse{}, nil
}

func (c *MockOCIBlockStorageClient) ListVolumes(ctx context.Context, request core.ListVolumesRequest) (response core.ListVolumesResponse, err error) {
	if *request.DisplayName == "get-volumes-by-name-timeout-volume" {
		select {
		// from retry.go
		case <-ctx.Done():
			return response, ctx.Err()
		default:
			return core.ListVolumesResponse{
				Items: []core.Volume{
					{
						LifecycleState: core.VolumeLifecycleStateAvailable,
					},
				},
				OpcNextPage: common.String("a"),
			}, nil
		}
	}
	return core.ListVolumesResponse{}, nil
}
