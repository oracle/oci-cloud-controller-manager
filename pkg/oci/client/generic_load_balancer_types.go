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

package client

type GenericBackendSetDetails struct {
	Name                            *string
	HealthChecker                   *GenericHealthChecker
	Policy                          *string
	Backends                        []GenericBackend
	SessionPersistenceConfiguration *GenericSessionPersistenceConfiguration
	// Only needed for LB
	SslConfiguration *GenericSslConfigurationDetails
	// Only needed for NLB
	IsPreserveSource *bool
}

type GenericSessionPersistenceConfiguration struct {
	CookieName      *string
	DisableFallback *bool
}

type GenericHealthChecker struct {
	Protocol          string
	IsForcePlainText  *bool
	Port              *int
	UrlPath           *string
	Retries           *int
	TimeoutInMillis   *int
	IntervalInMillis  *int
	ResponseBodyRegex *string
	// Only needed for NLB
	ReturnCode *int
}

type GenericBackend struct {
	Port      *int
	Name      *string
	IpAddress *string
	TargetId  *string
	Weight    *int
}

type GenericSslConfigurationDetails struct {
	VerifyDepth                    *int
	VerifyPeerCertificate          *bool
	TrustedCertificateAuthorityIds []string
	CertificateIds                 []string
	CertificateName                *string
	ServerOrderPreference          string
	CipherSuiteName                *string
	Protocols                      []string
}

type GenericListener struct {
	Name                    *string
	DefaultBackendSetName   *string
	Port                    *int
	Protocol                *string
	HostnameNames           []string
	PathRouteSetName        *string
	SslConfiguration        *GenericSslConfigurationDetails
	ConnectionConfiguration *GenericConnectionConfiguration
	RoutingPolicyName       *string
	RuleSetNames            []string
}

type GenericConnectionConfiguration struct {
	IdleTimeout                    *int64
	BackendTcpProxyProtocolVersion *int
	BackendTcpProxyProtocolOptions []string
}

type GenericCreateLoadBalancerDetails struct {
	CompartmentId               *string
	DisplayName                 *string
	ShapeName                   *string
	SubnetIds                   []string
	ShapeDetails                *GenericShapeDetails
	IsPrivate                   *bool
	IsPreserveSourceDestination *bool
	ReservedIps                 []GenericReservedIp
	Listeners                   map[string]GenericListener
	BackendSets                 map[string]GenericBackendSetDetails
	NetworkSecurityGroupIds     []string
	FreeformTags                map[string]string
	DefinedTags                 map[string]map[string]interface{}

	// Only needed for LB
	Certificates map[string]GenericCertificate
}

type GenericShapeDetails struct {
	MinimumBandwidthInMbps *int
	MaximumBandwidthInMbps *int
}

type GenericCertificate struct {
	CertificateName   *string
	Passphrase        *string
	PrivateKey        *string
	PublicCertificate *string
	CaCertificate     *string
}

type GenericReservedIp struct {
	Id *string
}

type GenericIpAddress struct {
	IpAddress  *string
	IsPublic   *bool
	ReservedIp *GenericReservedIp
}

type GenericUpdateLoadBalancerShapeDetails struct {
	ShapeName    *string
	ShapeDetails *GenericShapeDetails
}

type GenericLoadBalancer struct {
	Id                      *string
	CompartmentId           *string
	DisplayName             *string
	LifecycleState          *string
	ShapeName               *string
	IpAddresses             []GenericIpAddress
	ShapeDetails            *GenericShapeDetails
	IsPrivate               *bool
	SubnetIds               []string
	NetworkSecurityGroupIds []string
	Listeners               map[string]GenericListener
	Certificates            map[string]GenericCertificate
	BackendSets             map[string]GenericBackendSetDetails

	FreeformTags map[string]string
	DefinedTags  map[string]map[string]interface{}
	SystemTags   map[string]map[string]interface{}
}

type GenericWorkRequest struct {
	Id             *string
	LoadBalancerId *string
	Type           *string
	LifecycleState *string
	Message        *string
	CompartmentId  *string
	OperationType  string
	Status         string
}

type GenericUpdateNetworkSecurityGroupsDetails struct {
	NetworkSecurityGroupIds []string
}

type GenericUpdateLoadBalancerDetails struct {
	FreeformTags map[string]string
	DefinedTags  map[string]map[string]interface{}
}
