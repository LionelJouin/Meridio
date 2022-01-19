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
	"io"

	nspAPI "github.com/nordix/meridio/api/nsp/v1"
	"github.com/sirupsen/logrus"
)

type Configuration struct {
	Watcher                    Watcher
	Trench                     *nspAPI.Trench
	ConfigurationManagerClient nspAPI.ConfigurationManagerClient
	cancel                     context.CancelFunc
}

type Watcher interface {
	SetVIPs([]string) error
}

func NewConfiguration(watcher Watcher,
	trench *nspAPI.Trench,
	configurationManagerClient nspAPI.ConfigurationManagerClient) *Configuration {
	c := &Configuration{
		Watcher:                    watcher,
		Trench:                     trench,
		ConfigurationManagerClient: configurationManagerClient,
	}
	return c
}

func (c *Configuration) WatchVIPs(ctx context.Context) {
	ctx, c.cancel = context.WithCancel(ctx)
	for { // Todo: retry
		if ctx.Err() != nil {
			return
		}
		vipsToWatch := &nspAPI.Vip{
			Trench: c.Trench,
		}
		watchVIPClient, err := c.ConfigurationManagerClient.WatchVip(ctx, vipsToWatch)
		if err != nil {
			logrus.Warnf("err watchVIPClient.Recv: %v", err) // todo
			continue
		}
		for {
			vipResponse, err := watchVIPClient.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				logrus.Warnf("err watchVIPClient.Recv: %v", err) // todo
				break
			}
			err = c.Watcher.SetVIPs(vipResponse.ToSlice())
			if err != nil {
				logrus.Warnf("err set vips: %v", err) // todo
			}
		}
	}
}

func (c *Configuration) Delete() {
	if c.cancel != nil {
		c.cancel()
	}
}
