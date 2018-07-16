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
	"sync"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/cloudprovider"

	"strings"

	"github.com/golang/glog"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/util"
	"github.com/oracle/oci-go-sdk/core"
	"github.com/pkg/errors"
)

var _ cloudprovider.Routes = &CloudProvider{}

/*
OCI doesn't have distinct add/remove/update route commands, only the ability to
update the whole RouteTable.

K8s dispatches up to 200 concurrent creates.

As a result ADD/Create races with itself, especially in light of DELETE.

Ultimately the reconciliation restores order in the galaxy, but only after some
disruptions in the force.

One _could_ add some locking on the Get/Update of RouteTable, but it's not
clear that should be required.
*/

// CreateRoute for OCI, we have to udpate the whole RouteTable not just add the singular route
func (cp *CloudProvider) CreateRoute(ctx context.Context, clusterName string, nameHint string, route *cloudprovider.Route) error {
	glog.V(6).Info("Adding route to route table: ", route)

	node, err := cp.NodeLister.Get(string(route.TargetNode))
	if err != nil {
		return errors.WithStack(err)
	}

	ocid := util.MapProviderIDToInstanceID(node.Spec.ProviderID)
	vnic, err := cp.client.Compute().GetPrimaryVNICForInstance(ctx, "", ocid)
	if err != nil {
		return errors.WithStack(err)
	}

	ipOcid, err := cp.client.Networking().GetOCIDFromIP(ctx, *vnic.PrivateIp, *vnic.SubnetId)
	if err != nil {
		return errors.WithStack(err)
	}

	glog.V(6).Infof("ipOcid: ", ipOcid)

	for routeTableID := range routeTableIds {
		routeTable, err := cp.client.Networking().GetRouteTable(ctx, routeTableID)
		if err != nil {
			return errors.WithStack(err)
		}
		for _, oRoute := range routeTable.RouteRules {
			if *oRoute.CidrBlock == route.DestinationCIDR {
				if *oRoute.NetworkEntityId == ipOcid {
					// skipping existing route table entry
					glog.V(6).Infof("Skipping existing route entry: %s -> %s", route.DestinationCIDR, ipOcid)
					continue
				} else {
					oRoute.NetworkEntityId = &ipOcid
					glog.V(6).Infof("Update existing route entry: %s -> %s", route.DestinationCIDR, ipOcid)
					err := cp.client.Networking().UpdateRouteTable(ctx, routeTableID, routeTable.RouteRules)
					if err != nil {
						return errors.WithStack(err)
					}
					continue
				}
			}
		}

		// route rule doesn't exist in RouteTable
		glog.V(6).Info("Add new route entry: ", route.DestinationCIDR, " -> ", ipOcid)
		routeTable.RouteRules = append(routeTable.RouteRules, core.RouteRule{
			CidrBlock:       &route.DestinationCIDR,
			NetworkEntityId: &ipOcid,
		})

		err = cp.client.Networking().UpdateRouteTable(ctx, routeTableID, routeTable.RouteRules)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

// DeleteRoute delete a route entry, but OCI only has Update of the entire RouteTable
func (cp *CloudProvider) DeleteRoute(ctx context.Context, clusterName string, route *cloudprovider.Route) error {
	glog.V(6).Infof("Delete Route: %v", route)

	for routeID := range routeTableIds {
		oRoute, err := cp.client.Networking().GetRouteTable(ctx, routeID)
		if err != nil {
			return errors.WithStack(err)
		}

		dirty := false

		for i, r := range oRoute.RouteRules {
			if *r.CidrBlock != route.DestinationCIDR {
				continue
			}

			dirty = true

			glog.V(6).Infof("Deleting Route: %s %s", route.DestinationCIDR, route.TargetNode)
			oRoute.RouteRules = append(oRoute.RouteRules[:i], oRoute.RouteRules[i+1:]...)
		}

		if dirty {
			glog.V(6).Infof("Updating Route: %s %v", routeID, oRoute.RouteRules)
			err := cp.client.Networking().UpdateRouteTable(ctx, routeID, oRoute.RouteRules)
			if err != nil {
				return errors.WithStack(err)
			}
		}
	}

	return nil
}

func init() {
	subnetIDMutex = sync.Mutex{}
}

// subnet -> routetable
// subnets cannot change their route tables, so it's safe to cache
var (
	subnetIDs     map[string]string
	subnetIDMutex sync.Mutex
)

// SetSubnetIDs sets the subnetID map.
func SetSubnetIDs(new map[string]string, checkNil bool) {
	subnetIDMutex.Lock()
	defer subnetIDMutex.Unlock()
	if checkNil {
		if subnetIDs == nil {
			subnetIDs = new
		}
	} else {
		subnetIDs = new
	}

}

// GetSubnetIDs returns a map of subnet IDs.
func GetSubnetIDs() map[string]string {
	subnetIDMutex.Lock()
	defer subnetIDMutex.Unlock()
	if subnetIDs == nil {
		return nil
	}
	return subnetIDs
}

// routetable -> exists
// we may see new route tables if we see new subnets
var routeTableIds map[string]bool

// ListRoutes for OCI, discover all instances and their subnets and those route tables to iterate over
func (cp *CloudProvider) ListRoutes(ctx context.Context, clusterName string) ([]*cloudprovider.Route, error) {
	glog.V(6).Info("Listing Routes: ", clusterName)

	SetSubnetIDs(make(map[string]string), true)

	newRouteTableIds := make(map[string]bool)

	nodeList, err := cp.NodeLister.List(labels.Everything())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Get subnets and their route tables. Mark which route tables are in use.
	ipToNodeLookup := make(map[string]*v1.Node)
	for _, node := range nodeList {
		err := cp.mapNodeSubnetToRouteTable(ctx, node, newRouteTableIds)
		if err != nil {
			return nil, err
		}
		ip := util.NodeInternalIP(node)
		ipToNodeLookup[ip] = node
	}

	routeTableIds = newRouteTableIds

	// ultimately we need to find all the routes from all route tables
	// that are germane to this cluster, so if the destination cidr isn't
	// in the cluster range skip, also if we find mismatch destinations set
	// target to empty to invalidate and force a refresh
	var destinationsMap = make(map[string]string)

	for id := range routeTableIds {
		glog.V(6).Infof("Listing routes for id %s", id)

		oRoute, err := cp.client.Networking().GetRouteTable(ctx, id)
		if err != nil {
			glog.V(6).Infof("Failed to get routes for id %s", id)
			return nil, errors.WithStack(err)
		}

		for _, r := range oRoute.RouteRules {
			if !strings.Contains(*r.NetworkEntityId, "privateip") {
				glog.V(6).Infof("Skipping entry: %v", r)
				continue
			}

			err := cp.updateDestinationsMap(ctx, ipToNodeLookup, r, destinationsMap)
			if err != nil {
				return nil, err
			}
		}

	}

	var routes []*cloudprovider.Route
	for destinationCidr, target := range destinationsMap {
		routes = append(routes, &cloudprovider.Route{
			Name:            destinationCidr,
			TargetNode:      types.NodeName(target),
			Blackhole:       !(target == ""),
			DestinationCIDR: destinationCidr,
		})
	}

	glog.V(6).Infof("Listed Routes: %d", len(routes))
	return routes, nil
}

func (cp *CloudProvider) mapNodeSubnetToRouteTable(ctx context.Context, node *v1.Node, newRouteTableIds map[string]bool) error {
	instanceID := util.MapProviderIDToInstanceID(node.Spec.ProviderID)

	vnic, err := cp.client.Compute().GetPrimaryVNICForInstance(ctx, "", instanceID)
	if err != nil {
		return errors.WithStack(err)
	}

	// Should this be defaulted to true generally? Making a change to
	// 'ListRoutes' seems like an anti-pattern.
	if *vnic.SkipSourceDestCheck != true {
		glog.V(6).Info("Set SkipSourceDestCheck on ", vnic)
		err = cp.client.Networking().UpdateVnic(ctx, *vnic.Id, true)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	routeTable, ok := subnetIDs[*vnic.SubnetId]
	if !ok {
		subnet, err := cp.client.Networking().GetSubnet(ctx, *vnic.SubnetId)
		if err != nil {
			return errors.WithStack(err)
		}

		routeTable = *subnet.RouteTableId
		GetSubnetIDs()[*vnic.SubnetId] = routeTable
	}

	glog.V(6).Infof("Found Route Table id %s", routeTable)
	newRouteTableIds[routeTable] = true

	return nil
}

func (cp *CloudProvider) updateDestinationsMap(ctx context.Context, ipToNodeLookup map[string]*v1.Node,
	rule core.RouteRule, destinationsMap map[string]string) error {
	// TODO CACHE
	// Get the node ip to which these ips exist
	ipDetails, err := cp.client.Networking().GetIPFromOCID(ctx, *rule.NetworkEntityId)
	if err != nil {
		glog.V(6).Infof("Failed to resolve IP to OCID %v", *rule.NetworkEntityId)
		return errors.WithStack(err)
	}

	node, foundNode := ipToNodeLookup[*ipDetails.IpAddress]

	target, ok := destinationsMap[*rule.CidrBlock]

	glog.V(6).Infof("Evaluate entry: %v %s %s", node.Name, target, *rule.CidrBlock)

	if !ok {
		if !foundNode {
			destinationsMap[*rule.CidrBlock] = ""
		} else {
			destinationsMap[*rule.CidrBlock] = node.Name
		}
	} else {
		if target != node.Name {
			glog.V(6).Infof("Route Destination mismatch %s %s %s", target, node.Name, *rule.CidrBlock)
			destinationsMap[*rule.CidrBlock] = ""
		}
	}

	glog.V(6).Infof("Destination Map %s -> %s", *rule.CidrBlock, destinationsMap[*rule.CidrBlock])

	return nil
}
