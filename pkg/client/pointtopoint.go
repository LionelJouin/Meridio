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
	"sync"
	"time"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/sirupsen/logrus"
)

// PointToPoint manages the requests and closes of a point to point connection
type PointToPoint struct {
	// NetworkServiceClient
	NetworkServiceClient networkservice.NetworkServiceClient
	// MaxRetry represents the maximum number of attempt before returning an error
	MaxRetry int
	// Timeout represents maximum time in millisecond for each attempt
	Timeout time.Duration
	// Delay represents the time in millisecond between each attempt
	Delay time.Duration
	// ConnectionChannel returns an event when connection is established or closed
	ConnectionChannel chan<- *ConnectionEvent
	// context
	context context.Context
	// cancel function to cancel the request
	cancel context.CancelFunc
	// connection managed
	connection *networkservice.Connection
	// mutex to sync request and close calls
	mu sync.Mutex
}

// Request checks if the connection is already established, checks the validity of the request
// and request the connection to NSM. If the request fails, multiple attempts are made based on
// the MaxRetry variables. Once established, a connection event will be send via the connection channel.
func (ptp *PointToPoint) Request(ctx context.Context, request *networkservice.NetworkServiceRequest) error {
	if ptp.connection != nil {
		return fmt.Errorf("this connection is already established")
	}
	if !ptp.requestIsValid(request) {
		return fmt.Errorf("request is not valid: %v", request)
	}
	ptp.mu.Lock()
	defer ptp.mu.Unlock()
	ptp.context, ptp.cancel = context.WithCancel(ctx)
	defer ptp.cancel()
	attempt := 0
	for ; attempt < ptp.MaxRetry; attempt++ {
		if ptp.context.Err() == context.Canceled {
			return fmt.Errorf("Request canceled: %v", ptp.context.Err())
		}
		connection, err := ptp.requestAttempt(ptp.context, request)
		if err != nil {
			logrus.Errorf("Point to Point: Request err: %v", err)
			time.Sleep(ptp.Delay * time.Millisecond)
			continue
		}
		ptp.connection = connection
		break
	}
	if ptp.connection == nil || attempt >= ptp.MaxRetry {
		return fmt.Errorf("requesting the connection failed after max retry")
	}
	if ptp.ConnectionChannel == nil {
		return nil
	}
	ptp.ConnectionChannel <- &ConnectionEvent{
		Connection: ptp.connection,
		Status:     Created,
	}
	return nil
}

// Close cancels the request call if it still running and, with multiple attempts if needed, closes
// the current connection via the NSM API. Once closed, a connection event will be send via the
// connection channel.
func (ptp *PointToPoint) Close(ctx context.Context) error {
	if ptp.connection == nil {
		return nil
	}
	if ptp.cancel != nil {
		ptp.cancel() // Cancel pending request
	}
	ptp.mu.Lock()
	defer ptp.mu.Unlock()
	ptp.context, ptp.cancel = context.WithCancel(ctx)
	defer ptp.cancel()
	attempt := 0
	for ; attempt < ptp.MaxRetry; attempt++ {
		err := ptp.closeAttempt(ptp.context)
		if err != nil {
			logrus.Errorf("Point to Point: Close err: %v", err)
			time.Sleep(ptp.Delay * time.Millisecond)
			continue
		}
		break
	}
	if attempt > ptp.MaxRetry {
		return fmt.Errorf("closing the connection failed after max retry")
	}
	if ptp.ConnectionChannel != nil {
		ptp.ConnectionChannel <- &ConnectionEvent{
			Connection: ptp.connection,
			Status:     Deleted,
		}
	}
	ptp.connection = nil
	return nil
}

func (ptp *PointToPoint) requestAttempt(ctx context.Context, request *networkservice.NetworkServiceRequest) (*networkservice.Connection, error) {
	contextWithTimeout, cancel := context.WithTimeout(ctx, ptp.Timeout*time.Millisecond)
	defer cancel()
	connection, err := ptp.NetworkServiceClient.Request(contextWithTimeout, request)
	if err != nil {
		return nil, err
	}
	return connection, nil
}

func (ptp *PointToPoint) closeAttempt(ctx context.Context) error {
	contextWithTimeout, cancel := context.WithTimeout(ctx, ptp.Timeout*time.Millisecond)
	defer cancel()
	_, err := ptp.NetworkServiceClient.Close(contextWithTimeout, ptp.connection)
	return err
}

func (ptp *PointToPoint) requestIsValid(request *networkservice.NetworkServiceRequest) bool {
	if request == nil {
		return false
	}
	if request.GetMechanismPreferences() == nil || len(request.GetMechanismPreferences()) == 0 {
		return false
	}
	if request.GetConnection() == nil || request.GetConnection().NetworkService == "" {
		return false
	}
	return true
}

// NewPointToPoint creates and initializes a new PointToPoint
func NewPointToPoint(networkServiceClient networkservice.NetworkServiceClient,
	maxRetry int,
	timeout time.Duration,
	delay time.Duration,
	connectionChannel chan<- *ConnectionEvent) *PointToPoint {
	pointToPoint := &PointToPoint{
		NetworkServiceClient: networkServiceClient,
		MaxRetry:             maxRetry,
		Timeout:              timeout,
		Delay:                delay,
		ConnectionChannel:    connectionChannel,
	}
	return pointToPoint
}
