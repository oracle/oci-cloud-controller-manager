// Copyright 2018 Oracle and/or its affiliates. All rights reserved.
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

package oci

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"testing"

	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/oracle/oci-go-sdk/v65/filestorage"
	"github.com/oracle/oci-go-sdk/v65/identity"

	"github.com/oracle/oci-go-sdk/v65/loadbalancer"
	"go.uber.org/zap"
	authv1 "k8s.io/api/authentication/v1"
	v1 "k8s.io/api/core/v1"
	v1discovery "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	v1discoverylisters "k8s.io/client-go/listers/discovery/v1"
)

var (
	instanceVnics = map[string]*core.Vnic{
		"ocid1.default": {
			PrivateIp:     common.String("10.0.0.1"),
			PublicIp:      common.String("0.0.0.1"),
			HostnameLabel: common.String("default"),
			SubnetId:      common.String("subnetwithdnslabel"),
		},
		"ocid1.instance1": {
			PrivateIp:     common.String("10.0.0.1"),
			PublicIp:      common.String("0.0.0.1"),
			HostnameLabel: common.String("instance1"),
			SubnetId:      common.String("subnetwithdnslabel"),
		},
		"ocid1.basic-complete": {
			PrivateIp:     common.String("10.0.0.1"),
			PublicIp:      common.String("0.0.0.1"),
			HostnameLabel: common.String("basic-complete"),
			SubnetId:      common.String("subnetwithdnslabel"),
		},
		"ocid1.no-external-ip": {
			PrivateIp:     common.String("10.0.0.1"),
			HostnameLabel: common.String("no-external-ip"),
			SubnetId:      common.String("subnetwithdnslabel"),
		},
		"ocid1.no-internal-ip": {
			PublicIp:      common.String("0.0.0.1"),
			HostnameLabel: common.String("no-internal-ip"),
			SubnetId:      common.String("subnetwithdnslabel"),
		},
		"ocid1.invalid-internal-ip": {
			PrivateIp:     common.String("10.0.0."),
			HostnameLabel: common.String("no-internal-ip"),
			SubnetId:      common.String("subnetwithdnslabel"),
		},
		"ocid1.invalid-external-ip": {
			PublicIp:      common.String("0.0.0."),
			HostnameLabel: common.String("invalid-external-ip"),
			SubnetId:      common.String("subnetwithdnslabel"),
		},
		"ocid1.no-hostname-label": {
			PrivateIp: common.String("10.0.0.1"),
			PublicIp:  common.String("0.0.0.1"),
			SubnetId:  common.String("subnetwithdnslabel"),
		},
		"ocid1.no-subnet-dns-label": {
			PrivateIp:     common.String("10.0.0.1"),
			PublicIp:      common.String("0.0.0.1"),
			HostnameLabel: common.String("no-subnet-dns-label"),
			SubnetId:      common.String("subnetwithoutdnslabel"),
		},
		"ocid1.no-vcn-dns-label": {
			PrivateIp:     common.String("10.0.0.1"),
			PublicIp:      common.String("0.0.0.1"),
			HostnameLabel: common.String("no-vcn-dns-label"),
			SubnetId:      common.String("subnetwithnovcndnslabel"),
		},
		"ocid1.ipv6-instance": {
			HostnameLabel: common.String("no-vcn-dns-label"),
			SubnetId:      common.String("IPv6-subnet"),
			Ipv6Addresses: []string{"2001:0db8:85a3:0000:0000:8a2e:0370:7334"},
		},
		"ocid1.ipv6-instance-ula": {
			HostnameLabel: common.String("no-vcn-dns-label"),
			SubnetId:      common.String("IPv6-subnet"),
			Ipv6Addresses: []string{"fc00:0000:0000:0000:0000:0000:0000:0000"},
		},
		"ocid1.ipv6-instance-2": {
			PrivateIp:     common.String("10.0.0.1"),
			PublicIp:      common.String("0.0.0.1"),
			HostnameLabel: common.String("no-vcn-dns-label"),
			SubnetId:      common.String("IPv6-subnet"),
			Ipv6Addresses: []string{"2001:0db8:85a3:0000:0000:8a2e:0370:idfe"},
		},
		"ocid1.instance-id-ipv4-ipv6": {
			PrivateIp:     common.String("10.0.0.1"),
			PublicIp:      common.String("0.0.0.1"),
			HostnameLabel: common.String("no-vcn-dns-label"),
			SubnetId:      common.String("IPv4-IPv6-subnet"),
			Ipv6Addresses: []string{"2001:0db8:85a3:0000:0000:8a2e:0370:7334"},
		},
		"ocid1.instance-id-ipv6": {
			HostnameLabel: common.String("no-vcn-dns-label"),
			SubnetId:      common.String("ipv6-instance"),
			Ipv6Addresses: []string{"2001:0db8:85a3:0000:0000:8a2e:0370:7334"},
		},
		"ocid1.instance-id-ipv4": {
			PrivateIp:     common.String("10.0.0.1"),
			PublicIp:      common.String("0.0.0.1"),
			HostnameLabel: common.String("no-vcn-dns-label"),
			SubnetId:      common.String("subnetwithnovcndnslabel"),
		},
		"ocid1.ipv6-gua-ipv4-instance": {
			PrivateIp:     common.String("10.0.0.1"),
			HostnameLabel: common.String("no-vcn-dns-label"),
			SubnetId:      common.String("ipv6-gua-ipv4-instance"),
			Ipv6Addresses: []string{"2001:0db8:85a3:0000:0000:8a2e:0370:7334"},
		},
		"ocid1.openshift-instance-ipv4": {
			IsPrimary:     common.Bool(false),
			PrivateIp:     common.String("10.0.0.1"),
			PublicIp:      common.String("0.0.0.1"),
			HostnameLabel: common.String("openshift-instance"),
			SubnetId:      common.String("subnetwithdnslabel"),
		},
		"ocid1.openshift-instance-invalid": {
			PrivateIp:     common.String("10.0.0."),
			HostnameLabel: common.String("openshift-instance-invalid"),
			SubnetId:      common.String("subnetwithdnslabel"),
		},
		"ocid1.openshift-instance-ipv6": {
			IsPrimary:     common.Bool(false),
			HostnameLabel: common.String("openshift-instance"),
			SubnetId:      common.String("ipv6-instance"),
			Ipv6Addresses: []string{"2001:0db8:85a3:0000:0000:8a2e:0370:7334"},
		},
	}

	instances = map[string]*core.Instance{
		"basic-complete": {
			Id:            common.String("ocid1.basic-complete"),
			CompartmentId: common.String("default"),
		},
		"no-external-ip": {
			Id:            common.String("ocid1.no-external-ip"),
			CompartmentId: common.String("default"),
		},
		"no-internal-ip": {
			Id:            common.String("ocid1.no-internal-ip"),
			CompartmentId: common.String("default"),
		},
		"invalid-internal-ip": {
			Id:            common.String("ocid1.invalid-internal-ip"),
			CompartmentId: common.String("default"),
		},
		"invalid-external-ip": {
			Id:            common.String("ocid1.invalid-external-ip"),
			CompartmentId: common.String("default"),
		},
		"no-hostname-label": {
			Id:            common.String("ocid1.no-hostname-label"),
			CompartmentId: common.String("default"),
		},
		"no-subnet-dns-label": {
			Id:            common.String("ocid1.no-subnet-dns-label"),
			CompartmentId: common.String("default"),
		},
		"no-vcn-dns-label": {
			Id:            common.String("ocid1.no-vcn-dns-label"),
			CompartmentId: common.String("default"),
		},
		"instance1": {
			CompartmentId: common.String("compartment1"),
			Id:            common.String("ocid1.instance1"),
			Shape:         common.String("VM.Standard1.2"),
			DisplayName:   common.String("instance1"),
		},
		"instance_zone_test": {
			AvailabilityDomain: common.String("NWuj:PHX-AD-1"),
			CompartmentId:      common.String("compartment1"),
			Id:                 common.String("ocid1.instance_zone_test"),
			Region:             common.String("PHX"),
			Shape:              common.String("VM.Standard1.2"),
			DisplayName:        common.String("instance_zone_test"),
		},
		"ipv6-instance": {
			Id:            common.String("ocid1.ipv6-instance"),
			CompartmentId: common.String("ipv6-instance"),
		},
		"instance-id-ipv4-ipv6": {
			Id:            common.String("ocid1.instance-id-ipv4-ipv6"),
			CompartmentId: common.String("instance-id-ipv4-ipv6"),
		},
		"instance-id-ipv4": {
			Id:            common.String("ocid1.instance-id-ipv4"),
			CompartmentId: common.String("instance-id-ipv4"),
		},
		"instance-id-ipv6": {
			Id:            common.String("ocid1.instance-id-ipv6"),
			CompartmentId: common.String("instance-id-ipv6"),
		},
		"ipv6-gua-ipv4-instance": {
			Id:            common.String("ocid1.ipv6-gua-ipv4-instance"),
			CompartmentId: common.String("ipv6-gua-ipv4-instance"),
		},
		"openshift-instance-ipv4": {
			Id:            common.String("ocid1.openshift-instance-ipv4"),
			CompartmentId: common.String("default"),
			DefinedTags: map[string]map[string]interface{}{
				"openshift-namespace": {
					"role": "compute",
				},
			},
		},
		"openshift-instance-invalid": {
			Id:            common.String("ocid1.openshift-instance-invalid"),
			CompartmentId: common.String("default"),
			DefinedTags: map[string]map[string]interface{}{
				"openshift-namespace": {
					"role": "compute",
				},
			},
		},
		"openshift-instance-ipv6": {
			Id:            common.String("ocid1.openshift-instance-ipv6"),
			CompartmentId: common.String("default"),
			DefinedTags: map[string]map[string]interface{}{
				"openshift-namespace": {
					"role": "compute",
				},
			},
		},
	}
	subnets = map[string]*core.Subnet{
		"subnetwithdnslabel": {
			Id:       common.String("subnetwithdnslabel"),
			DnsLabel: common.String("subnetwithdnslabel"),
			VcnId:    common.String("vcnwithdnslabel"),
		},
		"subnetwithoutdnslabel": {
			Id:    common.String("subnetwithoutdnslabel"),
			VcnId: common.String("vcnwithdnslabel"),
		},
		"subnetwithnovcndnslabel": {
			Id:       common.String("subnetwithnovcndnslabel"),
			DnsLabel: common.String("subnetwithnovcndnslabel"),
			VcnId:    common.String("vcnwithoutdnslabel"),
		},
		"one": {
			Id:                 common.String("one"),
			DnsLabel:           common.String("subnetwithnovcndnslabel"),
			VcnId:              common.String("vcnwithoutdnslabel"),
			AvailabilityDomain: common.String("AD1"),
		},
		"two": {
			Id:                 common.String("two"),
			DnsLabel:           common.String("subnetwithnovcndnslabel"),
			VcnId:              common.String("vcnwithoutdnslabel"),
			AvailabilityDomain: common.String("AD2"),
		},
		"annotation-one": {
			Id:                 common.String("annotation-one"),
			DnsLabel:           common.String("subnetwithnovcndnslabel"),
			VcnId:              common.String("vcnwithoutdnslabel"),
			AvailabilityDomain: common.String("AD1"),
		},
		"annotation-two": {
			Id:                 common.String("annotation-two"),
			DnsLabel:           common.String("subnetwithnovcndnslabel"),
			VcnId:              common.String("vcnwithoutdnslabel"),
			AvailabilityDomain: common.String("AD2"),
		},
		"regional-subnet": {
			Id:                 common.String("regional-subnet"),
			DnsLabel:           common.String("subnetwithnovcndnslabel"),
			VcnId:              common.String("vcnwithoutdnslabel"),
			AvailabilityDomain: nil,
		},
		"IPv4-subnet": {
			Id:                 common.String("IPv4-subnet"),
			DnsLabel:           common.String("subnetwithnovcndnslabel"),
			VcnId:              common.String("vcnwithoutdnslabel"),
			AvailabilityDomain: nil,
			CidrBlock:          common.String("10.0.0.0/16"),
		},
		"IPv6-subnet": {
			Id:                 common.String("IPv6-subnet"),
			DnsLabel:           common.String("subnetwithnovcndnslabel"),
			VcnId:              common.String("vcnwithoutdnslabel"),
			AvailabilityDomain: nil,
			Ipv6CidrBlock:      common.String("IPv6Cidr"),
			Ipv6CidrBlocks:     []string{"IPv6Cidr"},
		},
		"IPv4-IPv6-subnet": {
			Id:                 common.String("IPv4-IPv6-subnet"),
			DnsLabel:           common.String("subnetwithnovcndnslabel"),
			VcnId:              common.String("vcnwithoutdnslabel"),
			AvailabilityDomain: nil,
			CidrBlock:          common.String("10.0.0.0/16"),
			Ipv6CidrBlocks:     []string{},
		},
		"ipv6-gua-ipv4-instance": {
			Id:                 common.String("ipv6-gua-ipv4-instance"),
			DnsLabel:           common.String("subnetwithnovcndnslabel"),
			VcnId:              common.String("vcnwithoutdnslabel"),
			AvailabilityDomain: nil,
			CidrBlock:          common.String("10.0.0.0/16"),
			Ipv6CidrBlocks:     []string{"2001:0db8:85a3::8a2e:0370:7334/64"},
		},
	}

	vcns = map[string]*core.Vcn{
		"vcnwithdnslabel": {
			Id:       common.String("vcnwithdnslabel"),
			DnsLabel: common.String("vcnwithdnslabel"),
		},
		"vcnwithoutdnslabel": {
			Id: common.String("vcnwithoutdnslabel"),
		},
	}

	nodeList = map[string]*v1.Node{
		"default": {
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					CompartmentIDAnnotation: "default",
				},
				Name: "default",
			},
			Spec: v1.NodeSpec{
				ProviderID: "ocid1.default",
			},
		},
		"instance1": {
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					CompartmentIDAnnotation: "compartment1",
				},
				Name: "instance1",
			},
			Spec: v1.NodeSpec{
				ProviderID: "ocid1.instance1",
			},
		},
		"instanceWithAddress1": {
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					CompartmentIDAnnotation: "compartment1",
				},
				Name: "instanceWithAddress1",
			},
			Spec: v1.NodeSpec{
				ProviderID: "ocid1.instanceWithAddress1",
			},
			Status: v1.NodeStatus{
				Addresses: []v1.NodeAddress{
					{
						Address: "0.0.0.0",
						Type:    "InternalIP",
					},
				},
			},
		},
		"instanceWithAddress2": {
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					CompartmentIDAnnotation: "compartment1",
				},
				Name: "instanceWithAddress2",
			},
			Spec: v1.NodeSpec{
				ProviderID: "ocid1.instanceWithAddress2",
			},
			Status: v1.NodeStatus{
				Addresses: []v1.NodeAddress{
					{
						Address: "0.0.0.1",
						Type:    "InternalIP",
					},
				},
			},
		},
		"instanceWithAddressIPv4IPv6": {
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					CompartmentIDAnnotation: "compartment1",
				},
				Name: "instanceWithAddressIPv4IPv6",
			},
			Spec: v1.NodeSpec{
				ProviderID: "ocid1.instanceWithAddressIPv4IPv6",
			},
			Status: v1.NodeStatus{
				Addresses: []v1.NodeAddress{
					{
						Address: "0.0.0.1",
						Type:    "InternalIP",
					},
					{
						Address: "2001:0db8:85a3:0000:0000:8a2e:0370:7335",
						Type:    "InternalIP",
					},
				},
			},
		},
		"ipv6-node": {
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					CompartmentIDAnnotation: "default",
				},
				Labels: map[string]string{
					IPv4NodeIPFamilyLabel: "true",
					IPv6NodeIPFamilyLabel: "true",
				},
				Name: "Node-Ipv6",
			},
			Spec: v1.NodeSpec{
				ProviderID: "ocid1.ipv6-instance",
			},
		},
		"ipv6-node-ula": {
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					CompartmentIDAnnotation: "default",
				},
				Labels: map[string]string{
					IPv4NodeIPFamilyLabel: "true",
					IPv6NodeIPFamilyLabel: "true",
				},
				Name: "Node-Ipv6",
			},
			Spec: v1.NodeSpec{
				ProviderID: "ocid1.ipv6-instance-ula",
			},
		},
		"instance-id-ipv4-ipv6": {
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					CompartmentIDAnnotation: "default",
				},
				Labels: map[string]string{
					IPv4NodeIPFamilyLabel: "true",
					IPv6NodeIPFamilyLabel: "true",
				},
				Name: "Node-Ipv6",
			},
			Spec: v1.NodeSpec{
				ProviderID: "ocid1.instance-id-ipv4-ipv6",
			},
		},
		"instance-id-ipv4": {
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					CompartmentIDAnnotation: "default",
				},
				Labels: map[string]string{
					IPv4NodeIPFamilyLabel: "true",
				},
				Name: "Node-Ipv6",
			},
			Spec: v1.NodeSpec{
				ProviderID: "ocid1.instance-id-ipv4",
			},
		},
		"instance-id-ipv6": {
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					CompartmentIDAnnotation: "default",
				},
				Labels: map[string]string{
					IPv6NodeIPFamilyLabel: "true",
				},
				Name: "Node-Ipv6",
			},
			Spec: v1.NodeSpec{
				ProviderID: "ocid1.instance-id-ipv6",
			},
		},
		"openshift-instance-id-ipv6": {
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					CompartmentIDAnnotation: "default",
				},
				Labels: map[string]string{
					IPv6NodeIPFamilyLabel: "true",
				},
				Name: "Node-Ipv6",
			},
			Spec: v1.NodeSpec{
				ProviderID: "ocid1.openshift-instance-ipv6",
			},
		},
	}

	podList = map[string]*v1.Pod{
		"virtualPod1": {
			ObjectMeta: metav1.ObjectMeta{
				Name: "virtualPod1",
				Labels: map[string]string{
					"app": "pod1",
				},
			},
			Spec: v1.PodSpec{
				NodeName: "virtualNodeDefault",
			},
		},
		"virtualPod2": {
			ObjectMeta: metav1.ObjectMeta{
				Name: "virtualPod2",
				Labels: map[string]string{
					"app": "pod2",
				},
			},
			Spec: v1.PodSpec{
				NodeName: "virtualNodeDefault",
			},
			Status: v1.PodStatus{
				PodIP: "0.0.0.10",
				PodIPs: []v1.PodIP{
					{"0.0.0.10"},
				},
			},
		},
		"virtualPodIPv4Ipv6": {
			ObjectMeta: metav1.ObjectMeta{
				Name: "virtualPodIPv4Ipv6",
				Labels: map[string]string{
					"app": "pod3",
				},
			},
			Spec: v1.PodSpec{
				NodeName: "virtualNodeDefault",
			},
			Status: v1.PodStatus{
				PodIP: "0.0.0.20",
				PodIPs: []v1.PodIP{
					{"0.0.0.20"},
					{"2001:0db8:85a3:0000:0000:8a2e:0370:7334"},
				},
			},
		},
		"regularPod1": {
			ObjectMeta: metav1.ObjectMeta{
				Name: "regularPod1",
			},
			Spec: v1.PodSpec{
				NodeName: "default",
			},
		},
		"regularPod2": {
			ObjectMeta: metav1.ObjectMeta{
				Name: "regularPod2",
			},
			Spec: v1.PodSpec{
				NodeName: "default",
			},
		},
	}

	ready             = true
	endpointSliceList = map[string]*v1discovery.EndpointSlice{
		"endpointSliceVirtual": {
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					v1discovery.LabelServiceName: "virtualService",
				},
			},
			Endpoints: []v1discovery.Endpoint{
				{
					TargetRef: &v1.ObjectReference{
						Kind: "Pod",
						Name: "virtualPod1",
					},
					Conditions: v1discovery.EndpointConditions{
						Ready: &ready,
					},
					Addresses: []string{"0.0.0.9"},
					NodeName:  common.String("virtualNodeDefault"),
				},
				{
					TargetRef: &v1.ObjectReference{
						Kind: "Pod",
						Name: "virtualPod2",
					},
					Conditions: v1discovery.EndpointConditions{
						Ready: &ready,
					},
					Addresses: []string{"0.0.0.10"},
					NodeName:  common.String("virtualNodeDefault"),
				},
			},
		},
		"endpointSliceRegular": {
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					v1discovery.LabelServiceName: "regularService",
				},
			},
			Endpoints: []v1discovery.Endpoint{
				{
					TargetRef: &v1.ObjectReference{
						Kind: "Pod",
						Name: "regularPod1",
					},
					Conditions: v1discovery.EndpointConditions{
						Ready: &ready,
					},
					Addresses: []string{"0.0.0.19"},
					NodeName:  common.String("default"),
				},
				{
					TargetRef: &v1.ObjectReference{
						Kind: "Pod",
						Name: "regularPod2",
					},
					Conditions: v1discovery.EndpointConditions{
						Ready: &ready,
					},
					Addresses: []string{"0.0.0.20"},
					NodeName:  common.String("default"),
				},
			},
		},

		"endpointSliceMixed": {
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					v1discovery.LabelServiceName: "mixedService",
				},
			},
			Endpoints: []v1discovery.Endpoint{
				{
					TargetRef: &v1.ObjectReference{
						Kind: "Pod",
						Name: "virtualPod1",
					},
					Conditions: v1discovery.EndpointConditions{
						Ready: &ready,
					},
					Addresses: []string{"0.0.0.9"},
					NodeName:  common.String("virtualNodeDefault"),
				},
				{
					TargetRef: &v1.ObjectReference{
						Kind: "Pod",
						Name: "regularPod1",
					},
					Conditions: v1discovery.EndpointConditions{
						Ready: &ready,
					},
					Addresses: []string{"0.0.0.19"},
					NodeName:  common.String("default"),
				},
				{
					TargetRef: &v1.ObjectReference{
						Kind: "Pod",
						Name: "virtualPod2",
					},
					Conditions: v1discovery.EndpointConditions{
						Ready: &ready,
					},
					Addresses: []string{"0.0.0.10"},
					NodeName:  common.String("virtualNodeDefault"),
				},
				{
					TargetRef: &v1.ObjectReference{
						Kind: "Pod",
						Name: "regularPod2",
					},
					Conditions: v1discovery.EndpointConditions{
						Ready: &ready,
					},
					Addresses: []string{"0.0.0.20"},
					NodeName:  common.String("default"),
				},
			},
		},
		"endpointSliceUnknownPod": {
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					v1discovery.LabelServiceName: "unknownService",
				},
			},
			Endpoints: []v1discovery.Endpoint{
				{
					TargetRef: &v1.ObjectReference{
						Kind: "Pod",
						Name: "unknown",
					},
					Conditions: v1discovery.EndpointConditions{
						Ready: &ready,
					},
					Addresses: []string{"0.0.0.100"},
				},
			},
		},
		"endpointSliceDuplicate1.1": {
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					v1discovery.LabelServiceName: "duplicateEndpointsService",
				},
			},
			Endpoints: []v1discovery.Endpoint{
				{
					TargetRef: &v1.ObjectReference{
						Kind: "Pod",
						Name: "virtualPod1",
					},
					Conditions: v1discovery.EndpointConditions{
						Ready: &ready,
					},
					Addresses: []string{"0.0.0.10"},
				},
				{
					TargetRef: &v1.ObjectReference{
						Kind: "Pod",
						Name: "virtualPod2",
					},
					Conditions: v1discovery.EndpointConditions{
						Ready: &ready,
					},
					Addresses: []string{"0.0.0.9"},
				},
			},
		},
		"endpointSliceDuplicate1.2": {
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					v1discovery.LabelServiceName: "duplicateEndpointsService",
				},
			},
			Endpoints: []v1discovery.Endpoint{
				{
					TargetRef: &v1.ObjectReference{
						Kind: "Pod",
						Name: "virtualPod1",
					},
					Conditions: v1discovery.EndpointConditions{
						Ready: &ready,
					},
					Addresses: []string{"0.0.0.10"},
				},
				{
					TargetRef: &v1.ObjectReference{
						Kind: "Pod",
						Name: "virtualPod2",
					},
					Conditions: v1discovery.EndpointConditions{
						Ready: &ready,
					},
					Addresses: []string{"0.0.0.9"},
				},
			},
		},
	}

	serviceList = map[string]*v1.Service{
		"default": {
			ObjectMeta: metav1.ObjectMeta{
				Name: "default",
			},
			Spec: v1.ServiceSpec{
				Type: v1.ServiceTypeLoadBalancer,
			},
		},
		"non-loadbalancer": {
			ObjectMeta: metav1.ObjectMeta{
				Name: "non-loadbalancer",
			},
		},
	}

	loadBalancers = map[string]*client.GenericLoadBalancer{
		"privateLB": {
			Id:          common.String("privateLB"),
			DisplayName: common.String("privateLB"),
			IpAddresses: []client.GenericIpAddress{
				{
					IpAddress: common.String("10.0.50.5"),
					IsPublic:  common.Bool(false),
				},
			},
		},
		"privateLB-no-IP": {
			Id:          common.String("privateLB-no-IP"),
			DisplayName: common.String("privateLB-no-IP"),
			IpAddresses: []client.GenericIpAddress{},
		},
		"test-uid": {
			Id:          common.String("test-uid"),
			DisplayName: common.String("test-uid"),
			IpAddresses: []client.GenericIpAddress{
				{
					IpAddress: common.String("10.0.50.5"),
					IsPublic:  common.Bool(false),
				},
			},
		},
		"test-uid-delete-err": {
			Id:          common.String("test-uid-delete-err"),
			DisplayName: common.String("test-uid-delete-err"),
			IpAddresses: []client.GenericIpAddress{
				{
					IpAddress: common.String("10.0.50.5"),
					IsPublic:  common.Bool(false),
				},
			},
		},
		"test-uid-node-err": {
			Id:          common.String("test-uid-delete-err"),
			DisplayName: common.String("test-uid-delete-err"),
			IpAddresses: []client.GenericIpAddress{
				{
					IpAddress: common.String("10.0.50.5"),
					IsPublic:  common.Bool(false),
				},
			},
			SubnetIds: []string{*subnets["one"].Id, *subnets["two"].Id},
			Listeners: map[string]client.GenericListener{
				"one": {
					Name:                  common.String("one"),
					DefaultBackendSetName: common.String("one"),
					Port:                  common.Int(5665),
				}},
			BackendSets: map[string]client.GenericBackendSetDetails{
				"one": {
					Backends: []client.GenericBackend{{
						Name:      common.String("one"),
						IpAddress: common.String("10.0.50.5"),
						Port:      common.Int(5665),
					}},
				}},
		},
		"lb-without-IP-address": {
			Id:          common.String("lb-without-IP-address"),
			IpAddresses: []client.GenericIpAddress{},
		},
	}
)

