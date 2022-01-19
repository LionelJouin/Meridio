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

package conduit

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/cls"
	kernelmech "github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/kernel"
	"github.com/networkservicemesh/api/pkg/api/networkservice/payload"
	nspAPI "github.com/nordix/meridio/api/nsp/v1"
	"github.com/nordix/meridio/pkg/ambassador/tap/stream"
	"github.com/nordix/meridio/pkg/ambassador/tap/types"
	"github.com/nordix/meridio/pkg/conduit"
	"github.com/nordix/meridio/pkg/networking"
	"github.com/sirupsen/logrus"
)

type Conduit struct {
	TargetName                 string
	Namespace                  string
	Conduit                    *nspAPI.Conduit
	ConfigurationManagerClient nspAPI.ConfigurationManagerClient
	TargetRegistryClient       nspAPI.TargetRegistryClient
	NetworkServiceClient       networkservice.NetworkServiceClient
	Configuration              *Configuration
	StreamRegistry             types.Registry
	NetUtils                   networking.Utils
	connection                 *networkservice.Connection
	streams                    []types.Stream
	mu                         sync.Mutex
	vips                       []*virtualIP
	tableID                    int
	cancelConnect              context.CancelFunc
}

func New(conduit *nspAPI.Conduit,
	targetName string,
	namespace string,
	configurationManagerClient nspAPI.ConfigurationManagerClient,
	targetRegistryClient nspAPI.TargetRegistryClient,
	networkServiceClient networkservice.NetworkServiceClient,
	streamRegistry types.Registry,
	netUtils networking.Utils) (types.Conduit, error) {
	c := &Conduit{
		TargetName:                 targetName,
		Namespace:                  namespace,
		Conduit:                    conduit,
		ConfigurationManagerClient: configurationManagerClient,
		TargetRegistryClient:       targetRegistryClient,
		NetworkServiceClient:       networkServiceClient,
		StreamRegistry:             streamRegistry,
		NetUtils:                   netUtils,
		connection:                 nil,
		streams:                    []types.Stream{},
		vips:                       []*virtualIP{},
		tableID:                    1,
	}
	c.Configuration = NewConfiguration(c, c.Conduit.GetTrench(), c.ConfigurationManagerClient)
	return c, nil
}

func (c *Conduit) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	ctx, c.cancelConnect = context.WithCancel(context.TODO())
	nsName := conduit.GetNetworkServiceNameWithProxy(c.Conduit.GetName(), c.Conduit.GetTrench().GetName(), c.Namespace)
	logrus.Infof("Connect to conduit: %v", c.Conduit)
	go func() {
		for { // Todo: retry
			if ctx.Err() != nil {
				return
			}
			connection, err := c.NetworkServiceClient.Request(ctx,
				&networkservice.NetworkServiceRequest{
					Connection: &networkservice.Connection{
						Id:             fmt.Sprintf("%s-%s-%d", c.TargetName, nsName, 0),
						NetworkService: nsName,
						Labels:         make(map[string]string),
						Payload:        payload.Ethernet,
					},
					MechanismPreferences: []*networkservice.Mechanism{
						{
							Cls:  cls.LOCAL,
							Type: kernelmech.MECHANISM,
						},
					},
				})
			if err != nil {
				logrus.Infof("NSM NS %v connection failed for conduit %v: %v", c.connection.GetNetworkService(), c.Conduit, err)
				continue
			}
			c.mu.Lock() // concurrency issue could happen if a stream is added at the same time, it will not be opened
			c.connection = connection
			c.mu.Unlock()
			logrus.Infof("NSM NS %v connected for conduit %v: %v", c.connection.GetNetworkService(), c.Conduit, c.connection)
			break
		}
		go c.Configuration.WatchVIPs(context.TODO())
		c.openStreams(ctx)
	}()
	return nil
}

