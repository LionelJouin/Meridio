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

package types

import (
	"context"

	ambassadorAPI "github.com/nordix/meridio/api/ambassador/v1"
	nspAPI "github.com/nordix/meridio/api/nsp/v1"
)

type Registry interface {
	Add(context.Context, *nspAPI.Stream) error
	Remove(context.Context, *nspAPI.Stream) error
	SetStatus(*nspAPI.Stream, ambassadorAPI.StreamStatus_Status)
	Watch(context.Context, *nspAPI.Stream) (Watcher, error)
}

type Watcher interface {
	Stop()
	ResultChan() <-chan []*ambassadorAPI.StreamStatus
}