type MockSecurityListManager struct{}

func (MockSecurityListManager) Update(ctx context.Context, sc securityRuleComponents) error {
	return nil
}

func (MockSecurityListManager) Delete(ctx context.Context, sc securityRuleComponents) error {
	return nil
}

type MockSecurityListManagerFactory func(mode string) MockSecurityListManager

type MockOCIClient struct{}

func (MockOCIClient) Compute() client.ComputeInterface {
	return &MockComputeClient{}
}

func (MockOCIClient) LoadBalancer(logger *zap.SugaredLogger, lbType string, tenancy string, token *authv1.TokenRequest) client.GenericLoadBalancerInterface {
	if lbType == "nlb" {
		return &MockNetworkLoadBalancerClient{}
	}
	return &MockLoadBalancerClient{}
}

func (MockOCIClient) Networking(ociClientConfig *client.OCIClientConfig) client.NetworkingInterface {
	return &MockVirtualNetworkClient{}
}

func (MockOCIClient) BlockStorage() client.BlockStorageInterface {
	return &MockBlockStorageClient{}
}

func (MockOCIClient) FSS(ociClientConfig *client.OCIClientConfig) client.FileStorageInterface {
	return &MockFileStorageClient{}
}

func (MockOCIClient) Identity(ociClientConfig *client.OCIClientConfig) client.IdentityInterface {
	return &MockIdentityClient{}
}

