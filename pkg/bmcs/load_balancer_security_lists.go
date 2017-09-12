// Copyright 2017 The Oracle Kubernetes Cloud Controller Manager Authors
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

package bmcs

import (
	"fmt"
	"reflect"

	"github.com/oracle/kubernetes-cloud-controller-manager/pkg/bmcs/client"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/golang/glog"

	baremetal "github.com/oracle/bmcs-go-sdk"
)

const (
	// ProtocolTCP is the IANA decimal protocol number for the Transmission
	// Control Protocol (TCP).
	ProtocolTCP = 6
	// ProtocolUDP is the IANA decimal protocol number for the User
	// Datagram Protocol (UDP).
	ProtocolUDP = 17
)

type securityListManager interface {
	EnsureRulesAdded(port uint64) error
	EnsureRulesRemoved(port uint64) error
	Save() error
}

// securityListManagerNOOP implements the securityListManager interface but does
// no logic, so that it can be used to not handle security lists if the user doesn't wish
// to use that feature.
type securityListManagerNOOP struct {
}

func (m *securityListManagerNOOP) EnsureRulesAdded(port uint64) error {
	return nil
}
func (m *securityListManagerNOOP) EnsureRulesRemoved(port uint64) error {
	return nil
}
func (m *securityListManagerNOOP) Save() error {
	return nil
}

func newSecurityListManagerNOOP() securityListManager {
	return &securityListManagerNOOP{}
}

// securityListManagerImpl manages SecurityList rules (ingress and egress)
// associated with a load balancer.
type securityListManagerImpl struct {
	BackendSubnets             []*baremetal.Subnet
	LoadBalancerSubnets        []*baremetal.Subnet
	SecurityLists              map[string]*baremetal.SecurityList
	SubnetDefaultSecurityLists map[string]string
	ModifiedSecurityLists      sets.String
	client                     client.Interface
}

func newSecurityListManagerFromLBSpec(config *client.Config, c client.Interface, spec *LBSpec) (securityListManager, error) {
	if config.Global.DisableSecurityListManagement {
		return newSecurityListManagerNOOP(), nil
	}

	backendSubnets, err := c.GetSubnetsForInternalIPs(spec.NodeIPs)
	if err != nil {
		return nil, err
	}

	lbSubnets, err := c.GetSubnets(spec.Subnets)
	if err != nil {
		return nil, err
	}

	mngr := newSecurityListManager(c, backendSubnets, lbSubnets)
	err = mngr.(*securityListManagerImpl).manageDefaultSecuriyLists()
	if err != nil {
		return nil, err
	}
	return mngr, nil
}

func newSecurityListManager(c client.Interface, backendSubnets, lbSubnets []*baremetal.Subnet) securityListManager {
	return &securityListManagerImpl{
		BackendSubnets:             backendSubnets,
		LoadBalancerSubnets:        lbSubnets,
		SecurityLists:              make(map[string]*baremetal.SecurityList),
		SubnetDefaultSecurityLists: make(map[string]string),
		ModifiedSecurityLists:      sets.NewString(),
		client:                     c,
	}
}

// TODO(apryde): UDP support.
func makeEgressSecurityRule(cidrBlock string, port uint64) baremetal.EgressSecurityRule {
	return baremetal.EgressSecurityRule{
		Destination: cidrBlock,
		Protocol:    fmt.Sprintf("%d", ProtocolTCP),
		TCPOptions: &baremetal.TCPOptions{
			DestinationPortRange: baremetal.PortRange{
				Min: port,
				Max: port,
			},
		},
		IsStateless: false,
	}
}

// TODO(apryde): UDP support.
func makeIngressSecurityRule(cidrBlock string, port uint64) baremetal.IngressSecurityRule {
	return baremetal.IngressSecurityRule{
		Source:   cidrBlock,
		Protocol: fmt.Sprintf("%d", ProtocolTCP),
		TCPOptions: &baremetal.TCPOptions{
			DestinationPortRange: baremetal.PortRange{
				Min: port,
				Max: port,
			},
		},
		IsStateless: false,
	}
}

