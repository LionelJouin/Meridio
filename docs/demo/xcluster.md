# Demo - xcluster - vlan

* [Kind - VLAN](readme.md) - Demo running on [Kind](https://kind.sigs.k8s.io/) using a vlan-forwarder to link the network service to an external host.
* [Kind - Static](static.md) - Demo running on [Kind](https://kind.sigs.k8s.io/) using a script to link the network service to an external host.
* [xcluster - VLAN](xcluster.md) - Demo running on [xcluster](https://github.com/Nordix/xcluster) using a vlan-forwarder to link the network service to an external host.
* [Kind - VLAN - Multi-trenches](multi-trenches.md) - Demo running 2 trenches on [Kind](https://kind.sigs.k8s.io/) using vlan-forwarders to link the network services to an external host.


## Installation

### Kubernetes cluster

Deploy a Kubernetes cluster with xcluster while using meridio ovl
```
unset __mem1
export __mem201=1024
export __mem202=1024
xc mkcdrom meridio; xc starts --nets_vm=0,1,2 --nvm=2 --mem=4096 --smp=4
```

### External host / External connectivity

Deploy external gateway and traffic generator  
prerequisite; Multus is ready (deployed by meridio ovl)
```
# default interface setup
helm install xcluster/ovl/meridio/helm/gateway --generate-name
# eth1:meridio--gateways, eth2:gateways--tg
helm install xcluster/ovl/meridio/helm/gateway --generate-name --set masterItf=eth1,tgMasterItf=eth2
```

### NSM

Deploy Spire
```
helm install docs/demo/deployments/spire/ --generate-name
```

Configure Spire
```
./docs/demo/scripts/spire-config.sh
```

Deploy NSM
```
helm install docs/demo/deployments/nsm-vlan/ --generate-name
```

### Meridio

Install Meridio  
Note: the external interface must match the one used by gateway "routers"
```
# ipv4
helm install deployments/helm/ --generate-name --set vlan.interface=eth1
# ipv6
helm install deployments/helm/ --generate-name --set ipFamily=ipv6 
# dualstack
helm install deployments/helm/ --generate-name --set ipFamily=dualstack 
```

## Traffic

Connect to the Traffic Generator POD
```
# exec into traffic generator POD
kubectl exec -ti tg -- bash
```

Ping
```
ping 20.0.0.1 -c 3
ping 2000::1 -c 3

```

Generate traffic
```
# ipv4
/ctraffic -address 20.0.0.1:5000 -nconn 400 -rate 100 -monitor -stats all > v4traffic.json
# ipv6
/ctraffic -address [2000::1]:5000 -nconn 400 -rate 100 -monitor -stats all > v4traffic.json
```

Verification
```
/ctraffic -analyze hosts -stat_file v4traffic.json
```

## Scaling

Scale-In/Scale-Out target
```
kubectl scale deployment target --replicas=5
```

Scale-In/Scale-Out load-balancer
```
kubectl scale deployment load-balancer --replicas=1
```