// MockComputeClient mocks Compute client implementation
type MockComputeClient struct{}

func (c MockComputeClient) ListInstancesByCompartmentAndAD(ctx context.Context, compartmentId, availabilityDomain string) (response []core.Instance, err error) {
	return nil, nil
}

func (MockComputeClient) GetInstance(ctx context.Context, id string) (*core.Instance, error) {
	if instance, ok := instances[id]; ok {
		return instance, nil
	}
	return &core.Instance{
		AvailabilityDomain: common.String("NWuj:PHX-AD-1"),
		CompartmentId:      common.String("default"),
		Id:                 &id,
		Region:             common.String("PHX"),
		Shape:              common.String("VM.Standard1.2"),
	}, nil
}

func (MockComputeClient) GetInstanceByNodeName(ctx context.Context, compartmentID, vcnID, nodeName string) (*core.Instance, error) {
	if instance, ok := instances[nodeName]; ok {
		return instance, nil
	}
	return &core.Instance{
		AvailabilityDomain: common.String("NWuj:PHX-AD-1"),
		CompartmentId:      &compartmentID,
		Id:                 &nodeName,
		Region:             common.String("PHX"),
		Shape:              common.String("VM.Standard1.2"),
	}, nil
}

func (MockComputeClient) GetPrimaryVNICForInstance(ctx context.Context, compartmentID, instanceID string) (*core.Vnic, error) {
	return instanceVnics[instanceID], nil
}

