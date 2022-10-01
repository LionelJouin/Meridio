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
#   create_vlan_and_bridge
#     Create a bridge and a vlan interface on eth0.
#     NOTE: This must run as root on a KinD node!
cmd_create_vlan_and_bridge() {
	test -n "$2" || die "No tag"
	whoami | grep -q root || die "Must run as root"
	local br=$1
	local iface=eth0.$2

	ip link add name $br type bridge
	echo 0 > /proc/sys/net/ipv6/conf/$br/accept_dad
	ip link set up dev $br

	ip link add link eth0 name $iface type vlan id 100
	echo 0 > /proc/sys/net/ipv6/conf/$iface/accept_dad
	ip link set up dev $iface
	ip link set dev $iface master $br
}
# 
emit_nad() {
	cat <<EOF
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: $1
spec:
  config: '{
    "cniVersion": "0.4.0",
    "type": "bridge",
    "bridge": "$2",
    "isGateway": true,
    "ipam": {
      "type": "node-annotation",
      "annotation": "meridio/$2"
    }
  }'
EOF
}
##   multus_prepare
##     Prepare a started KinD cluster for e2e test with Multus.
##      - Install Multus
##      - Create bridges and vlan interfaces
##      - Install kubeconfig and configure node-annoration ipam on workers
##      - Annotate worker nodes with ranges
##      - Create NAD's in namespace "default"
cmd_multus_prepare() {
	check_kind
	mkdir -p $tmp
	kubectl apply -f $dir/manifest/multus-install.yaml || die "Install Multus"
	local w i=0 s e br=br1
	for w in $(kind --name=$KIND_CLUSTER_NAME get nodes); do
		echo $w | grep -q control-plane && continue
		docker cp $dir/$prg $w:bin     # Copy myself
		docker exec $w /bin/$prg create_vlan_and_bridge $br 100  # Run myself
		kind get kubeconfig --internal | \
			docker exec -i $w tee /etc/kubernetes/kubeconfig > /dev/null
		echo "{ \"kubeconfig\": \"/etc/kubernetes/kubeconfig\" }" | \
			docker exec -i $w tee /etc/cni/node-annotation.conf > /dev/null
		s=$((i*8))
		e=$((s+7))
		kubectl annotate node $w meridio/$br="\"ranges\": [
  [{ \"subnet\":\"100:100::/120\", \"rangeStart\":\"100:100::$s\" , \"rangeEnd\":\"100:100::$e\"}],
  [{ \"subnet\":\"169.254.100.0/24\", \"rangeStart\":\"169.254.100.$s\" , \"rangeEnd\":\"169.254.100.$e\"}]
]"
		i=$((i+1))
	done
	emit_nad meridio-100 $br > $tmp/nad
	kubectl apply -f $tmp/nad || die "Create NAD"
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