func (c *Conduit) Disconnect(ctx context.Context) error {
	if c.cancelConnect != nil {
		c.cancelConnect()
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	logrus.Infof("Disconnect from conduit: %v", c.Conduit)
	var errFinal error
	err := c.closeStreams(ctx)
	if err != nil {
		errFinal = fmt.Errorf("%w; %v", errFinal, err) // todo
	}
	if c.isConnected() {
		logrus.Infof("Close NSM NS: %v - %v", c.connection.GetNetworkService(), c.connection)
		_, err = c.NetworkServiceClient.Close(ctx, c.connection)
		if err != nil {
			errFinal = fmt.Errorf("%w; %v", errFinal, err) // todo
		}
		c.connection = nil
	}
	c.Configuration.Delete()
	c.deleteVIPs(c.vips)
	c.tableID = 1
	return errFinal
}

func (c *Conduit) AddStream(ctx context.Context, strm *nspAPI.Stream) (types.Stream, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	s := c.getStream(strm)
	if s != nil {
		return nil, errors.New("this stream is already opened")
	}
	s, err := stream.New(strm, c.TargetRegistryClient, c.StreamRegistry, stream.MaxNumberOfTargets, c)
	if err != nil {
		return nil, err
	}
	c.streams = append(c.streams, s)
	if c.isConnected() {
		return s, s.Open(ctx)
	}
	return s, nil
}

func (c *Conduit) RemoveStream(ctx context.Context, strm *nspAPI.Stream) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	index := c.getStreamIndex(strm)
	if index < 0 {
		return nil
	}
	s := c.streams[index]
	err := s.Close(ctx)
	c.streams = append(c.streams[:index], c.streams[index+1:]...)
	return err
}

func (c *Conduit) GetStreams() []types.Stream {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.streams
}

func (c *Conduit) Equals(conduit *nspAPI.Conduit) bool {
	return c.Conduit.Equals(conduit)
}

func (c *Conduit) GetIPs() []string {
	if c.connection != nil {
		return c.connection.GetContext().GetIpContext().GetSrcIpAddrs()
	}
	return []string{}
}

func (c *Conduit) SetVIPs(vips []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.isConnected() {
		return nil
	}
	currentVIPs := make(map[string]*virtualIP)
	for _, vip := range c.vips {
		currentVIPs[vip.prefix] = vip
	}
	for _, vip := range vips {
		if _, ok := currentVIPs[vip]; !ok {
			newVIP, err := newVirtualIP(vip, c.tableID, c.NetUtils)
			if err != nil {
				logrus.Errorf("SimpleTarget: Error adding SourceBaseRoute: %v", err) // todo: err handling
				continue
			}
			c.tableID++
			c.vips = append(c.vips, newVIP)
			for _, nexthop := range c.getGateways() {
				err = newVIP.AddNexthop(nexthop)
				if err != nil {
					logrus.Errorf("Client: Error adding nexthop: %v", err) // todo: err handling
				}
			}
		}
		delete(currentVIPs, vip)
	}
	// delete remaining vips
	vipsSlice := []*virtualIP{}
	for _, vip := range currentVIPs {
		vipsSlice = append(vipsSlice, vip)
	}
	c.deleteVIPs(vipsSlice)
	return nil
}

func (c *Conduit) getStreamIndex(strm *nspAPI.Stream) int {
	for i, s := range c.streams {
		equal := s.Equals(strm)
		if equal {
			return i
		}
	}
	return -1
}

func (c *Conduit) getStream(strm *nspAPI.Stream) types.Stream {
	index := c.getStreamIndex(strm)
	if index < 0 {
		return nil
	}
	return c.streams[index]
}

func (c *Conduit) openStreams(ctx context.Context) error {
	var errFinal error
	for _, stream := range c.streams {
		err := stream.Open(ctx)
		if err != nil {
			errFinal = fmt.Errorf("%w; %v", errFinal, err) // todo
		}
	}
	return errFinal
}

func (c *Conduit) closeStreams(ctx context.Context) error {
	var errFinal error
	for _, stream := range c.streams {
		err := stream.Open(ctx)
		if err != nil {
			errFinal = fmt.Errorf("%w; %v", errFinal, err) // todo
		}
	}
	return errFinal
}

func (c *Conduit) isConnected() bool {
	return c.connection != nil
}

func (c *Conduit) deleteVIPs(vips []*virtualIP) {
	vipsMap := make(map[string]*virtualIP)
	for _, vip := range vips {
		vipsMap[vip.prefix] = vip
	}
	for index := 0; index < len(c.vips); index++ {
		vip := c.vips[index]
		if _, ok := vipsMap[vip.prefix]; ok {
			c.vips = append(c.vips[:index], c.vips[index+1:]...)
			index--
			err := vip.Delete()
			if err != nil {
				logrus.Errorf("Client: Error deleting vip: %v", err) // todo: err handling
			}
		}
	}
}

func (c *Conduit) getGateways() []string {
	if c.connection != nil {
		return c.connection.GetContext().GetIpContext().GetDstIpAddrs()
	}
	return []string{}
}
