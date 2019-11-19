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

// UpdateDefaultStreamPoolDetails Object used to update the default stream pool's details.
type UpdateDefaultStreamPoolDetails struct {

	// Enable auto creation of topic on the server
	KafkaAutoCreateTopicsEnable *bool `mandatory:"false" json:"kafkaAutoCreateTopicsEnable"`

	// The number of hours to keep a log file before deleting it (in hours)
	KafkaLogRetentionHours *int `mandatory:"false" json:"kafkaLogRetentionHours"`

	// The default number of log partitions per topic
	KafkaNumPartitions *int `mandatory:"false" json:"kafkaNumPartitions"`
}

func (m UpdateDefaultStreamPoolDetails) String() string {
	return common.PointerString(m)
}
