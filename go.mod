module github.com/nordix/meridio

go 1.16

require (
	github.com/antonfisher/nested-logrus-formatter v1.3.1
	github.com/edwarnicke/grpcfd v1.1.2
	github.com/golang/mock v1.5.0
	github.com/golang/protobuf v1.5.2
	github.com/google/nftables v0.0.0-20210916140115-16a134723a96
	github.com/google/uuid v1.2.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/networkservicemesh/api v1.4.0
	github.com/networkservicemesh/sdk v1.4.0
	github.com/networkservicemesh/sdk-sriov v1.4.0
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.17.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/spiffe/go-spiffe/v2 v2.0.0
	github.com/stretchr/testify v1.7.1
	github.com/vishvananda/netlink v1.1.1-0.20220118170537-d6b03fdeb845
	go.uber.org/goleak v1.1.12
	golang.org/x/sys v0.0.0-20220307203707-22a9840ba4d7
	golang.org/x/tools v0.1.9 // indirect
	google.golang.org/grpc v1.46.0
	google.golang.org/protobuf v1.28.0
	gopkg.in/yaml.v2 v2.4.0
	gorm.io/driver/sqlite v1.2.6
	gorm.io/gorm v1.22.3
	k8s.io/api v0.23.0
	k8s.io/apimachinery v0.23.0
	k8s.io/client-go v0.23.0
	sigs.k8s.io/controller-runtime v0.11.0
)

require (
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/jinzhu/now v1.1.3 // indirect
	github.com/open-policy-agent/opa v0.30.1 // indirect
	go.uber.org/atomic v1.8.0 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
	k8s.io/klog/v2 v2.40.1 // indirect
)

// https://github.com/Nordix/go-spiffe/commit/8666009a0f65d94256bbc20bf382f677a2baa3ce
// go get github.com/nordix/go-spiffe/v2@8666009a0f65d94256bbc20bf382f677a2baa3ce
// go mod edit -replace="github.com/spiffe/go-spiffe/v2@v2.0.0=github.com/nordix/go-spiffe/v2@v2.1.2-0.20220712151211-8666009a0f65"

replace github.com/spiffe/go-spiffe/v2 v2.0.0 => github.com/nordix/go-spiffe/v2 v2.1.2-0.20220712151211-8666009a0f65
