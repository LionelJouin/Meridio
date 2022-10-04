#! /bin/sh
##
## meridio-e2e.sh --
##
##   Help script for Meridio e2e
##
## Commands;
##

prg=$(basename $0)
dir=$(dirname $0); dir=$(readlink -f $dir)
tmp=/tmp/${prg}_$$

die() {
    echo "ERROR: $*" >&2
    rm -rf $tmp
    exit 1
}
help() {
    grep '^##' $0 | cut -c3-
    rm -rf $tmp
    exit 0
}
test -n "$1" || help
echo "$1" | grep -qi "^help\|-h" && help

log() {
	echo "$prg: $*" >&2
}

##   env
##     Print environment.
##
cmd_env() {
	test -n "$KIND_CLUSTER_NAME" || export KIND_CLUSTER_NAME=kind
	test "$cmd" = "env" && set | grep -E '^(__.*|KIND_CLUSTER_NAME)='
}
check_kind() {
	test "$checked" = "yes" && return 0
	cmd_env
	kind get clusters | grep -q "^$KIND_CLUSTER_NAME" || \
		die "KinD cluster is not running; $KIND_CLUSTER_NAME"
	checked=yes
}
#   create_vlan_and_bridge <bridge> <iface> <vlan>
#     Create a bridge and a vlan interface on eth0.
#     NOTE: This must run as root on a KinD node!
cmd_create_vlan_and_bridge() {
	test -n "$2" || die "Parameter missing"
	whoami | grep -q root || die "Must run as root"
	local br=$1
	local iface=$2
	local vlan=$3
	local dev=$iface.$vlan

	ip link add name $br type bridge
	echo 0 > /proc/sys/net/ipv6/conf/$br/accept_dad
	echo 0 > /proc/sys/net/ipv4/conf/$br/rp_filter
	ip link set up dev $br

	ip link add link $iface name $dev type vlan id $vlan
	echo 0 > /proc/sys/net/ipv6/conf/$dev/accept_dad
	ip link set up dev $dev
	ip link set dev $dev master $br
}
# 
emit_nad() {
	cat <<EOF
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: meridio-100
spec:
  config: '{
    "cniVersion": "0.4.0",
    "type": "bridge",
    "bridge": "br1",
    "isGateway": true,
    "ipam": {
      "type": "node-annotation",
      "annotation": "meridio-br1"
    }
  }'
---
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: meridio-200
spec:
  config: '{
    "cniVersion": "0.4.0",
    "type": "bridge",
    "bridge": "br2",
    "isGateway": true,
    "ipam": {
      "type": "node-annotation",
      "annotation": "meridio-br2"
    }
  }'
EOF
}
##   multus_prepare
##     Prepare a started KinD cluster for e2e test with Multus.
##      - Install Multus
##      - Create bridges and vlan interfaces
##      - Install kubeconfig and configure node-annotation ipam on workers
##      - Annotate worker nodes with ranges
##      - Create NAD's "meridio-100" and "meridio-200" in namespace "default"
cmd_multus_prepare() {
	check_kind
	mkdir -p $tmp
	kubectl apply -f $dir/manifest/multus-install.yaml || die "Install Multus"
	local w i=0
	for w in $(kind --name=$KIND_CLUSTER_NAME get nodes); do
		echo $w | grep -q control-plane && continue

		kind get kubeconfig --internal | \
			docker exec -i $w tee /etc/kubernetes/kubeconfig > /dev/null
		echo "{ \"kubeconfig\": \"/etc/kubernetes/kubeconfig\", \"log\":\"/var/log/node-annotation\" }" | \
			docker exec -i $w tee /etc/cni/node-annotation.conf > /dev/null

		docker cp $dir/$prg $w:bin     # Copy myself
		docker exec $w /bin/$prg create_vlan_and_bridge br1 eth0 100
		annotate $w $i br1
		docker exec $w /bin/$prg create_vlan_and_bridge br2 eth0 200
		annotate $w $i br2

		i=$((i+1))
	done

	emit_nad > $tmp/nad
	kubectl apply -f $tmp/nad || die "Create NAD"
}
# annotate <worker> <index> <bridge>
annotate() {
	local w=$1
	local i=$2
	local br=$3
	
	local s e gw
	s=$((i*8+2))   # Start; Leave room for the GW
	e=$((i*8+6))
	gw=$((i*8+1))
	#test "$br" = "br2" && gw=$((i*8+7))

	# \"gateway\":\"100:100::$gw\" \"gateway\":\"169.254.100.$gw\"
	kubectl annotate node $w meridio-$br="\"ranges\": [
  [{ \"subnet\":\"100:100::/64\", \"rangeStart\":\"100:100::$s\" , \"rangeEnd\":\"100:100::$e\", \"gateway\":\"100:100::$gw\"}],
  [{ \"subnet\":\"169.254.100.0/24\", \"rangeStart\":\"169.254.100.$s\" , \"rangeEnd\":\"169.254.100.$e\", \"gateway\":\"169.254.100.$gw\"}]
]"
}

##
# Get the command
cmd=$1
shift
grep -q "^cmd_$cmd()" $0 $hook || die "Invalid command [$cmd]"

while echo "$1" | grep -q '^--'; do
    if echo $1 | grep -q =; then
	o=$(echo "$1" | cut -d= -f1 | sed -e 's,-,_,g')
	v=$(echo "$1" | cut -d= -f2-)
	eval "$o=\"$v\""
    else
	o=$(echo "$1" | sed -e 's,-,_,g')
	eval "$o=yes"
    fi
    shift
done
unset o v
long_opts=`set | grep '^__' | cut -d= -f1`

# Execute command
trap "die Interrupted" INT TERM
cmd_$cmd "$@"
status=$?
rm -rf $tmp
exit $status
