# Multus in Meridio

The Frontend (FE) must have an interface to the external network. The
interface can be created in several ways. In Meridio we use either the
NSM `nse-remote-vlan` service endpoint or [Multus](
https://github.com/k8snetworkplumbingwg/multus-cni).

<img src="resources/multus-interface.svg" width="50%" />

It is important to note that the FE is unaware of how the interface is
created, it just uses it. If Multus is used the network
service client (nsc) is not needed and is removed from the
load-balancer POD.



## Multiple FEs using the same external network

To connect several FEs to the same external network we can use a Linux
`bridge`.

<img src="resources/multus-bridge.svg" width="60%" />

With Multus this is easily done with the [bridge cni-plugin](
https://www.cni.dev/plugins/current/main/bridge/). The external
intercace must be created and attached to the bridge. An example with vlan;

```
ip link add link eth0 name eth.100 type vlan id 100
echo 0 > /proc/sys/net/ipv6/conf/eth.100/accept_dad
ip link set up dev eth.100
ip link set dev eth0.100 master cbr2
```

It doesn't matter if the load-balancer PODs belong to different trenches.


