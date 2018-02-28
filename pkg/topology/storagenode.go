/*
Package topology defines how to get information from the infrastructure
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
package topology

import (
	"fmt"

	"github.com/libopenstorage/rico/pkg/config"
)

func (n *StorageNode) Utilization(class *config.Class) int {
	sum, num := n.RawUtilization(class)
	if num == 0 {
		return 0
	}
	return int(sum / num)
}

func (n *StorageNode) RawUtilization(class *config.Class) (int, int) {
	sum, num := 0, 0
	if len(n.Pools) != 0 {
		for _, pool := range n.Pools {
			if class.Name == pool.Class {
				sum += pool.Utilization
				num++
			}
		}
	} else {
		for _, device := range n.Devices {
			if class.Name == device.Class {
				sum += device.Utilization
				num++
			}
		}
	}
	return sum, num
}

func (n *StorageNode) TotalStorage(class *config.Class) int64 {
	total := int64(0)
	for _, d := range n.Devices {
		if d.Class == class.Name {
			total += d.Size
		}
	}
	return total
}

// Rename Function
func (n *StorageNode) NumDisks(class *config.Class) (int, *Pool) {
	var (
		numDisks int
		p        *Pool
		ok       bool
	)

	numDisks = 1
	if p, ok = n.Pools[class.Name]; ok {
		numDisks = p.SetSize
	}

	return numDisks, p
}

func (n *StorageNode) Verify() error {
	if len(n.Metadata.ID) == 0 {
		return fmt.Errorf("Node missing instance metadata id")
	}
	for _, pool := range n.Pools {
		if err := pool.Verify(); err != nil {
			return err
		}
	}
	for _, device := range n.Devices {
		if err := device.Verify(); err != nil {
			return err
		}
	}

	return nil
}

// TODO: DeviceInPool
func (n *StorageNode) DevicesOnPool(p *Pool) []*Device {
	devices := make([]*Device, 0)
	if p != nil {
		for _, device := range n.Devices {
			if device.Pool == p.Name {
				devices = append(devices, device)
			}
		}
	}

	return devices
}

func (n *StorageNode) String() string {
	s := fmt.Sprintf("N[%s|%d]: ",
		n.Metadata.ID,
		len(n.Devices))
	for _, device := range n.Devices {
		s += device.String()
	}
	return s + "\n"
}

func (n *StorageNode) DevicesForClass(class *config.Class) []*Device {
	devices := make([]*Device, 0)
	for _, device := range n.Devices {
		if device.Class == class.Name {
			devices = append(devices, device)
		}
	}
	return devices
}
