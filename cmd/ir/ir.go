/*
Package config provides the configuration to the Manager
Copyright 2018 Portworx

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
	"github.com/libopenstorage/rico/pkg/config"
	"github.com/libopenstorage/rico/pkg/inframanager"
	"github.com/libopenstorage/rico/pkg/storageprovider/fake"
)

type FakeCloud struct{}

func (f *FakeCloud) SetConfig(config *config.Config) {}
func (f *FakeCloud) DeviceCreate(
	instanceID string,
	class *config.Class,
) (*cloudprovider.Device, error) {
	return &cloudprovider.Device{
		ID:   "nothing",
		Path: "nothing",
		Size: class.DiskSizeGb,
	}, nil
}
func (f *FakeCloud) DeviceDelete(instanceID string, deviceID string) error {
	return nil
}

func main() {
	fc := &FakeCloud{}
	fs := fake.New()
}
