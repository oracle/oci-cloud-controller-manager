package integtest

import (
	"context"
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/storagegateway"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStorageGateway_Create(t *testing.T) {

	p := configurationProvider()
	cId, _ := p.TenancyOCID()
	c, _ := storagegateway.NewStorageGatewayClientWithConfigurationProvider(p)

	res, err := c.CreateStorageGateway(context.Background(), storagegateway.CreateStorageGatewayRequest{
		CreateStorageGatewayDetails: storagegateway.CreateStorageGatewayDetails{
			CompartmentId: common.String(cId),
			DisplayName: common.String("TestStorage"),
			Description: common.String("some sample des"),
		},
	})

	assert.NoError(t, err)
	defer deleteStorageGateway(res.Id, t)
}

func deleteStorageGateway(id *string, t *testing.T) {
	p := configurationProvider()
	c, _ := storagegateway.NewStorageGatewayClientWithConfigurationProvider(p)

	_, err := c.DeleteStorageGateway(context.Background(), storagegateway.DeleteStorageGatewayRequest{
		StorageGatewayId: id,
	})

	assert.NoError(t, err)

}