func (MockComputeClient) GetSecondaryVNICsForInstance(ctx context.Context, compartmentID, instanceID string) ([]*core.Vnic, error) {
	return []*core.Vnic{instanceVnics[instanceID]}, nil
}
func (c *MockComputeClient) ListVnicAttachments(ctx context.Context, compartmentID, instanceID string) ([]core.VnicAttachment, error) {
	return nil, nil
}

func (c *MockComputeClient) GetVnicAttachment(ctx context.Context, vnicAttachmentId *string) (response *core.VnicAttachment, err error) {
	return nil, nil
}

func (c *MockComputeClient) AttachVnic(ctx context.Context, instanceID, subnetID *string, nsgIds []*string, skipSourceDestCheck *bool) (response core.VnicAttachment, err error) {
	return core.VnicAttachment{}, nil
}

func (MockComputeClient) FindVolumeAttachment(ctx context.Context, compartmentID, volumeID string) (core.VolumeAttachment, error) {
	return nil, nil
}

func (MockComputeClient) AttachParavirtualizedVolume(ctx context.Context, instanceID, volumeID string, isPvEncryptionInTransitEnabled bool) (core.VolumeAttachment, error) {
	return nil, nil
}

func (MockComputeClient) AttachVolume(ctx context.Context, instanceID, volumeID string) (core.VolumeAttachment, error) {
	return nil, nil
}

func (MockComputeClient) WaitForVolumeAttached(ctx context.Context, attachmentID string) (core.VolumeAttachment, error) {
	return nil, nil
}

func (MockComputeClient) DetachVolume(ctx context.Context, id string) error {
	return nil
}

func (MockComputeClient) WaitForVolumeDetached(ctx context.Context, attachmentID string) error {
	return nil
}

func (c *MockComputeClient) FindActiveVolumeAttachment(ctx context.Context, compartmentID, volumeID string) (core.VolumeAttachment, error) {
	return nil, nil
}

func (c *MockComputeClient) WaitForUHPVolumeLoggedOut(ctx context.Context, attachmentID string) error {
	return nil
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

// MockVirtualNetworkClient mocks VirtualNetwork client implementation
type MockVirtualNetworkClient struct {
}

func (c *MockVirtualNetworkClient) ListPrivateIps(ctx context.Context, vnicId string) ([]core.PrivateIp, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) CreatePrivateIp(ctx context.Context, vnicID string) (*core.PrivateIp, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) GetIpv6(ctx context.Context, id string) (*core.Ipv6, error) {
	return &core.Ipv6{}, nil
}

func (c *MockVirtualNetworkClient) IsRegionalSubnet(ctx context.Context, id string) (bool, error) {
	return subnets[id].AvailabilityDomain == nil, nil
}

func (c *MockVirtualNetworkClient) GetPrivateIp(ctx context.Context, id string) (*core.PrivateIp, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) ListIpv6s(ctx context.Context, vnicId string) ([]core.Ipv6, error) {
	return []core.Ipv6{}, nil
}

func (c *MockVirtualNetworkClient) CreateIpv6(ctx context.Context, vnicID string) (*core.Ipv6, error) {
	return &core.Ipv6{}, nil
}

func (c *MockVirtualNetworkClient) GetSubnet(ctx context.Context, id string) (*core.Subnet, error) {
	if subnet, ok := subnets[id]; ok {
		return subnet, nil
	}
	return nil, errors.New("Subnet not found")
}

func (c *MockVirtualNetworkClient) GetVcn(ctx context.Context, id string) (*core.Vcn, error) {
	return vcns[id], nil
}

func (c *MockVirtualNetworkClient) GetVNIC(ctx context.Context, id string) (*core.Vnic, error) {
	return &core.Vnic{}, nil
}

func (c *MockVirtualNetworkClient) GetSubnetFromCacheByIP(ip client.IpAddresses) (*core.Subnet, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) GetSecurityList(ctx context.Context, id string) (core.GetSecurityListResponse, error) {
	return core.GetSecurityListResponse{}, nil
}

func (c *MockVirtualNetworkClient) UpdateSecurityList(ctx context.Context, id string, etag string, ingressRules []core.IngressSecurityRule, egressRules []core.EgressSecurityRule) (core.UpdateSecurityListResponse, error) {
	return core.UpdateSecurityListResponse{}, nil
}

func (c *MockVirtualNetworkClient) GetPublicIpByIpAddress(ctx context.Context, id string) (*core.PublicIp, error) {
	return nil, nil
}

// MockLoadBalancerClient mocks LoadBalancer client implementation.
type MockLoadBalancerClient struct{}

func (c *MockLoadBalancerClient) ListWorkRequests(ctx context.Context, compartmentId, lbId string) ([]*client.GenericWorkRequest, error) {
	return nil, nil
}

func (c *MockLoadBalancerClient) CreateLoadBalancer(ctx context.Context, details *client.GenericCreateLoadBalancerDetails, serviceUid *string) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) GetLoadBalancer(ctx context.Context, id string) (*client.GenericLoadBalancer, error) {
	return nil, nil
}

