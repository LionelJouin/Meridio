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

package trench

import (
	"context"
	"fmt"
	"sync"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	nspAPI "github.com/nordix/meridio/api/nsp/v1"
	"github.com/nordix/meridio/pkg/ambassador/tap/conduit"
	"github.com/nordix/meridio/pkg/ambassador/tap/types"
	"github.com/nordix/meridio/pkg/networking"
	"github.com/nordix/meridio/pkg/nsp"
	"github.com/nordix/meridio/pkg/security/credentials"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type Trench struct {
	Trench                     *nspAPI.Trench
	TargetName                 string
	Namespace                  string
	ConfigurationManagerClient nspAPI.ConfigurationManagerClient
	TargetRegistryClient       nspAPI.TargetRegistryClient
	NetworkServiceClient       networkservice.NetworkServiceClient
	StreamRegistry             types.Registry
	NetUtils                   networking.Utils
	nspConn                    *grpc.ClientConn
	conduits                   []types.Conduit
	mu                         sync.Mutex
}

func New(trench *nspAPI.Trench,
	targetName string,
	namespace string,
	networkServiceClient networkservice.NetworkServiceClient,
	streamRegistry types.Registry,
	nspServiceName string,
	nspServicePort int,
	netUtils networking.Utils) (*Trench, error) {

	t := &Trench{
		TargetName:           targetName,
		Namespace:            namespace,
		Trench:               trench,
		NetworkServiceClient: networkServiceClient,
		StreamRegistry:       streamRegistry,
		NetUtils:             netUtils,
		conduits:             []types.Conduit{},
	}

	var err error
	t.nspConn, err = t.connectNSPService(context.TODO(), nspServiceName, nspServicePort)
	if err != nil {
		return nil, err
	}
	t.ConfigurationManagerClient = nspAPI.NewConfigurationManagerClient(t.nspConn)
	t.TargetRegistryClient = nspAPI.NewTargetRegistryClient(t.nspConn)
	return t, nil
}

func (t *Trench) Delete(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	var errFinal error
	var err error
	// todo: delete streams in conduits
	err = t.disconnectConduits(ctx)
	if err != nil {
		errFinal = fmt.Errorf("%w; %v", errFinal, err) // todo
	}
	t.conduits = []types.Conduit{}
	err = t.nspConn.Close()
	if err != nil {
		errFinal = fmt.Errorf("%w; %v", errFinal, err) // todo
	}
	t.ConfigurationManagerClient = nil
	t.TargetRegistryClient = nil
	t.nspConn = nil
	return errFinal
}

func (t *Trench) AddConduit(ctx context.Context, cndt *nspAPI.Conduit) (types.Conduit, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	c := t.getConduit(cndt)
	if c != nil {
		return c, nil
	}
	c, err := conduit.New(cndt,
		t.TargetName,
		t.Namespace,
		t.ConfigurationManagerClient,
		t.TargetRegistryClient,
		t.NetworkServiceClient,
		t.StreamRegistry,
		t.NetUtils)
	if err != nil {
		return nil, err
	}
	t.conduits = append(t.conduits, c)
	return c, c.Connect(ctx)
}

func (t *Trench) RemoveConduit(ctx context.Context, cndt *nspAPI.Conduit) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	index := t.getConduitIndex(cndt)
	if index < 0 {
		return nil
	}
	c := t.conduits[index]
	err := c.Disconnect(ctx)
	t.conduits = append(t.conduits[:index], t.conduits[index+1:]...)
	return err
}

func (t *Trench) GetConduits() []types.Conduit {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.conduits
}

func (t *Trench) GetConduit(conduit *nspAPI.Conduit) types.Conduit {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.getConduit(conduit)
}

func (t *Trench) Equals(trench *nspAPI.Trench) bool {
	return t.Trench.Equals(trench)
}

// todo: fix wait for ready
func (t *Trench) connectNSPService(ctx context.Context, nspServiceName string, nspServicePort int) (*grpc.ClientConn, error) {
	service := nsp.GetService(nspServiceName, t.Trench.GetName(), t.Namespace, nspServicePort)
	logrus.Infof("Connect to NSP Service: %v", service)
	return grpc.Dial(service,
		grpc.WithTransportCredentials(
			credentials.GetClient(ctx),
		),
		grpc.WithDefaultCallOptions(
			grpc.WaitForReady(true),
		))
}

func (t *Trench) connectConduits(ctx context.Context) error {
	var errFinal error
	for _, conduit := range t.conduits {
		err := conduit.Connect(ctx)
		if err != nil {
			errFinal = fmt.Errorf("%w; %v", errFinal, err) // todo
		}
	}
	return errFinal
}

func (t *Trench) disconnectConduits(ctx context.Context) error {
	var errFinal error
	for _, conduit := range t.conduits {
		err := conduit.Disconnect(ctx)
		if err != nil {
			errFinal = fmt.Errorf("%w; %v", errFinal, err) // todo
		}
	}
	return errFinal
}

func (t *Trench) getConduitIndex(cndt *nspAPI.Conduit) int {
	for i, c := range t.conduits {
		equal := c.Equals(cndt)
		if equal {
			return i
		}
	}
	return -1
}

func (t *Trench) getConduit(cndt *nspAPI.Conduit) types.Conduit {
	index := t.getConduitIndex(cndt)
	if index < 0 {
		return nil
	}
	return t.conduits[index]
}
