// Copyright 2015 CNI authors
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

package neutron

import (
	"fmt"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
	"log"
	"net"
	"sync"

	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"
	"github.com/yaoice/cni-ipam-neutron/backend/allocator"

	"github.com/yaoice/cni-ipam-neutron/backend"
)

// Store is a simple disk-backed store that creates one file per IP
// address in a given directory. The contents of the file are the container ID.
type Store struct {
	lock          sync.RWMutex
	NetworkClient *gophercloud.ServiceClient
	Networks      []string
}

// Store implements the Store interface
var _ backend.Store = &Store{}

func New(name string, ipamConfig *allocator.IPAMConfig) (backend.Store, error) {
	networkClient, err := connectStore(ipamConfig.OpenStackConf)
	if err != nil {
		return nil, err
	}
	if len(ipamConfig.NeutronConf.Networks) == 0 {
		return nil, fmt.Errorf("neutron networks is none")
	}
	// write values in Store object
	store := &Store{
		NetworkClient: networkClient,
		Networks:      ipamConfig.NeutronConf.Networks,
	}
	return store, nil
}

func (s *Store) Reserve(id string) (*net.IPNet, net.IP, error) {
	network, err := networks.Get(s.NetworkClient, s.Networks[0]).Extract()
	if err != nil {
		return nil, nil, err
	}
	if len(network.Subnets) == 0 {
		return nil, nil, fmt.Errorf("eutron subnets is none")
	}

	port, err := ports.Create(s.NetworkClient, ports.CreateOpts{
		NetworkID: network.ID,
		Name:      id,
		AdminStateUp: getBoolPointer(true),
	}).Extract()
	if err != nil {
		log.Printf("create neutron port %s err: %v", id, err.Error())
		return nil, nil, err
	}

	if len(port.FixedIPs) == 0 {
		return nil, nil, fmt.Errorf("port doesn't have fixed ip")
	}

	subnetID := port.FixedIPs[0].SubnetID
	subnet, err := subnets.Get(s.NetworkClient, subnetID).Extract()
	if err != nil {
		log.Printf("get neutron subnet %s err: %v", subnetID, err.Error())
		return nil, nil, err
	}

	gw := net.ParseIP(subnet.GatewayIP)
	currentIP := net.ParseIP(port.FixedIPs[0].IPAddress)

	_, ipnet, err := net.ParseCIDR(subnet.CIDR)
	if err != nil {
		log.Printf("parse neutron subnet %s err: %v", subnetID, err.Error())
		return nil, nil, err
	}

	return &net.IPNet{IP: currentIP, Mask: ipnet.Mask}, gw, nil
}

// N.B. This function eats errors to be tolerant and
// release as much as possible
func (s *Store) ReleaseByID(id string) error {
	var portID string
	portList, err := ports.List(s.NetworkClient, ports.ListOpts{Name: id, Limit: 1}).AllPages()
	if err != nil {
		log.Printf("get neutron port name %s err: %v", id, err.Error())
		return err
	}

	portSlice, err := ports.ExtractPorts(portList)
	if err != nil {
		return err
	}

	if len(portSlice) < 1 {
		log.Println("Already deleted neutron port")
		return nil
	}

	for _, port := range portSlice {
		portID = port.ID
	}

	if err := ports.Delete(s.NetworkClient, portID).ExtractErr(); err != nil {
		log.Printf("delete neutron port %s err: %v", portID, err.Error())
		return err
	}
	return nil
}

func (s *Store) Close() error {
	// stub we don't need close anything
	return nil
}

func (s *Store) Lock() error {
	s.lock.Lock()
	return nil
}

func (s *Store) Unlock() error {
	s.lock.Unlock()
	return nil
}