func (c *MockLoadBalancerClient) GetLoadBalancerByName(ctx context.Context, compartmentID string, name string) (*client.GenericLoadBalancer, error) {
	if lb, ok := loadBalancers[name]; ok {
		return lb, nil
	}
	return nil, nil
}

func (c *MockLoadBalancerClient) DeleteLoadBalancer(ctx context.Context, id string) (string, error) {
	if id == "test-uid-delete-err" {
		return "workReqId", errors.New("error")
	}
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

var updateLoadBalancerErrors = map[string]error{
	"work request fail": errors.New("internal server error"),
}

func (c *MockLoadBalancerClient) CreateRuleSet(ctx context.Context, lbID string, name string, details *loadbalancer.RuleSetDetails) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) UpdateRuleSet(ctx context.Context, lbID string, name string, details *loadbalancer.RuleSetDetails) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) DeleteRuleSet(ctx context.Context, lbID string, name string) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) UpdateLoadBalancer(ctx context.Context, lbID string, details *client.GenericUpdateLoadBalancerDetails) (string, error) {
	if err, ok := updateLoadBalancerErrors[lbID]; ok {
		return "", err
	}

	return "", nil
}

var awaitLoadbalancerWorkrequestMap = map[string]error{
	"failedToGetUpdateNetworkSecurityGroupsWorkRequest": errors.New("internal server error for get workrequest call"),
}

func (c *MockLoadBalancerClient) AwaitWorkRequest(ctx context.Context, id string) (*client.GenericWorkRequest, error) {
	if err, ok := awaitLoadbalancerWorkrequestMap[id]; ok {
		return nil, err
	}
	return nil, nil
}

func (c *MockLoadBalancerClient) UpdateLoadBalancerShape(context.Context, string, *client.GenericUpdateLoadBalancerShapeDetails) (string, error) {
	return "", nil
}

var updateNetworkSecurityGroupsLBsFailures = map[string]error{
	"":                      errors.New("provided LB ID is empty"),
	"failedToCreateRequest": errors.New("internal server error"),
}
var updateNetworkSecurityGroupsLBsWorkRequests = map[string]string{
	"failedToGetUpdateNetworkSecurityGroupsWorkRequest": "failedToGetUpdateNetworkSecurityGroupsWorkRequest",
}

func (c *MockLoadBalancerClient) UpdateNetworkSecurityGroups(ctx context.Context, lbId string, nsgIds []string) (string, error) {
	if err, ok := updateNetworkSecurityGroupsLBsFailures[lbId]; ok {
		return "", err
	}
	if wrID, ok := updateNetworkSecurityGroupsLBsWorkRequests[lbId]; ok {
		return wrID, nil
	}
	return "", nil
}

// MockNetworkLoadBalancerClient mocks NetworkLoadBalancer client implementation.
type MockNetworkLoadBalancerClient struct{}

func (c *MockNetworkLoadBalancerClient) ListWorkRequests(ctx context.Context, compartmentId, lbId string) ([]*client.GenericWorkRequest, error) {
	return nil, nil
}

func (c *MockNetworkLoadBalancerClient) CreateLoadBalancer(ctx context.Context, details *client.GenericCreateLoadBalancerDetails, serviceUid *string) (string, error) {
	return "", nil
}

func (c *MockNetworkLoadBalancerClient) GetLoadBalancer(ctx context.Context, id string) (*client.GenericLoadBalancer, error) {
	return nil, nil
}

func (c *MockNetworkLoadBalancerClient) GetLoadBalancerByName(ctx context.Context, compartmentID string, name string) (*client.GenericLoadBalancer, error) {
	if lb, ok := loadBalancers[name]; ok {
		return lb, nil
	}
	return nil, nil
}

func (c *MockNetworkLoadBalancerClient) DeleteLoadBalancer(ctx context.Context, id string) (string, error) {
	if id == "test-uid-delete-err" {
		return "workReqId", errors.New("error")
	}
	return "", nil
}

func (c *MockNetworkLoadBalancerClient) GetCertificateByName(ctx context.Context, lbID string, name string) (*client.GenericCertificate, error) {
	return nil, nil
}

func (c *MockNetworkLoadBalancerClient) CreateCertificate(ctx context.Context, lbID string, cert *client.GenericCertificate) (string, error) {
	return "", nil
}

func (c *MockNetworkLoadBalancerClient) CreateBackendSet(ctx context.Context, lbID string, name string, details *client.GenericBackendSetDetails) (string, error) {
	return "", nil
}

func (c *MockNetworkLoadBalancerClient) UpdateBackendSet(ctx context.Context, lbID string, name string, details *client.GenericBackendSetDetails) (string, error) {
	return "", nil
}

func (c *MockNetworkLoadBalancerClient) DeleteBackendSet(ctx context.Context, lbID, name string) (string, error) {
	return "", nil
}

func (c *MockNetworkLoadBalancerClient) UpdateListener(ctx context.Context, lbID string, name string, details *client.GenericListener) (string, error) {
	return "", nil
}

func (c *MockNetworkLoadBalancerClient) CreateListener(ctx context.Context, lbID string, name string, details *client.GenericListener) (string, error) {
	return "", nil
}

func (c *MockNetworkLoadBalancerClient) DeleteListener(ctx context.Context, lbID, name string) (string, error) {
	return "", nil
}

func (c *MockNetworkLoadBalancerClient) CreateRuleSet(ctx context.Context, lbID string, name string, details *loadbalancer.RuleSetDetails) (string, error) {
	return "", nil
}

func (c *MockNetworkLoadBalancerClient) UpdateRuleSet(ctx context.Context, lbID string, name string, details *loadbalancer.RuleSetDetails) (string, error) {
	return "", nil
}

func (c *MockNetworkLoadBalancerClient) DeleteRuleSet(ctx context.Context, lbID string, name string) (string, error) {
	return "", nil
}

func (c *MockNetworkLoadBalancerClient) AwaitWorkRequest(ctx context.Context, id string) (*client.GenericWorkRequest, error) {
	if err, ok := awaitLoadbalancerWorkrequestMap[id]; ok {
		return nil, err
	}
	return nil, nil
}

func (c *MockNetworkLoadBalancerClient) UpdateLoadBalancerShape(context.Context, string, *client.GenericUpdateLoadBalancerShapeDetails) (string, error) {
	return "", nil
}

func (c *MockNetworkLoadBalancerClient) UpdateNetworkSecurityGroups(ctx context.Context, lbId string, nsgIds []string) (string, error) {
	if err, ok := updateNetworkSecurityGroupsLBsFailures[lbId]; ok {
		return "", err
	}
	if wrID, ok := updateNetworkSecurityGroupsLBsWorkRequests[lbId]; ok {
		return wrID, nil
	}
	return "", nil
}

func (c *MockNetworkLoadBalancerClient) UpdateLoadBalancer(ctx context.Context, lbID string, details *client.GenericUpdateLoadBalancerDetails) (string, error) {
	return "", nil
}

// MockBlockStorageClient mocks BlockStorage client implementation
type MockBlockStorageClient struct{}

// AwaitVolumeCloneAvailableOrTimeout implements client.BlockStorageInterface.
func (*MockBlockStorageClient) AwaitVolumeCloneAvailableOrTimeout(ctx context.Context, id string) (*core.Volume, error) {
	return nil, nil
}

