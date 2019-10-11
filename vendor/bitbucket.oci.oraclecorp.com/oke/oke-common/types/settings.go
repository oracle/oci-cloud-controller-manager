package types

import (
	"time"

	"bitbucket.oci.oraclecorp.com/oke/oke-common/protobuf"
)

// SettingV1 stores information about a setting
type SettingV1 struct {
	Name      string    `json:"name"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ToV1 converts a Setting object to a SettingV1 object understood by the higher layers
func (src *Setting) ToV1() *SettingV1 {
	var dst SettingV1
	if src == nil {
		return &dst
	}

	dst.Name = src.Name
	dst.Value = src.Value
	dst.UpdatedAt = protobuf.ToTime(src.UpdatedAt).Truncate(time.Second)
	return &dst
}

// ToV1 converts a SettingsGetResponse object to a SettingV1 object understood by the higher layers
func (src *SettingsGetResponse) ToV1() SettingV1 {
	dst := SettingV1{}

	if src != nil {
		dst = *src.Setting.ToV1()
	}
	return dst
}

// SettingsListResponseV1 are a collection that stores settings information
type SettingsListResponseV1 struct {
	Settings []*SettingV1 `json:"settings"`
}

// ToV1 converts a SettingsListResponse object to a SettingsListResponseV1 object understood by the higher layers
func (src *SettingsListResponse) ToV1() SettingsListResponseV1 {
	v1 := SettingsListResponseV1{
		Settings: make([]*SettingV1, 0),
	}

	if src != nil {
		for _, setting := range src.Settings {
			v1.Settings = append(v1.Settings, setting.ToV1())
		}
	}
	return v1
}

// SettingNewRequestV1 is the request to create a new setting
type SettingNewRequestV1 struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ToProto converts a SettingNewRequestV1 to a K8Instance object understood by grpc
func (v1 *SettingNewRequestV1) ToProto() *SettingsNewRequest {
	var dst SettingsNewRequest
	if v1 != nil {
		dst.Name = v1.Name
		dst.Value = v1.Value
	}

	return &dst
}

// ToV1 converts a SettingsGetResponse object to a SettingV1 object understood by the higher layers
func (src *SettingsNewResponse) ToV1() SettingV1 {
	dst := SettingV1{}

	if src != nil {
		dst.Name = src.Name
		dst.Value = src.Value
		dst.UpdatedAt = protobuf.ToTime(src.UpdatedAt).Truncate(time.Second)
	}
	return dst
}

// SettingsDeleteRequestV1 is the request to delete a setting
type SettingsDeleteRequestV1 struct {
	Name string `json:"name"`
}

// ToProto converts a SettingsDeleteRequestV1 to a K8Instance object understood by grpc
func (v1 *SettingsDeleteRequestV1) ToProto() *SettingsDeleteRequest {
	var dst SettingsDeleteRequest
	if v1 != nil {
		dst.Name = v1.Name
	}

	return &dst
}

// SettingUpdateRequestV1 is the request to update a setting
type SettingUpdateRequestV1 struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ToProto converts a SettingUpdateRequestV1 to a SettingsUpdateRequest object understood by grpc
func (v1 *SettingUpdateRequestV1) ToProto() *SettingsUpdateRequest {
	var dst SettingsUpdateRequest
	if v1 != nil {
		dst.Name = v1.Name
		dst.Value = v1.Value
	}

	return &dst
}

// ToV1 converts a SettingsUpdateResponse object to a SettingV1 object understood by the higher layers
func (src *SettingsUpdateResponse) ToV1() SettingV1 {
	dst := SettingV1{}

	if src != nil {
		dst.Name = src.Name
		dst.Value = src.Value
		dst.UpdatedAt = protobuf.ToTime(src.UpdatedAt).Truncate(time.Second)
	}
	return dst
}
