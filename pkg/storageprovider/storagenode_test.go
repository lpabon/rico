/*
Package storageprovider provides an interface to storage providers
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

package storageprovider

import (
	"fmt"
	"testing"

	"github.com/libopenstorage/rico/pkg/config"
)

func TestStorageNodeDetermineStorageToRemove(t *testing.T) {
	topology := &Topology{
		Cluster: StorageCluster{
			StorageNodes: []*StorageNode{
				&StorageNode{
					Metadata: InstanceMetadata{
						ID: "one",
					},
					Devices: []*Device{
						&Device{
							Class:       "c1",
							Utilization: 30,
							Metadata: DeviceMetadata{
								ID: "d1",
							},
						},
						&Device{
							Class:       "c1",
							Utilization: 30,
							Metadata: DeviceMetadata{
								ID: "d2",
							},
						},
						&Device{
							Class:       "c2",
							Utilization: 30,
							Metadata: DeviceMetadata{
								ID: "d3",
							},
						},
						&Device{
							Class:       "c2",
							Utilization: 30,
							Metadata: DeviceMetadata{
								ID: "d4",
							},
						},
					},
				},
				&StorageNode{
					Metadata: InstanceMetadata{
						ID: "two",
					},
					Devices: []*Device{
						&Device{
							Class:       "c1",
							Utilization: 30,
							Metadata: DeviceMetadata{
								ID: "d1",
							},
						},
						&Device{
							Class:       "c1",
							Utilization: 30,
							Metadata: DeviceMetadata{
								ID: "d2",
							},
						},
						&Device{
							Class:       "c2",
							Utilization: 30,
							Metadata: DeviceMetadata{
								ID: "d3",
							},
						},
						&Device{
							Class:       "c2",
							Utilization: 30,
							Metadata: DeviceMetadata{
								ID: "d4",
							},
						},
					},
				},
			},
		},
	}

	n, p, d := topology.DetermineStorageToRemove(&config.Class{
		Name: "c2",
	})
	fmt.Printf("%s %s %s", n, p, d)

}
