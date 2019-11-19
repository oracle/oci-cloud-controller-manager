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

// DefaultStreamPool The default stream pool for a compartment.
type DefaultStreamPool struct {

	// The OCID of the default stream pool.
	Id *string `mandatory:"true" json:"id"`

	// The handle used in the kafka compatibility username.
	KafkaHandle *string `mandatory:"true" json:"kafkaHandle"`

	// Compartment OCID that the pool belongs to.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// Enable auto creation of topic on the server.
	KafkaAutoCreateTopicsEnable *bool `mandatory:"true" json:"kafkaAutoCreateTopicsEnable"`

	// The number of hours to keep a log file before deleting it (in hours).
	KafkaLogRetentionHours *int `mandatory:"true" json:"kafkaLogRetentionHours"`

	// The default number of log partitions per topic.
	KafkaNumPartitions *int `mandatory:"true" json:"kafkaNumPartitions"`

	KafkaConnectionSettings *KafkaConnectionSettings `mandatory:"false" json:"kafkaConnectionSettings"`
}

func (m DefaultStreamPool) String() string {
	return common.PointerString(m)
}