func (MockBlockStorageClient) AwaitVolumeAvailableORTimeout(ctx context.Context, id string) (*core.Volume, error) {
	return nil, nil
}

func (MockBlockStorageClient) CreateVolume(ctx context.Context, details core.CreateVolumeDetails) (*core.Volume, error) {
	return nil, nil
}

var updateVolumeErrors = map[string]error{
	"work-request-fails": errors.New("UpdateVolume request failed: internal server error"),
	"api-returns-too-many-requests": mockServiceError{
		StatusCode: http.StatusTooManyRequests,
		Code:       client.HTTP429TooManyRequestsCode,
		Message:    "Too many requests",
	},
}

func (c MockBlockStorageClient) UpdateVolume(ctx context.Context, volumeId string, details core.UpdateVolumeDetails) (*core.Volume, error) {
	if err, ok := updateVolumeErrors[volumeId]; ok {
		return nil, err
	}
	return nil, nil
}

func (MockBlockStorageClient) GetVolume(ctx context.Context, id string) (*core.Volume, error) {
	return nil, nil
}

func (MockBlockStorageClient) GetVolumesByName(ctx context.Context, volumeName, compartmentID string) ([]core.Volume, error) {
	return nil, nil
}

func (MockBlockStorageClient) DeleteVolume(ctx context.Context, id string) error {
	return nil
}

// MockFileStorageClient mocks FileStorage client implementation.
type MockFileStorageClient struct{}

func (c MockFileStorageClient) GetFileSystemSummaryByDisplayName(ctx context.Context, compartmentID, ad, displayName string) (bool, []filestorage.FileSystemSummary, error) {
	return false, nil, nil
}

func (c MockFileStorageClient) FindExport(ctx context.Context, fsID, path, exportSetID string) (*filestorage.ExportSummary, error) {
	return nil, nil
}

func (c MockFileStorageClient) GetMountTarget(ctx context.Context, id string) (*filestorage.MountTarget, error) {
	return nil, nil
}

func (c MockFileStorageClient) GetMountTargetSummaryByDisplayName(ctx context.Context, compartmentID, ad, mountTargetName string) (bool, []filestorage.MountTargetSummary, error) {
	return false, nil, nil
}

func (MockFileStorageClient) AwaitMountTargetActive(ctx context.Context, logger *zap.SugaredLogger, id string) (*filestorage.MountTarget, error) {
	return nil, nil
}

func (MockFileStorageClient) GetFileSystem(ctx context.Context, id string) (*filestorage.FileSystem, error) {
	return nil, nil
}

func (MockFileStorageClient) AwaitFileSystemActive(ctx context.Context, logger *zap.SugaredLogger, id string) (*filestorage.FileSystem, error) {
	return nil, nil
}

func (MockFileStorageClient) CreateFileSystem(ctx context.Context, details filestorage.CreateFileSystemDetails) (*filestorage.FileSystem, error) {
	return nil, nil
}

func (MockFileStorageClient) DeleteFileSystem(ctx context.Context, id string) error {
	return nil
}

func (MockFileStorageClient) CreateExport(ctx context.Context, details filestorage.CreateExportDetails) (*filestorage.Export, error) {
	return nil, nil
}

func (MockFileStorageClient) AwaitExportActive(ctx context.Context, logger *zap.SugaredLogger, id string) (*filestorage.Export, error) {
	return nil, nil
}

func (MockFileStorageClient) DeleteExport(ctx context.Context, id string) error {
	return nil
}

func (MockFileStorageClient) CreateMountTarget(ctx context.Context, details filestorage.CreateMountTargetDetails) (*filestorage.MountTarget, error) {
	return nil, nil
}

func (MockFileStorageClient) DeleteMountTarget(ctx context.Context, id string) error {
	return nil
}

// MockIdentityClient mocks Identity client implementaion
type MockIdentityClient struct{}

func (MockIdentityClient) GetAvailabilityDomainByName(ctx context.Context, compartmentID, name string) (*identity.AvailabilityDomain, error) {
	return nil, nil
}

func (MockIdentityClient) ListAvailabilityDomains(ctx context.Context, compartmentID string) ([]identity.AvailabilityDomain, error) {
	return nil, nil
}

type mockInstanceCache struct{}

func (m mockInstanceCache) Add(obj interface{}) error {
	return nil
}

func (m mockInstanceCache) Update(obj interface{}) error {
	return nil
}

func (m mockInstanceCache) Delete(obj interface{}) error {
	return nil
}

func (m mockInstanceCache) List() []interface{} {
	return nil
}

func (m mockInstanceCache) ListKeys() []string {
	return nil
}

func (m mockInstanceCache) Get(obj interface{}) (item interface{}, exists bool, err error) {
	return instances["default"], true, nil
}

func (m mockInstanceCache) GetByKey(key string) (item interface{}, exists bool, err error) {
	for _, instance := range instances {
		if *instance.Id == key {
			return instance, true, nil
		}
	}

	return nil, false, nil
}

func (m mockInstanceCache) Replace(i []interface{}, s string) error {
	return nil
}

func (m mockInstanceCache) Resync() error {
	return nil
}

func TestExtractNodeAddresses(t *testing.T) {
	testCases := []struct {
		name string
		in   string
		out  []v1.NodeAddress
		err  error
	}{
		{
			name: "basic-complete",
			in:   "ocid1.basic-complete",
			out: []v1.NodeAddress{
				{Type: v1.NodeInternalIP, Address: "10.0.0.1"},
				{Type: v1.NodeExternalIP, Address: "0.0.0.1"},
				// v1.NodeAddress{Type: v1.NodeHostName, Address: "basic-complete.subnetwithdnslabel.vcnwithdnslabel.oraclevcn.com"},
				// v1.NodeAddress{Type: v1.NodeInternalDNS, Address: "basic-complete.subnetwithdnslabel.vcnwithdnslabel.oraclevcn.com"},
			},
			err: nil,
		},
		{
			name: "no-external-ip",
			in:   "ocid1.no-external-ip",
			out: []v1.NodeAddress{
				{Type: v1.NodeInternalIP, Address: "10.0.0.1"},
				// v1.NodeAddress{Type: v1.NodeHostName, Address: "no-external-ip.subnetwithdnslabel.vcnwithdnslabel.oraclevcn.com"},
				// v1.NodeAddress{Type: v1.NodeInternalDNS, Address: "no-external-ip.subnetwithdnslabel.vcnwithdnslabel.oraclevcn.com"},
			},
			err: nil,
		},
		{
			name: "no-internal-ip",
			in:   "ocid1.no-internal-ip",
			out: []v1.NodeAddress{
				{Type: v1.NodeExternalIP, Address: "0.0.0.1"},
				// v1.NodeAddress{Type: v1.NodeHostName, Address: "no-internal-ip.subnetwithdnslabel.vcnwithdnslabel.oraclevcn.com"},
				// v1.NodeAddress{Type: v1.NodeInternalDNS, Address: "no-internal-ip.subnetwithdnslabel.vcnwithdnslabel.oraclevcn.com"},
			},
			err: nil,
		},
		{
			name: "invalid-external-ip",
			in:   "ocid1.invalid-external-ip",
			out:  nil,
			err:  errors.New(`instance has invalid public address: "0.0.0."`),
		},
		{
			name: "invalid-internal-ip",
			in:   "ocid1.invalid-internal-ip",
			out:  nil,
			err:  errors.New(`instance has invalid private address: "10.0.0."`),
		},
		{
			name: "no-hostname-label",
			in:   "ocid1.no-hostname-label",
			out: []v1.NodeAddress{
				{Type: v1.NodeInternalIP, Address: "10.0.0.1"},
				{Type: v1.NodeExternalIP, Address: "0.0.0.1"},
			},
			err: nil,
		},
		{
			name: "no-subnet-dns-label",
			in:   "ocid1.no-subnet-dns-label",
			out: []v1.NodeAddress{
				{Type: v1.NodeInternalIP, Address: "10.0.0.1"},
				{Type: v1.NodeExternalIP, Address: "0.0.0.1"},
			},
			err: nil,
		},
		{
			name: "no-vcn-dns-label",
			in:   "ocid1.no-vcn-dns-label",
			out: []v1.NodeAddress{
				{Type: v1.NodeInternalIP, Address: "10.0.0.1"},
				{Type: v1.NodeExternalIP, Address: "0.0.0.1"},
			},
			err: nil,
		},
		{
			name: "ipv6-instance",
			in:   "ocid1.ipv6-instance",
			out: []v1.NodeAddress{
				{Type: v1.NodeExternalIP, Address: "2001:db8:85a3::8a2e:370:7334"},
			},
			err: nil,
		},
		{
			name: "ipv6-instance-ULA",
			in:   "ocid1.ipv6-instance-ula",
			out: []v1.NodeAddress{
				{Type: v1.NodeInternalIP, Address: "fc00::"},
			},
			err: nil,
		},
		{
			name: "openshift-instance-ipv4",
			in:   "ocid1.openshift-instance-ipv4",
			out: []v1.NodeAddress{
				{Type: v1.NodeInternalIP, Address: "10.0.0.1"},
				{Type: v1.NodeExternalIP, Address: "0.0.0.1"},
			},
			err: nil,
		},
		{
			name: "openshift-instance-invalid",
			in:   "ocid1.openshift-instance-invalid",
			out:  nil,
			err:  errors.New(`instance has invalid private address: "10.0.0."`),
		},
		{
			name: "openshift-instance-ipv6",
			in:   "ocid1.openshift-instance-ipv6",
			out: []v1.NodeAddress{
				{Type: v1.NodeExternalIP, Address: "2001:db8:85a3::8a2e:370:7334"},
			},
			err: nil,
		},
	}

	cp := &CloudProvider{
		client:        MockOCIClient{},
		config:        &providercfg.Config{CompartmentID: "testCompartment"},
		NodeLister:    &mockNodeLister{},
		instanceCache: &mockInstanceCache{},
		logger:        zap.S(),
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cp.extractNodeAddresses(context.Background(), tt.in)
			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("extractNodeAddresses(context, %+v) got error %v, expected %v", tt.in, err, tt.err)
			}
			if !reflect.DeepEqual(result, tt.out) {
				t.Errorf("extractNodeAddresses(context, %+v) => %+v, want %+v", tt.in, result, tt.out)
			}
		})
	}
}

