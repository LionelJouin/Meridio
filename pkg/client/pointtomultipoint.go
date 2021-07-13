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

package client

import (
	"context"
	"fmt"
	"io"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/api/pkg/api/registry"
	"github.com/sirupsen/logrus"
)

// PointToMultiPoint manages the requests and closes of a point to multi point connection
// The network service endpoints of the network service precised in the request will be
// monitor and a point to point connection will be open for each of them.
type PointToMultiPoint struct {
	// NetworkServiceClient
	NetworkServiceClient networkservice.NetworkServiceClient
	// NetworkServiceEndpointRegistryClient
	NetworkServiceEndpointRegistryClient registry.NetworkServiceEndpointRegistryClient
	// ConnectionChannel returns an event when connection is established or closed
	ConnectionChannel chan<- *ConnectionEvent
	// pointToPointMap contains the list of point to point connections managed
	pointToPointMap *pointToPointMap
	// baseRequest is the request used as base for all point to point connection
	baseRequest *networkservice.NetworkServiceRequest
	// networkServiceDiscoveryStream returns the new/removed network service endpoints
	networkServiceDiscoveryStream registry.NetworkServiceEndpointRegistry_FindClient
	// context
	context context.Context
	// cancel function to cancel the request
	cancel context.CancelFunc
	// connectionIndex is the index of the last connection
	connectionIndex int
}

// Request opens a stream using the NSM API to monitor the network service endpoint (NSE)
// which belong to the network service. For each NSE, a new point to point connection is
// opened. The request parameter will be used as a base for each point to point connection to the
// network service endpoints. The ID will be modified, and the NetworkServiceEndpointName
// will be filled with a specific network service endpoint name. For each point to point connection
// established, a connection event will be send via the connection channel.
func (ptmp *PointToMultiPoint) Request(ctx context.Context, request *networkservice.NetworkServiceRequest) error {
	ptmp.baseRequest = request
	ptmp.context, ptmp.cancel = context.WithCancel(ctx)
	var err error
	// Prepare query
	networkServiceEndpoint := &registry.NetworkServiceEndpoint{
		NetworkServiceNames: []string{ptmp.baseRequest.Connection.NetworkService},
	}
	networkServiceEndpointQuery := &registry.NetworkServiceEndpointQuery{
		NetworkServiceEndpoint: networkServiceEndpoint,
		Watch:                  true,
	}
	// Get stream from NSM API
	ptmp.networkServiceDiscoveryStream, err = ptmp.NetworkServiceEndpointRegistryClient.Find(ptmp.context, networkServiceEndpointQuery)
	if err != nil {
		return err
	}
	return ptmp.receiveNeworkServiceEndpoints()
}

// Close stops the client stream watching the network service endpoints and closes all point to
// point connections managed. For each point to point connection closed, a connection event will
// be send via the connection channel.
func (ptmp *PointToMultiPoint) Close(ctx context.Context) error {
	if ptmp.cancel != nil {
		ptmp.cancel() // Cancel pending request
	}
	ptmp.context, ptmp.cancel = context.WithCancel(ctx)
	for _, networkServiceEndpointName := range ptmp.pointToPointMap.list() {
		ptmp.closePointToPoint(networkServiceEndpointName)
	}
	return nil
}

func (ptmp *PointToMultiPoint) receiveNeworkServiceEndpoints() error {
	for {
		networkServiceEndpoint, err := ptmp.networkServiceDiscoveryStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// If the expiration time is null, the network service endpoint has been removed.
		if !expirationTimeIsNull(networkServiceEndpoint.ExpirationTime) {
			go ptmp.createPointToPoint(networkServiceEndpoint.Name)
		} else {
			go ptmp.closePointToPoint(networkServiceEndpoint.Name)
		}
	}
	return nil
}

func (ptmp *PointToMultiPoint) createPointToPoint(networkServiceEndpointName string) {
	request := copyRequest(ptmp.baseRequest)
	request.Connection.NetworkServiceEndpointName = networkServiceEndpointName
	request.Connection.Id = fmt.Sprintf("%s-%d", request.Connection.Id, ptmp.connectionIndex)
	ptmp.connectionIndex++
	pointToPoint := &PointToPoint{
		NetworkServiceClient: ptmp.NetworkServiceClient,
		MaxRetry:             5,
		Timeout:              10000,
		Delay:                5000,
		ConnectionChannel:    ptmp.ConnectionChannel,
	}
	success := ptmp.pointToPointMap.add(networkServiceEndpointName, pointToPoint)
	if !success {
		return
	}
	err := pointToPoint.Request(ptmp.context, request)
	if err != nil {
		logrus.Errorf("error requesting point to point, request: %v, error: %v", request, err)
		return
	}
}

func (ptmp *PointToMultiPoint) closePointToPoint(networkServiceEndpointName string) {
	pointToPoint := ptmp.pointToPointMap.delete(networkServiceEndpointName)
	if pointToPoint == nil {
		return
	}
	err := pointToPoint.Close(ptmp.context)
	if err != nil {
		logrus.Errorf("error closing point to point, error: %v", err)
	}
}

// NewPointToMultiPoint creates and initializes a new PointToMultiPoint
func NewPointToMultiPoint(networkServiceClient networkservice.NetworkServiceClient,
	networkServiceEndpointRegistryClient registry.NetworkServiceEndpointRegistryClient,
	connectionChannel chan<- *ConnectionEvent) *PointToMultiPoint {
	pointToMultiPoint := &PointToMultiPoint{
		NetworkServiceClient:                 networkServiceClient,
		NetworkServiceEndpointRegistryClient: networkServiceEndpointRegistryClient,
		ConnectionChannel:                    connectionChannel,
		pointToPointMap:                      newPointToPointMap(),
		connectionIndex:                      0,
	}
	return pointToMultiPoint
}