func (m *securityListManagerImpl) manageDefaultSecuriyLists() error {
	for _, subnet := range m.BackendSubnets {
		err := m.manageDefaultSecurityListForSubnet(subnet)
		if err != nil {
			return err
		}
	}

	for _, subnet := range m.LoadBalancerSubnets {
		err := m.manageDefaultSecurityListForSubnet(subnet)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *securityListManagerImpl) manageDefaultSecurityListForSubnet(subnet *baremetal.Subnet) error {
	if m.listIsManagedForSubnet(subnet.ID) {
		glog.V(4).Infof("Default SecurityList for Subnet '%s' already managed", subnet.ID)
		return nil
	}

	list, err := m.client.GetDefaultSecurityList(subnet)
	if err != nil {
		return err
	}
	m.addSecurityListForSubnet(list, subnet.ID)

	glog.V(4).Infof("Default SecurityList '%s' managed for Subnet '%s'", list.ID, subnet.ID)
	return nil
}

// addSecurityListForSubnet adds a defaullt SecurityList for the given Subnet
// ID.
func (m *securityListManagerImpl) addSecurityListForSubnet(list *baremetal.SecurityList, subnetID string) {
	m.SecurityLists[list.ID] = list
	m.SubnetDefaultSecurityLists[subnetID] = list.ID
}

func (m *securityListManagerImpl) listIsManaged(id string) bool {
	_, ok := m.SecurityLists[id]
	return ok
}

func (m *securityListManagerImpl) listIsManagedForSubnet(id string) bool {
	_, ok := m.SubnetDefaultSecurityLists[id]
	return ok
}

func (m *securityListManagerImpl) getSecurityListForSubnet(subnetID string) (*baremetal.SecurityList, error) {
	listID, ok := m.SubnetDefaultSecurityLists[subnetID]
	if !ok {
		return nil, fmt.Errorf("no default SecurityList manged for Subnet '%s'", subnetID)
	}

	return m.SecurityLists[listID], nil
}

func (m *securityListManagerImpl) EnsureRulesAdded(port uint64) error {
	for _, backendSubnet := range m.BackendSubnets {
		for _, lbSubnet := range m.LoadBalancerSubnets {
			// 1. Egress lbSubnet -> clusterSubnet
			egress := makeEgressSecurityRule(backendSubnet.CIDRBlock, port)
			err := m.addEgressRule(lbSubnet.ID, egress)
			if err != nil {
				return err
			}

			// 2. Ingress clusterSubnet <- lbSubnet
			ingress := makeIngressSecurityRule(lbSubnet.CIDRBlock, port)
			err = m.addIngressRule(backendSubnet.ID, ingress)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *securityListManagerImpl) EnsureRulesRemoved(port uint64) error {
	for _, backendSubnet := range m.BackendSubnets {
		for _, lbSubnet := range m.LoadBalancerSubnets {
			// 1. Egress lbSubnet -> clusterSubnet
			egress := makeEgressSecurityRule(backendSubnet.CIDRBlock, port)
			err := m.removeEgressRule(lbSubnet.ID, egress)
			if err != nil {
				return err
			}

			// 2. Ingress clusterSubnet <- lbSubnet
			ingress := makeIngressSecurityRule(lbSubnet.CIDRBlock, port)
			err = m.removeIngressRule(backendSubnet.ID, ingress)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *securityListManagerImpl) addIngressRule(subnetID string, rule baremetal.IngressSecurityRule) error {
	glog.V(4).Infof("Adding IngressSecurityRule %+v to default SecurityList for Subnet '%s'", rule, subnetID)
	list, err := m.getSecurityListForSubnet(subnetID)
	if err != nil {
		return err
	}

	for _, existingRule := range list.IngressSecurityRules {
		if reflect.DeepEqual(existingRule, rule) {
			glog.V(4).Infof("IngressSecurityRule %v exists. Nothing to do.", rule)
			return nil
		}
	}

	list.IngressSecurityRules = append(list.IngressSecurityRules, rule)
	m.ModifiedSecurityLists.Insert(list.ID)

	return nil
}

func (m *securityListManagerImpl) addEgressRule(subnetID string, rule baremetal.EgressSecurityRule) error {
	glog.V(4).Infof("Adding EgressSecurityRule %+v to default SecurityList for Subnet '%s'", rule, subnetID)
	list, err := m.getSecurityListForSubnet(subnetID)
	if err != nil {
		return err
	}

	for _, existingRule := range list.EgressSecurityRules {
		if reflect.DeepEqual(existingRule, rule) {
			glog.V(4).Infof("EgressSecurityRule %v exists. Nothing to do.", rule)
			return nil
		}
	}

	list.EgressSecurityRules = append(list.EgressSecurityRules, rule)
	m.ModifiedSecurityLists.Insert(list.ID)

	return nil
}

func (m *securityListManagerImpl) removeIngressRule(subnetID string, rule baremetal.IngressSecurityRule) error {
	glog.V(4).Infof("Removing IngressSecurityRule %+v from default SecurityList for Subnet '%s'", rule, subnetID)
	list, err := m.getSecurityListForSubnet(subnetID)
	if err != nil {
		return err
	}

	for i, existingRule := range list.IngressSecurityRules {
		if reflect.DeepEqual(existingRule, rule) {
			list.IngressSecurityRules = append(
				list.IngressSecurityRules[:i],
				list.IngressSecurityRules[i+1:]...)
			m.ModifiedSecurityLists.Insert(list.ID)
			return nil
		}
	}
	glog.V(4).Infof("Did not find IngressSecurityRule %v. Nothing to do.", rule)
	return nil
}

func (m *securityListManagerImpl) removeEgressRule(subnetID string, rule baremetal.EgressSecurityRule) error {
	glog.V(4).Infof("Removing EgressSecurityRule %+v from default SecurityList for Subnet '%s'", rule, subnetID)
	list, err := m.getSecurityListForSubnet(subnetID)
	if err != nil {
		return err
	}

	for i, existingRule := range list.EgressSecurityRules {
		if reflect.DeepEqual(existingRule, rule) {
			list.EgressSecurityRules = append(
				list.EgressSecurityRules[:i],
				list.EgressSecurityRules[i+1:]...)
			m.ModifiedSecurityLists.Insert(list.ID)
			return nil
		}
	}
	glog.V(4).Infof("Did not find EgressSecurityRule %v. Nothing to do.", rule)
	return nil
}

// Save stores the updated SecurityLists in the cloud.
func (m *securityListManagerImpl) Save() error {
	modified := m.ModifiedSecurityLists.List()
	glog.V(4).Infof("Saving %d modified SecurityLists", len(modified))

	for _, listID := range modified {
		list := m.SecurityLists[listID]
		opts := &baremetal.UpdateSecurityListOptions{
			EgressRules:  list.EgressSecurityRules,
			IngressRules: list.IngressSecurityRules,
		}
		_, err := m.client.UpdateSecurityList(list.ID, opts)
		if err != nil {
			return err
		}
	}
	return nil
}
