// Copyright (c) 2016, 2018, 2021, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Core Services API
//
// API covering the Networking (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/overview.htm),
// Compute (https://docs.cloud.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm), and
// Block Volume (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/overview.htm) services. Use this API
// to manage resources such as virtual cloud networks (VCNs), compute instances, and
// block storage volumes.
//

package core

import (
	"github.com/oracle/oci-go-sdk/v49/common"
)

// IpSecConnectionMigrationStatus The IPSec connection's migration status.
type IpSecConnectionMigrationStatus struct {

	// The IPSec connection's migration status.
	MigrationStatus IpSecConnectionMigrationStatusMigrationStatusEnum `mandatory:"true" json:"migrationStatus"`

	// The start timestamp for Site-to-Site VPN migration work.
	StartTimeStamp *common.SDKTime `mandatory:"true" json:"startTimeStamp"`
}

func (m IpSecConnectionMigrationStatus) String() string {
	return common.PointerString(m)
}

// IpSecConnectionMigrationStatusMigrationStatusEnum Enum with underlying type: string
type IpSecConnectionMigrationStatusMigrationStatusEnum string

// Set of constants representing the allowable values for IpSecConnectionMigrationStatusMigrationStatusEnum
const (
	IpSecConnectionMigrationStatusMigrationStatusReady           IpSecConnectionMigrationStatusMigrationStatusEnum = "READY"
	IpSecConnectionMigrationStatusMigrationStatusMigrated        IpSecConnectionMigrationStatusMigrationStatusEnum = "MIGRATED"
	IpSecConnectionMigrationStatusMigrationStatusMigrating       IpSecConnectionMigrationStatusMigrationStatusEnum = "MIGRATING"
	IpSecConnectionMigrationStatusMigrationStatusMigrationFailed IpSecConnectionMigrationStatusMigrationStatusEnum = "MIGRATION_FAILED"
	IpSecConnectionMigrationStatusMigrationStatusRolledBack      IpSecConnectionMigrationStatusMigrationStatusEnum = "ROLLED_BACK"
	IpSecConnectionMigrationStatusMigrationStatusRollingBack     IpSecConnectionMigrationStatusMigrationStatusEnum = "ROLLING_BACK"
	IpSecConnectionMigrationStatusMigrationStatusRollbackFailed  IpSecConnectionMigrationStatusMigrationStatusEnum = "ROLLBACK_FAILED"
	IpSecConnectionMigrationStatusMigrationStatusNotApplicable   IpSecConnectionMigrationStatusMigrationStatusEnum = "NOT_APPLICABLE"
	IpSecConnectionMigrationStatusMigrationStatusManual          IpSecConnectionMigrationStatusMigrationStatusEnum = "MANUAL"
	IpSecConnectionMigrationStatusMigrationStatusValidating      IpSecConnectionMigrationStatusMigrationStatusEnum = "VALIDATING"
)

var mappingIpSecConnectionMigrationStatusMigrationStatus = map[string]IpSecConnectionMigrationStatusMigrationStatusEnum{
	"READY":            IpSecConnectionMigrationStatusMigrationStatusReady,
	"MIGRATED":         IpSecConnectionMigrationStatusMigrationStatusMigrated,
	"MIGRATING":        IpSecConnectionMigrationStatusMigrationStatusMigrating,
	"MIGRATION_FAILED": IpSecConnectionMigrationStatusMigrationStatusMigrationFailed,
	"ROLLED_BACK":      IpSecConnectionMigrationStatusMigrationStatusRolledBack,
	"ROLLING_BACK":     IpSecConnectionMigrationStatusMigrationStatusRollingBack,
	"ROLLBACK_FAILED":  IpSecConnectionMigrationStatusMigrationStatusRollbackFailed,
	"NOT_APPLICABLE":   IpSecConnectionMigrationStatusMigrationStatusNotApplicable,
	"MANUAL":           IpSecConnectionMigrationStatusMigrationStatusManual,
	"VALIDATING":       IpSecConnectionMigrationStatusMigrationStatusValidating,
}

// GetIpSecConnectionMigrationStatusMigrationStatusEnumValues Enumerates the set of values for IpSecConnectionMigrationStatusMigrationStatusEnum
func GetIpSecConnectionMigrationStatusMigrationStatusEnumValues() []IpSecConnectionMigrationStatusMigrationStatusEnum {
	values := make([]IpSecConnectionMigrationStatusMigrationStatusEnum, 0)
	for _, v := range mappingIpSecConnectionMigrationStatusMigrationStatus {
		values = append(values, v)
	}
	return values
}
