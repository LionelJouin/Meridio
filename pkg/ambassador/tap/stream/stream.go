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

package stream

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	ambassadorAPI "github.com/nordix/meridio/api/ambassador/v1"
	nspAPI "github.com/nordix/meridio/api/nsp/v1"
	"github.com/nordix/meridio/pkg/ambassador/tap/types"
	lbTypes "github.com/nordix/meridio/pkg/loadbalancer/types"
	"github.com/sirupsen/logrus"
)

// Stream implements types.Stream
type Stream struct {
	Stream               *nspAPI.Stream
	TargetRegistryClient nspAPI.TargetRegistryClient
	StreamRegistry       types.Registry
	Conduit              Conduit
	MaxNumberOfTargets   int // Maximum number of targets registered in this stream
	targetStatus         nspAPI.Target_Status
	identifier           int
	ips                  []string
	cancelOpen           context.CancelFunc
	mu                   sync.Mutex
}

// New is the constructor of Stream.
// The constructor will add the new created stream to the conduit.
func New(stream *nspAPI.Stream,
	targetRegistryClient nspAPI.TargetRegistryClient,
	streamRegistry types.Registry,
	maxNumberOfTargets int,
	conduit Conduit) (types.Stream, error) {
	// todo: check if stream valid
	// todo: check if target registry client valid
	s := &Stream{
		Stream:               stream,
		TargetRegistryClient: targetRegistryClient,
		StreamRegistry:       streamRegistry,
		Conduit:              conduit,
		MaxNumberOfTargets:   maxNumberOfTargets,
		identifier:           -1,
		targetStatus:         nspAPI.Target_DISABLED,
	}
	s.StreamRegistry.Add(context.TODO(), s.Stream)
	s.setPendingStatus(PendingTime)
	return s, nil
}

// todo: docs
// Open the stream in the conduit by generating a identifier and registering
// the target to the NSP service while avoiding the identifier collisions.
// If success, no error will be returned and an event will be send via the streamWatcher.
// If not, an error will be returned.
func (s *Stream) Open(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	ctx, s.cancelOpen = context.WithCancel(ctx)
	s.ips = s.Conduit.GetIPs()
	go func() {
		for { // Todo: retry
			if ctx.Err() != nil {
				return
			}
			err := s.open(ctx)
			if err != nil {
				logrus.Infof("stream fail %v: %v", s.Stream, err)
				continue
			}
			break
		}
		logrus.Infof("stream opened %v with target: %v", s.Stream, s.getTarget())
	}()
	return nil
}

// todo: docs
// Close the stream in the conduit by unregistering target from the NSP service.
// If success, no error will be returned and an event will be send via the streamWatcher.
// If not, an error will be returned.
func (s *Stream) Close(ctx context.Context) error {
	if s.cancelOpen != nil {
		s.cancelOpen()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.TargetRegistryClient.Unregister(ctx, s.getTarget())
	s.targetStatus = nspAPI.Target_DISABLED
	s.setStatus(ambassadorAPI.StreamStatus_UNAVAILABLE)
	if err != nil {
		return err
	}
	return nil
}

// todo: docs
func (s *Stream) Equals(stream *nspAPI.Stream) bool {
	return s.Stream.Equals(stream)
}

func (s *Stream) setStatus(status ambassadorAPI.StreamStatus_Status) {
	s.StreamRegistry.SetStatus(s.Stream, status)
}

func (s *Stream) setIdentifier(exclusionList []string) {
	exclusionListMap := make(map[string]struct{})
	for _, item := range exclusionList {
		exclusionListMap[item] = struct{}{}
	}
	for !s.isIdentifierValid(exclusionListMap, 1, s.MaxNumberOfTargets) {
		rand.Seed(time.Now().UnixNano())
		s.identifier = rand.Intn(s.MaxNumberOfTargets) + 1
	}
}

func (s *Stream) isIdentifierValid(exclusionList map[string]struct{}, min int, max int) bool {
	_, exists := exclusionList[strconv.Itoa(s.identifier)]
	return !exists && s.identifier >= min && s.identifier <= max
}

func (s *Stream) checkIdentifierCollision(identifiersInUse []string) bool {
	found := 0
	for _, identifier := range identifiersInUse {
		if identifier == strconv.Itoa(s.identifier) {
			found++
		}
	}
	return found > 1
}

func (s *Stream) getIdentifiersInUse(ctx context.Context) ([]string, error) {
	identifiers := []string{}
	context, cancel := context.WithCancel(ctx)
	defer cancel()
	watchClient, err := s.TargetRegistryClient.Watch(context, &nspAPI.Target{
		Status: nspAPI.Target_ANY,
		Type:   nspAPI.Target_DEFAULT,
		Stream: &nspAPI.Stream{
			Conduit: s.Stream.GetConduit(),
		},
	})
	if err != nil {
		return identifiers, err
	}
	responseTargets, err := watchClient.Recv()
	if err != nil {
		return identifiers, err
	}
	for _, target := range responseTargets.Targets {
		identifiers = append(identifiers, target.Context[lbTypes.IdentifierKey])
	}
	return identifiers, nil
}

func (s *Stream) getTarget() *nspAPI.Target {
	return &nspAPI.Target{
		Ips: s.ips,
		Context: map[string]string{
			lbTypes.IdentifierKey: strconv.Itoa(s.identifier),
		},
		Status: s.targetStatus,
		Stream: s.Stream,
	}
}

func (s *Stream) open(ctx context.Context) error {
	identifiersInUse, err := s.getIdentifiersInUse(ctx)
	if err != nil {
		return err
	}
	if len(identifiersInUse) >= s.MaxNumberOfTargets {
		return errors.New("no identifier available to register the target")
	}
	if s.targetStatus != nspAPI.Target_DISABLED {
		return nil
	}
	s.setIdentifier(identifiersInUse)
	_, err = s.TargetRegistryClient.Register(ctx, s.getTarget()) // register the target as disabled status
	if err != nil {
		return err
	}
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		identifiersInUse, err := s.getIdentifiersInUse(ctx)
		if err != nil {
			return err
		}
		if len(identifiersInUse) > s.MaxNumberOfTargets {
			err = errors.New("no identifier available to register the target")
			_, errUnregister := s.TargetRegistryClient.Unregister(ctx, s.getTarget())
			if errUnregister != nil {
				return fmt.Errorf("%w ; %v", errUnregister, err)
			}
			return err
		}
		// Checks if there is any collision since the last registration/update
		// of the target.
		collision := s.checkIdentifierCollision(identifiersInUse)
		if !collision {
			break
		}
		s.setIdentifier(identifiersInUse)
		_, err = s.TargetRegistryClient.Update(ctx, s.getTarget()) // Update the target identifier
		if err != nil {
			return err
		}
	}
	s.targetStatus = nspAPI.Target_ENABLED
	_, err = s.TargetRegistryClient.Update(ctx, s.getTarget()) // Update the target as enabled status
	if err != nil {
		return err
	}
	s.setStatus(ambassadorAPI.StreamStatus_OPEN)
	return nil
}

func (s *Stream) setPendingStatus(pendingTime time.Duration) {
	s.setStatus(ambassadorAPI.StreamStatus_PENDING)
	go func() {
		<-time.After(pendingTime)
		s.mu.Lock()
		defer s.mu.Unlock()
		if s.targetStatus == nspAPI.Target_DISABLED {
			s.setStatus(ambassadorAPI.StreamStatus_UNAVAILABLE)
		}
	}()
}
