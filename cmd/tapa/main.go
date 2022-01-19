/*
Copyright (c) 2021 Nordix Foundation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/kelseyhightower/envconfig"
	"github.com/networkservicemesh/api/pkg/api/networkservice"
	kernelmech "github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/kernel"
	vfiomech "github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/vfio"
	"github.com/networkservicemesh/sdk-sriov/pkg/networkservice/common/mechanisms/vfio"
	sriovtoken "github.com/networkservicemesh/sdk-sriov/pkg/networkservice/common/token"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/authorize"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/mechanisms"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/mechanisms/kernel"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/mechanisms/sendfd"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/refresh"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/serialize"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/updatepath"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/chain"
	"github.com/networkservicemesh/sdk/pkg/networkservice/utils/metadata"
	tapAPI "github.com/nordix/meridio/api/ambassador/v1"
	"github.com/nordix/meridio/pkg/ambassador/tap"
	"github.com/nordix/meridio/pkg/health"
	linuxKernel "github.com/nordix/meridio/pkg/kernel"
	"github.com/nordix/meridio/pkg/nsm"
	"github.com/nordix/meridio/pkg/nsm/interfacename"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	grpcHealth "google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer cancel()

	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)

	var config Config
	err := envconfig.Process("nsm", &config)
	if err != nil {
		logrus.Fatalf("%v", err)
	}

	netUtils := &linuxKernel.KernelUtils{}

	healthChecker, err := health.NewChecker(health.WithCtx(ctx))
	if err != nil {
		logrus.Fatalf("Unable to create Health checker: %v", err)
	}
	go func() {
		err := healthChecker.Start()
		if err != nil {
			logrus.Fatalf("Unable to start Health checker: %v", err)
		}
	}()

	apiClientConfig := &nsm.Config{
		Name:             config.Name,
		ConnectTo:        config.ConnectTo,
		DialTimeout:      config.DialTimeout,
		RequestTimeout:   config.RequestTimeout,
		MaxTokenLifetime: config.MaxTokenLifetime,
	}
	nsmAPIClient := nsm.NewAPIClient(ctx, apiClientConfig)
	networkServiceClient := chain.NewNetworkServiceClient(
		updatepath.NewClient(config.Name),
		serialize.NewClient(),
		refresh.NewClient(ctx),
		metadata.NewClient(),
		sriovtoken.NewClient(),
		mechanisms.NewClient(map[string]networkservice.NetworkServiceClient{
			vfiomech.MECHANISM:   chain.NewNetworkServiceClient(vfio.NewClient()),
			kernelmech.MECHANISM: chain.NewNetworkServiceClient(kernel.NewClient()),
		}),
		interfacename.NewClient("nsc", &interfacename.RandomGenerator{}),
		// interfaceMonitorClient,
		authorize.NewClient(),
		sendfd.NewClient(),
		networkservice.NewNetworkServiceClient(nsmAPIClient.GRPCClient),
	)

	if err := os.RemoveAll(config.Socket); err != nil {
		log.Fatal(err)
	}
	lis, err := net.Listen("unix", config.Socket)
	if err != nil {
		log.Fatal(err)
	}
	s := grpc.NewServer()

	// targetConfig := tap.NewConfig(config.ConfigMapName, config.NSPServiceName, config.NSPServicePort, netUtils, apiClientConfig)
	ambassador, err := tap.New(config.Name, config.Namespace, networkServiceClient, config.NSPServiceName, config.NSPServicePort, netUtils)
	if err != nil {
		logrus.Fatalf("Error creating new tap ambassador: %v", err)
	}
	healthServer := grpcHealth.NewServer()
	grpc_health_v1.RegisterHealthServer(s, healthServer)
	tapAPI.RegisterTapServer(s, ambassador)

	if err := s.Serve(lis); err != nil {
		logrus.Errorf("TAP Ambassador: failed to serve: %v", err)
	}

	// defer func() {
	// 	err = ambassador.Delete()
	// 	if err != nil {
	// 		logrus.Fatalf("Error deleting ambassador: %v", err)
	// 	}
	// }()

	// go func() {
	// 	err = ambassador.Start(ctx)
	// 	if err != nil {
	// 		logrus.Fatalf("Error starting ambassador: %v", err)
	// 	}
	// }()

	<-ctx.Done()
}
