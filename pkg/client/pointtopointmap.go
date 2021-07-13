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

import "sync"

// pointToPointMap
type pointToPointMap struct {
	pointToPoints map[string]*PointToPoint
	mu            sync.Mutex
}

// add
func (ptpm *pointToPointMap) add(networkServiceEndpointName string, pointToPoint *PointToPoint) bool {
	ptpm.mu.Lock()
	defer ptpm.mu.Unlock()
	if ptpm.pointToPoints[networkServiceEndpointName] != nil {
		return false
	}
	ptpm.pointToPoints[networkServiceEndpointName] = pointToPoint
	return true
}

// delete
func (ptpm *pointToPointMap) delete(networkServiceEndpointName string) *PointToPoint {
	ptpm.mu.Lock()
	defer ptpm.mu.Unlock()
	pointToPoint := ptpm.pointToPoints[networkServiceEndpointName]
	if pointToPoint == nil {
		return nil
	}
	delete(ptpm.pointToPoints, networkServiceEndpointName)
	return pointToPoint
}

// list
func (ptpm *pointToPointMap) list() []string {
	ptpm.mu.Lock()
	defer ptpm.mu.Unlock()
	pointToPointList := []string{}
	for networkServiceEndpointName := range ptpm.pointToPoints {
		pointToPointList = append(pointToPointList, networkServiceEndpointName)
	}
	return pointToPointList
}

// newPointToPointMap creates and initializes a new pointToPointMap
func newPointToPointMap() *pointToPointMap {
	pointToPointMap := &pointToPointMap{
		pointToPoints: map[string]*PointToPoint{},
	}
	return pointToPointMap
}
