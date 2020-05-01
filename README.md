# neutron IP address management plugin

Inspired by [host-local](https://github.com/containernetworking/plugins/tree/master/plugins/ipam/host-local) and [cni-ipam-consul](https://github.com/logingood/cni-ipam-consul) cni plugins. The etcd IPAM allocates IPv4 and IPv6 addresses out of a specified address range. Optionally,
it can include a DNS configuration from a `resolv.conf` file on the host.

## Overview

neutron IPAM plugin allocates ip addresses from OpenStack neutron services.

## Example configurations

This example configuration returns 1 IP addresses.

```json
{
	"ipam": {
		"name": "myetcd-ipam",
		"type": "ipam-neutron",
		"openstackConf": {
			"username": "admin",
			"password": "c111f3c44f352e91ce76",
			"project": "admin",
			"domain": "default",
            "authUrl": "http://1.1.1.1:35357/v3"
		},
        "neutronConf": {
            "networks": ["782ec9ac-44f9-4318-8c67-a2fed2ccca4f"]
        },
		"routes": [
			{ "dst": "0.0.0.0/0" },
			{ "dst": "192.168.0.0/16", "gw": "10.10.5.1" }
		]
	}
}
```

We can test it out on the command-line:

```bash
# awk BEGIN{RS=EOF}'{gsub(/\n/,"");print}' /tmp/xxx.json
# Test add operation
$ echo '{"cniVersion": "0.3.1","name": "examplenet","ipam": {"name": "myetcd-ipam","type": "ipam-neutron","openstackConf": {"username": "admin","password": "c111f3c44f352e91ce76","project": "admin","domain": "default","authUrl": "http://10.125.224.21:35357/v3"},"neutronConf": {"networks": ["782ec9ac-44f9-4318-8c67-a2fed2ccca4f"]}}}' | CNI_COMMAND=ADD CNI_CONTAINERID=example CNI_NETNS=/dev/null CNI_IFNAME=dummy0 CNI_PATH=. ./cni-ipam-neutron

# Test del operation 
$ echo '{"cniVersion": "0.3.1","name": "examplenet","ipam": {"name": "myetcd-ipam","type": "ipam-neutron","openstackConf": {"username": "admin","password": "c111f3c44f352e91ce76","project": "admin","domain": "default","authUrl": "http://10.125.224.21:35357/v3"},"neutronConf": {"networks": ["782ec9ac-44f9-4318-8c67-a2fed2ccca4f"]}}}' | CNI_COMMAND=DEL CNI_CONTAINERID=example CNI_NETNS=/dev/null CNI_IFNAME=dummy0 CNI_PATH=. ./cni-ipam-neutron

```

```json
{
    "ips": [
        {
            "version": "4",
            "address": "203.0.113.2/24",
            "gateway": "203.0.113.1"
        }
    ],
    "dns": {}
}
```

## Network configuration reference

* `type` (string, required): "ipam-neutron".
* `routes` (string, optional): list of routes to add to the container namespace. Each route is a dictionary with "dst" and optional "gw" fields. If "gw" is omitted, value of "gateway" will be used.
* `resolvConf` (string, optional): Path to a `resolv.conf` on the host to parse and return as the DNS configuration
* `neutronConf`, (array, required, nonempty) an array of arrays of network id:
* `openstackConf`, OpenStack Auth Configuration 
  * `username` (string, required): The usernmae of auth
  * `password` (string, required): The password of auth(use scripts/encrypt.go to encrypt)
  * `project` (string, required): The project of auth
  * `domain` (string, required): The domain of auth(keystone v3)
  * `authUrl` (string, required): The auth url

## Supported arguments
The following [CNI_ARGS](https://github.com/containernetworking/cni/blob/master/SPEC.md#parameters) are supported:

* `ip`: request a specific IP address from a subnet.

The following [args conventions](https://github.com/containernetworking/cni/blob/master/CONVENTIONS.md) are supported:

* `ips` (array of strings): A list of custom IPs to attempt to allocate

## KVs

The kvs written by this cni plugin can query by the script:

```bash
$ neutron port-list | grep example
```