func TestInstanceID(t *testing.T) {
	testCases := []struct {
		name string
		in   types.NodeName
		out  string
		err  error
	}{
		{
			name: "get instance id from instance in the cache",
			in:   "instance1",
			out:  "ocid1.instance1",
			err:  nil,
		},
		{
			name: "get instance id from instance not in the cache",
			in:   "default",
			out:  "default",
			err:  nil,
		},
	}

	cp := &CloudProvider{
		NodeLister:    &mockNodeLister{},
		client:        MockOCIClient{},
		config:        &providercfg.Config{CompartmentID: "testCompartment"},
		logger:        zap.S(),
		instanceCache: &mockInstanceCache{},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cp.InstanceID(context.Background(), tt.in)
			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("InstanceID(context, %+v) got error %v, expected %v", tt.in, err, tt.err)
			}
			if !reflect.DeepEqual(result, tt.out) {
				t.Errorf("InstanceID(context, %+v) => %+v, want %+v", tt.in, result, tt.out)
			}
		})
	}
}

func TestInstanceType(t *testing.T) {
	testCases := []struct {
		name string
		in   types.NodeName
		out  string
		err  error
	}{
		{
			name: "check node shape of instance in cache",
			in:   "instance1",
			out:  "VM.Standard1.2",
			err:  nil,
		},
		{
			name: "check node shape of instance not in cache",
			in:   "default",
			out:  "VM.Standard1.2",
			err:  nil,
		},
	}

	cp := &CloudProvider{
		NodeLister:    &mockNodeLister{},
		client:        MockOCIClient{},
		config:        &providercfg.Config{CompartmentID: "testCompartment"},
		logger:        zap.S(),
		instanceCache: &mockInstanceCache{},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cp.InstanceType(context.Background(), tt.in)
			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("InstanceType(context, %+v) got error %v, expected %v", tt.in, err, tt.err)
			}
			if !reflect.DeepEqual(result, tt.out) {
				t.Errorf("InstanceType(context, %+v) => %+v, want %+v", tt.in, result, tt.out)
			}
		})
	}
}

func TestGetNodeIpFamily(t *testing.T) {
	testCases := []struct {
		name string
		in   string
		out  []string
		err  error
	}{
		{
			name: "IPv4",
			in:   "ocid1.instance-id-ipv4",
			out:  []string{IPv4},
			err:  nil,
		},
		{
			name: "IPv6",
			in:   "ocid1.instance-id-ipv6",
			out:  []string{IPv6},
			err:  nil,
		},
		{
			name: "IPv4 & IPv6",
			in:   "ocid1.instance-id-ipv4-ipv6",
			out:  []string{IPv4, IPv6},
			err:  nil,
		},
	}

	cp := &CloudProvider{
		NodeLister:    &mockNodeLister{},
		client:        MockOCIClient{},
		config:        &providercfg.Config{CompartmentID: "testCompartment"},
		logger:        zap.S(),
		instanceCache: &mockInstanceCache{},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cp.getNodeIpFamily(tt.in)
			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("getNodeIpFamily(context, %+v) got error %v, expected %v", tt.in, err, tt.err)
			}
			if !reflect.DeepEqual(result, tt.out) {
				t.Errorf("getNodeIpFamily(context, %+v) => %+v, want %+v", tt.name, result, tt.out)
			}
		})
	}
}

func TestInstanceTypeByProviderID(t *testing.T) {
	testCases := []struct {
		name string
		in   string
		out  string
		err  error
	}{
		{
			name: "provider id without provider prefix",
			in:   "ocid1.instance1",
			out:  "VM.Standard1.2",
			err:  nil,
		},
		{
			name: "provider id with provider prefix",
			in:   providerPrefix + "instance1",
			out:  "VM.Standard1.2",
			err:  nil,
		},
		{
			name: "provider id with provider prefix and instance not in cache",
			in:   providerPrefix + "noncacheinstance",
			out:  "VM.Standard1.2",
			err:  nil,
		},
	}

	cp := &CloudProvider{
		NodeLister:    &mockNodeLister{},
		client:        MockOCIClient{},
		config:        &providercfg.Config{CompartmentID: "testCompartment"},
		logger:        zap.S(),
		instanceCache: &mockInstanceCache{},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cp.InstanceTypeByProviderID(context.Background(), tt.in)
			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("InstanceTypeByProviderID(context, %+v) got error %v, expected %v", tt.in, err, tt.err)
			}
			if !reflect.DeepEqual(result, tt.out) {
				t.Errorf("InstanceTypeByProviderID(context, %+v) => %+v, want %+v", tt.in, result, tt.out)
			}
		})
	}
}

func TestNodeAddressesByProviderID(t *testing.T) {
	testCases := []struct {
		name string
		in   string
		out  []v1.NodeAddress
		err  error
	}{
		{
			name: "provider id without provider prefix",
			in:   "ocid1.basic-complete",
			out: []v1.NodeAddress{
				{Type: v1.NodeInternalIP, Address: "10.0.0.1"},
				{Type: v1.NodeExternalIP, Address: "0.0.0.1"},
			},
			err: nil,
		},
		{
			name: "provider id with provider prefix",
			in:   providerPrefix + "ocid1.basic-complete",
			out: []v1.NodeAddress{
				{Type: v1.NodeInternalIP, Address: "10.0.0.1"},
				{Type: v1.NodeExternalIP, Address: "0.0.0.1"},
			},
			err: nil,
		},
	}

	cp := &CloudProvider{
		NodeLister:    &mockNodeLister{},
		client:        MockOCIClient{},
		config:        &providercfg.Config{CompartmentID: "testCompartment"},
		logger:        zap.S(),
		instanceCache: &mockInstanceCache{},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cp.NodeAddressesByProviderID(context.Background(), tt.in)
			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("NodeAddressesByProviderID(context, %+v) got error %v, expected %v", tt.in, err, tt.err)
			}
			if !reflect.DeepEqual(result, tt.out) {
				t.Errorf("NodeAddressesByProviderID(context, %+v) => %+v, want %+v", tt.in, result, tt.out)
			}
		})
	}
}

