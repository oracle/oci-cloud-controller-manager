// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Streaming Service API
//
// The API for the Streaming Service.
//

package streaming

import (
	"github.com/oracle/oci-go-sdk/common"
)

// KafkaConnectionSettings Connection settings to use Kafka compat API.
type KafkaConnectionSettings struct {

	// Admin bootstrap servers.
	AdminBootstrapServers *string `mandatory:"false" json:"adminBootstrapServers"`

	// Bootstrap servers.
	BootstrapServers *string `mandatory:"false" json:"bootstrapServers"`
}

func (m KafkaConnectionSettings) String() string {
	return common.PointerString(m)
}
