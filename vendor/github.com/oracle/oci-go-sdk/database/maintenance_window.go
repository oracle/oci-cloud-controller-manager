// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Database Service API
//
// The API for the Database Service.
//

package database

import (
	"github.com/oracle/oci-go-sdk/common"
)

// MaintenanceWindow The scheduling details for the quarterly maintenance window. Patching and system updates take place during the maintenance window.
type MaintenanceWindow struct {

	// Months during the year when maintenance should be performed.
	Months []MaintenanceWindowMonthsEnum `mandatory:"false" json:"months,omitempty"`

	// Weeks during the month when maintenance should be performed. Weeks start on the 1st, 8th, 15th, and 22nd days of the month, and have a duration of 7 days. Weeks start and end based on calendar dates, not days of the week.
	// For example, to allow maintenance during the 2nd week of the month (from the 8th day to the 14th day of the month), use the value 2. Maintenance cannot be scheduled for the fifth week of months that contain more than 28 days.
	// Note that this parameter works in conjunction with the  daysOfWeek and hoursOfDay parameters to allow you to specify specific days of the week and hours that maintenance will be performed.
	WeeksOfMonth []int `mandatory:"false" json:"weeksOfMonth"`

	// Days during the week when maintenance should be performed.
	DaysOfWeek []MaintenanceWindowDaysOfWeekEnum `mandatory:"false" json:"daysOfWeek,omitempty"`

	// The window of hours during the day when maintenance should be performed.
	HoursOfDay []int `mandatory:"false" json:"hoursOfDay"`
}

func (m MaintenanceWindow) String() string {
	return common.PointerString(m)
}

// MaintenanceWindowMonthsEnum Enum with underlying type: string
type MaintenanceWindowMonthsEnum string

// Set of constants representing the allowable values for MaintenanceWindowMonthsEnum
const (
	MaintenanceWindowMonthsJanuary   MaintenanceWindowMonthsEnum = "JANUARY"
	MaintenanceWindowMonthsFebruary  MaintenanceWindowMonthsEnum = "FEBRUARY"
	MaintenanceWindowMonthsMarch     MaintenanceWindowMonthsEnum = "MARCH"
	MaintenanceWindowMonthsApril     MaintenanceWindowMonthsEnum = "APRIL"
	MaintenanceWindowMonthsMay       MaintenanceWindowMonthsEnum = "MAY"
	MaintenanceWindowMonthsJune      MaintenanceWindowMonthsEnum = "JUNE"
	MaintenanceWindowMonthsJuly      MaintenanceWindowMonthsEnum = "JULY"
	MaintenanceWindowMonthsAugust    MaintenanceWindowMonthsEnum = "AUGUST"
	MaintenanceWindowMonthsSeptember MaintenanceWindowMonthsEnum = "SEPTEMBER"
	MaintenanceWindowMonthsOctober   MaintenanceWindowMonthsEnum = "OCTOBER"
	MaintenanceWindowMonthsNovember  MaintenanceWindowMonthsEnum = "NOVEMBER"
	MaintenanceWindowMonthsDecember  MaintenanceWindowMonthsEnum = "DECEMBER"
)

var mappingMaintenanceWindowMonths = map[string]MaintenanceWindowMonthsEnum{
	"JANUARY":   MaintenanceWindowMonthsJanuary,
	"FEBRUARY":  MaintenanceWindowMonthsFebruary,
	"MARCH":     MaintenanceWindowMonthsMarch,
	"APRIL":     MaintenanceWindowMonthsApril,
	"MAY":       MaintenanceWindowMonthsMay,
	"JUNE":      MaintenanceWindowMonthsJune,
	"JULY":      MaintenanceWindowMonthsJuly,
	"AUGUST":    MaintenanceWindowMonthsAugust,
	"SEPTEMBER": MaintenanceWindowMonthsSeptember,
	"OCTOBER":   MaintenanceWindowMonthsOctober,
	"NOVEMBER":  MaintenanceWindowMonthsNovember,
	"DECEMBER":  MaintenanceWindowMonthsDecember,
}

// GetMaintenanceWindowMonthsEnumValues Enumerates the set of values for MaintenanceWindowMonthsEnum
func GetMaintenanceWindowMonthsEnumValues() []MaintenanceWindowMonthsEnum {
	values := make([]MaintenanceWindowMonthsEnum, 0)
	for _, v := range mappingMaintenanceWindowMonths {
		values = append(values, v)
	}
	return values
}

// MaintenanceWindowDaysOfWeekEnum Enum with underlying type: string
type MaintenanceWindowDaysOfWeekEnum string

// Set of constants representing the allowable values for MaintenanceWindowDaysOfWeekEnum
const (
	MaintenanceWindowDaysOfWeekMonday    MaintenanceWindowDaysOfWeekEnum = "MONDAY"
	MaintenanceWindowDaysOfWeekTuesday   MaintenanceWindowDaysOfWeekEnum = "TUESDAY"
	MaintenanceWindowDaysOfWeekWednesday MaintenanceWindowDaysOfWeekEnum = "WEDNESDAY"
	MaintenanceWindowDaysOfWeekThursday  MaintenanceWindowDaysOfWeekEnum = "THURSDAY"
	MaintenanceWindowDaysOfWeekFriday    MaintenanceWindowDaysOfWeekEnum = "FRIDAY"
	MaintenanceWindowDaysOfWeekSaturday  MaintenanceWindowDaysOfWeekEnum = "SATURDAY"
	MaintenanceWindowDaysOfWeekSunday    MaintenanceWindowDaysOfWeekEnum = "SUNDAY"
)

var mappingMaintenanceWindowDaysOfWeek = map[string]MaintenanceWindowDaysOfWeekEnum{
	"MONDAY":    MaintenanceWindowDaysOfWeekMonday,
	"TUESDAY":   MaintenanceWindowDaysOfWeekTuesday,
	"WEDNESDAY": MaintenanceWindowDaysOfWeekWednesday,
	"THURSDAY":  MaintenanceWindowDaysOfWeekThursday,
	"FRIDAY":    MaintenanceWindowDaysOfWeekFriday,
	"SATURDAY":  MaintenanceWindowDaysOfWeekSaturday,
	"SUNDAY":    MaintenanceWindowDaysOfWeekSunday,
}

// GetMaintenanceWindowDaysOfWeekEnumValues Enumerates the set of values for MaintenanceWindowDaysOfWeekEnum
func GetMaintenanceWindowDaysOfWeekEnumValues() []MaintenanceWindowDaysOfWeekEnum {
	values := make([]MaintenanceWindowDaysOfWeekEnum, 0)
	for _, v := range mappingMaintenanceWindowDaysOfWeek {
		values = append(values, v)
	}
	return values
}