func TestInstanceExistsByProviderID(t *testing.T) {
	testCases := []struct {
		name string
		in   string
		out  bool
		err  error
	}{
		{
			name: "provider id without provider prefix",
			in:   "ocid1.instance1",
			out:  true,
			err:  nil,
		},
		{
			name: "provider id with provider prefix",
			in:   providerPrefix + "ocid1.instance1",
			out:  true,
			err:  nil,
		},
		{
			name: "provider id with provider prefix and instance not in cache",
			in:   providerPrefix + "ocid1.noncacheinstance",
			out:  true,
			err:  nil,
		},
	}

	cp := &CloudProvider{
		NodeLister:    &mockNodeLister{},
		client:        MockOCIClient{},
		config:        &providercfg.Config{CompartmentID: "testCompartment"},
		logger:        zap.S(),
		instanceCache: &mockInstanceCache{},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cp.InstanceExistsByProviderID(context.Background(), tt.in)
			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("InstanceExistsByProviderID(context, %+v) got error %v, expected %v", tt.in, err, tt.err)
			}
			if !reflect.DeepEqual(result, tt.out) {
				t.Errorf("InstanceExistsByProviderID(context, %+v) => %+v, want %+v", tt.in, result, tt.out)
			}
		})
	}
}

func TestInstanceShutdownByProviderID(t *testing.T) {
	testCases := []struct {
		name string
		in   string
		out  bool
		err  error
	}{
		{
			name: "provider id without provider prefix",
			in:   "ocid1.instance1",
			out:  false,
			err:  nil,
		},
		{
			name: "provider id with provider prefix",
			in:   providerPrefix + "ocid1.instance1",
			out:  false,
			err:  nil,
		},
		{
			name: "provider id with provider prefix and instance not in cache",
			in:   providerPrefix + "ocid1.noncacheinstance",
			out:  false,
			err:  nil,
		},
	}

	cp := &CloudProvider{
		NodeLister:    &mockNodeLister{},
		client:        MockOCIClient{},
		config:        &providercfg.Config{CompartmentID: "testCompartment"},
		logger:        zap.S(),
		instanceCache: &mockInstanceCache{},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cp.InstanceShutdownByProviderID(context.Background(), tt.in)
			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("InstanceShutdownByProviderID(context, %+v) got error %v, expected %v", tt.in, err, tt.err)
			}
			if !reflect.DeepEqual(result, tt.out) {
				t.Errorf("InstanceShutdownByProviderID(context, %+v) => %+v, want %+v", tt.in, result, tt.out)
			}
		})
	}
}

func TestGetCompartmentIDByInstanceID(t *testing.T) {
	testCases := []struct {
		name string
		in   string
		out  string
		err  error
	}{
		{
			name: "instance found in cache",
			in:   "ocid1.instance1",
			out:  "compartment1",
			err:  nil,
		},
		{
			name: "instance found in node lister",
			in:   "ocid1.default",
			out:  "default",
			err:  nil,
		},
		{
			name: "instance neither found in cache nor node lister",
			in:   "ocid1.instancex",
			out:  "",
			err:  errors.New("compartmentID annotation missing in the node. Would retry"),
		},
	}

	cp := &CloudProvider{
		client:        MockOCIClient{},
		config:        &providercfg.Config{CompartmentID: "testCompartment"},
		NodeLister:    &mockNodeLister{},
		instanceCache: &mockInstanceCache{},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cp.getCompartmentIDByInstanceID(tt.in)
			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("getCompartmentIDByInstanceID(%s) got error %s, expected %s", tt.in, err, tt.err)
			}
			if !reflect.DeepEqual(result, tt.out) {
				t.Errorf("getCompartmentIDByInstanceID(%s) => %s, want %s", tt.in, result, tt.out)
			}
		})
	}
}

type mockNodeLister struct {
	nodes []*v1.Node
}

func (s *mockNodeLister) List(selector labels.Selector) (ret []*v1.Node, err error) {
	var nodes, allNodes []*v1.Node
	if len(s.nodes) > 0 {
		allNodes = s.nodes
	} else {
		for _, n := range nodeList {
			allNodes = append(allNodes, n)
		}
	}

	for _, n := range allNodes {
		if selector != nil {
			if selector.Matches(labels.Set(n.ObjectMeta.GetLabels())) {
				nodes = append(nodes, n)
			}
		} else {
			nodes = append(nodes, n)
		}
	}
	return nodes, nil
}

func (s *mockNodeLister) Get(name string) (*v1.Node, error) {
	if len(s.nodes) > 0 {
		for _, n := range s.nodes {
			if n.Name == name {
				return n, nil
			}
		}
	} else if node, ok := nodeList[name]; ok {
		return node, nil
	}
	return nil, errors.New("get node error")
}

func (s *mockNodeLister) ListWithPredicate() ([]*v1.Node, error) {
	return nil, nil
}

type mockEndpointSliceLister struct{}

func (s *mockEndpointSliceLister) List(selector labels.Selector) (ret []*v1discovery.EndpointSlice, err error) {
	return []*v1discovery.EndpointSlice{}, nil
}

func (s *mockEndpointSliceLister) EndpointSlices(namespace string) v1discoverylisters.EndpointSliceNamespaceLister {
	return &mockEndpointSliceNamespaceLister{}
}

type mockEndpointSliceNamespaceLister struct{}

func (s *mockEndpointSliceNamespaceLister) List(selector labels.Selector) (ret []*v1discovery.EndpointSlice, err error) {
	var endpointSlices []*v1discovery.EndpointSlice
	for _, es := range endpointSliceList {
		if selector != nil {
			if selector.Matches(labels.Set(es.ObjectMeta.GetLabels())) {
				endpointSlices = append(endpointSlices, es)
			}
		} else {
			endpointSlices = append(endpointSlices, es)
		}
	}
	return endpointSlices, nil
}

func (s *mockEndpointSliceNamespaceLister) Get(name string) (ret *v1discovery.EndpointSlice, err error) {
	if es, ok := endpointSliceList[name]; ok {
		return es, nil
	}
	return nil, errors.New("get endpointSlice error")
}

type mockServiceError struct {
	StatusCode   int
	Code         string
	Message      string
	OpcRequestID string
}

func (m mockServiceError) GetHTTPStatusCode() int {
	return m.StatusCode
}

func (m mockServiceError) GetMessage() string {
	return m.Message
}

func (m mockServiceError) GetCode() string {
	return m.Code
}

func (m mockServiceError) GetOpcRequestID() string {
	return m.OpcRequestID
}
func (m mockServiceError) Error() string {
	return m.Message
}

func TestGetOpenShiftTagNamespaceByInstance(t *testing.T) {
	testCases := []struct {
		name       string
		instanceID string
		expected   string
	}{
		{
			name:       "Instance with OpenShift namespace",
			instanceID: "openshift-instance-ipv4",
			expected:   "openshift-namespace",
		},
		{
			name:       "Instance without OpenShift namespace",
			instanceID: "basic-complete",
			expected:   "",
		},
		{
			name:       "Non-existent instance",
			instanceID: "non-existent-instance-id",
			expected:   "",
		},
	}

	cp := &CloudProvider{
		client:        MockOCIClient{},
		config:        &providercfg.Config{CompartmentID: "testCompartment"},
		NodeLister:    &mockNodeLister{},
		instanceCache: &mockInstanceCache{},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result := cp.getOpenShiftTagNamespaceByInstance(context.Background(), tt.instanceID)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestCheckOpenShiftISCSIBootVolumeTagByVnic(t *testing.T) {
	testCases := []struct {
		name      string
		vnic      *core.Vnic
		namespace string
		expected  bool
	}{
		{
			name: "VNIC with ISCSI boot volume tag",
			vnic: &core.Vnic{
				DefinedTags: map[string]map[string]interface{}{
					"openshift-namespace": {
						"boot-volume-type": "ISCSI",
					},
				},
			},
			namespace: "openshift-namespace",
			expected:  true,
		},
		{
			name: "VNIC without ISCSI boot volume tag",
			vnic: &core.Vnic{
				DefinedTags: map[string]map[string]interface{}{
					"openshift-namespace": {
						"boot-volume-type": "NVMe",
					},
				},
			},
			namespace: "openshift-namespace",
			expected:  false,
		},
		{
			name: "VNIC with no defined tags",
			vnic: &core.Vnic{
				DefinedTags: nil,
			},
			namespace: "openshift-namespace",
			expected:  false,
		},
		{
			name: "Namespace not found in VNIC tags",
			vnic: &core.Vnic{
				DefinedTags: map[string]map[string]interface{}{
					"another-namespace": {
						"bootVolumeType": "ISCSI",
					},
				},
			},
			namespace: "openshift-namespace",
			expected:  false,
		},
	}

	cp := &CloudProvider{
		client:        MockOCIClient{},
		config:        &providercfg.Config{CompartmentID: "testCompartment"},
		NodeLister:    &mockNodeLister{},
		instanceCache: &mockInstanceCache{},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result := cp.checkOpenShiftISCSIBootVolumeTagByVnic(context.Background(), tt.vnic, tt.namespace)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
