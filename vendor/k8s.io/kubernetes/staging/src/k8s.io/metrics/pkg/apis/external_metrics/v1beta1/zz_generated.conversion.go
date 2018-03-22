// +build !ignore_autogenerated

/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by conversion-gen. DO NOT EDIT.

package v1beta1

import (
	unsafe "unsafe"

	conversion "k8s.io/apimachinery/pkg/conversion"
	runtime "k8s.io/apimachinery/pkg/runtime"
	external_metrics "k8s.io/metrics/pkg/apis/external_metrics"
)

func init() {
	localSchemeBuilder.Register(RegisterConversions)
}

// RegisterConversions adds conversion functions to the given scheme.
// Public to allow building arbitrary schemes.
func RegisterConversions(scheme *runtime.Scheme) error {
	return scheme.AddGeneratedConversionFuncs(
		Convert_v1beta1_ExternalMetricValue_To_external_metrics_ExternalMetricValue,
		Convert_external_metrics_ExternalMetricValue_To_v1beta1_ExternalMetricValue,
		Convert_v1beta1_ExternalMetricValueList_To_external_metrics_ExternalMetricValueList,
		Convert_external_metrics_ExternalMetricValueList_To_v1beta1_ExternalMetricValueList,
	)
}

func autoConvert_v1beta1_ExternalMetricValue_To_external_metrics_ExternalMetricValue(in *ExternalMetricValue, out *external_metrics.ExternalMetricValue, s conversion.Scope) error {
	out.MetricName = in.MetricName
	out.MetricLabels = *(*map[string]string)(unsafe.Pointer(&in.MetricLabels))
	out.Timestamp = in.Timestamp
	out.WindowSeconds = (*int64)(unsafe.Pointer(in.WindowSeconds))
	out.Value = in.Value
	return nil
}

// Convert_v1beta1_ExternalMetricValue_To_external_metrics_ExternalMetricValue is an autogenerated conversion function.
func Convert_v1beta1_ExternalMetricValue_To_external_metrics_ExternalMetricValue(in *ExternalMetricValue, out *external_metrics.ExternalMetricValue, s conversion.Scope) error {
	return autoConvert_v1beta1_ExternalMetricValue_To_external_metrics_ExternalMetricValue(in, out, s)
}

func autoConvert_external_metrics_ExternalMetricValue_To_v1beta1_ExternalMetricValue(in *external_metrics.ExternalMetricValue, out *ExternalMetricValue, s conversion.Scope) error {
	out.MetricName = in.MetricName
	out.MetricLabels = *(*map[string]string)(unsafe.Pointer(&in.MetricLabels))
	out.Timestamp = in.Timestamp
	out.WindowSeconds = (*int64)(unsafe.Pointer(in.WindowSeconds))
	out.Value = in.Value
	return nil
}

// Convert_external_metrics_ExternalMetricValue_To_v1beta1_ExternalMetricValue is an autogenerated conversion function.
func Convert_external_metrics_ExternalMetricValue_To_v1beta1_ExternalMetricValue(in *external_metrics.ExternalMetricValue, out *ExternalMetricValue, s conversion.Scope) error {
	return autoConvert_external_metrics_ExternalMetricValue_To_v1beta1_ExternalMetricValue(in, out, s)
}

func autoConvert_v1beta1_ExternalMetricValueList_To_external_metrics_ExternalMetricValueList(in *ExternalMetricValueList, out *external_metrics.ExternalMetricValueList, s conversion.Scope) error {
	out.ListMeta = in.ListMeta
	out.Items = *(*[]external_metrics.ExternalMetricValue)(unsafe.Pointer(&in.Items))
	return nil
}

// Convert_v1beta1_ExternalMetricValueList_To_external_metrics_ExternalMetricValueList is an autogenerated conversion function.
func Convert_v1beta1_ExternalMetricValueList_To_external_metrics_ExternalMetricValueList(in *ExternalMetricValueList, out *external_metrics.ExternalMetricValueList, s conversion.Scope) error {
	return autoConvert_v1beta1_ExternalMetricValueList_To_external_metrics_ExternalMetricValueList(in, out, s)
}

func autoConvert_external_metrics_ExternalMetricValueList_To_v1beta1_ExternalMetricValueList(in *external_metrics.ExternalMetricValueList, out *ExternalMetricValueList, s conversion.Scope) error {
	out.ListMeta = in.ListMeta
	out.Items = *(*[]ExternalMetricValue)(unsafe.Pointer(&in.Items))
	return nil
}

// Convert_external_metrics_ExternalMetricValueList_To_v1beta1_ExternalMetricValueList is an autogenerated conversion function.
func Convert_external_metrics_ExternalMetricValueList_To_v1beta1_ExternalMetricValueList(in *external_metrics.ExternalMetricValueList, out *ExternalMetricValueList, s conversion.Scope) error {
	return autoConvert_external_metrics_ExternalMetricValueList_To_v1beta1_ExternalMetricValueList(in, out, s)
}
