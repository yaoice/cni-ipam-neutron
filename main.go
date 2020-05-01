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

package main

import (
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/yaoice/cni-ipam-neutron/backend/allocator"
	"github.com/yaoice/cni-ipam-neutron/backend/neutron"
)

func cmdAdd(args *skel.CmdArgs) error {
	ipamConf, confVersion, err := allocator.LoadIPAMConfig(args.StdinData, args.Args)
	if err != nil {
		return err
	}

	result := &current.Result{}

	if ipamConf.ResolvConf != "" {
		dns, err := parseResolvConf(ipamConf.ResolvConf)
		if err != nil {
			return err
		}
		result.DNS = *dns
	}

	store, err := neutron.New(ipamConf.Name, ipamConf)
	if err != nil {
		return err
	}
	defer store.Close()

	allocator := allocator.NewIPAllocator(store)

	ipConf, err := allocator.Get(args.ContainerID)
	if err != nil {
		return err
	}
	result.IPs = append(result.IPs, ipConf)

	result.Routes = ipamConf.Routes

	return types.PrintResult(result, confVersion)
}

func cmdDel(args *skel.CmdArgs) error {
	ipamConf, _, err := allocator.LoadIPAMConfig(args.StdinData, args.Args)
	if err != nil {
		return err
	}

	store, err := neutron.New(ipamConf.Name, ipamConf)
	if err != nil {
		return err
	}
	defer store.Close()

	ipAllocator := allocator.NewIPAllocator(store)

	err = ipAllocator.Release(args.ContainerID)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	skel.PluginMain(cmdAdd, cmdDel, version.All)
}
