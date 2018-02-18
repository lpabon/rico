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
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/libopenstorage/rico/pkg/cloudprovider"
	"github.com/libopenstorage/rico/pkg/config"
	"github.com/libopenstorage/rico/pkg/inframanager"
	"github.com/libopenstorage/rico/pkg/storageprovider"
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
	fs := fake.New(&storageprovider.Topology{})
	class := config.Class{
		Name:               "gp2",
		WatermarkHigh:      75,
		WatermarkLow:       25,
		DiskSizeGb:         8,
		MaximumTotalSizeGb: 1024,
		MinimumTotalSizeGb: 32,
	}
	configuration := &config.Config{
		Classes: []config.Class{class},
	}
	im := inframanager.NewManager(configuration, fc, fs)
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("> ")
		input, _ := reader.ReadString('\n')
		input = strings.Replace(input, "\n", "", -1)
		text := strings.Split(string(input), " ")

		if len(text) < 1 {
			continue
		}

		cmd := strings.ToLower(text[0])
		switch {
		case cmd == "quit":
			fmt.Println("Exiting...")
			os.Exit(0)
		case cmd == "r" || cmd == "reconcile":
			if err := im.Reconcile(); err == nil {
				fmt.Println("OK")
			} else {
				fmt.Printf("ERROR: %v\n", err)
			}
		case cmd == "na" || cmd == "node-add":
			if len(text) < 2 {
				fmt.Println("node-add <id>")
				continue
			}
			fs.NodeAdd(&storageprovider.StorageNode{
				Name: text[1],
				Metadata: storageprovider.InstanceMetadata{
					ID: text[1],
				},
			})
		case cmd == "us" || cmd == "utilization-set":
			if len(text) < 3 {
				fmt.Println("us <class> <int>")
				continue
			}
			utilization, err := strconv.Atoi(text[2])
			if err != nil {
				fmt.Printf("ERROR: %v\n", err)
				continue
			}
			for _, class := range configuration.Classes {
				if class.Name == text[1] {
					fs.SetUtilization(&class, utilization)
					fmt.Println("OK")
					break
				}
			}
		case cmd == "t" || cmd == "topology":
			t, _ := fs.GetTopology()
			fmt.Println(t.String(im.Config()))
		case cmd == "c" || cmd == "classes":
			for _, class := range configuration.Classes {
				fmt.Printf("%v\n", class)
			}
		case cmd == "cd" || cmd == "class-delete":
			if len(text) < 2 {
				fmt.Printf("Missing class name: class-delte <name>\n")
				continue
			}

			// This should be part of config
			found := false
			index := 0
			for i, class := range configuration.Classes {
				if class.Name == text[1] {
					found = true
					index = i
					break
				}
			}
			if !found {
				fmt.Printf("Class %s not found\n", text[1])
				continue
			}
			configuration.Classes[index] = configuration.Classes[len(configuration.Classes)-1]
			configuration.Classes = configuration.Classes[:len(configuration.Classes)-1]
			im.SetConfig(configuration)
			fmt.Println("OK")
		case cmd == "ca" || cmd == "class-add":
			class := config.Class{}
			for _, param := range text[1:] {
				kv := strings.Split(strings.ToLower(param), "=")
				if len(kv) != 2 {
					fmt.Printf("Bad param: %s\n", param)
				}
				switch kv[0] {
				case "name":
					class.Name = kv[1]
				case "wh":
					i, err := strconv.Atoi(kv[1])
					if err != nil {
						fmt.Printf("ERROR: %v\n", err)
					}
					class.WatermarkHigh = i
				case "wl":
					i, err := strconv.Atoi(kv[1])
					if err != nil {
						fmt.Printf("ERROR: %v\n", err)
					}
					class.WatermarkLow = i
				case "size":
					i, err := strconv.ParseInt(kv[1], 10, 64)
					if err != nil {
						fmt.Printf("ERROR: %v\n", err)
					}
					class.DiskSizeGb = i
				case "max":
					i, err := strconv.ParseInt(kv[1], 10, 64)
					if err != nil {
						fmt.Printf("ERROR: %v\n", err)
					}
					class.MaximumTotalSizeGb = i
				case "min":
					i, err := strconv.ParseInt(kv[1], 10, 64)
					if err != nil {
						fmt.Printf("ERROR: %v\n", err)
					}
					class.MinimumTotalSizeGb = i
				default:
					fmt.Printf("Unknown key: %s\n", kv[0])
				}
			}
			if len(class.Name) == 0 {
				fmt.Println("Name missing: name=<name>")
				continue
			}
			if class.WatermarkHigh == 0 || class.WatermarkLow == 0 {
				fmt.Println("Watermarks missing: wh=<int> wl=<int>")
				continue
			}
			if class.MinimumTotalSizeGb == 0 || class.MaximumTotalSizeGb == 0 {
				fmt.Println("Max or min missing: max=<int> min=<int>")
				continue
			}
			if class.DiskSizeGb == 0 {
				fmt.Println("Size missing: size=<int>")
				continue
			}
			configuration.Classes = append(configuration.Classes, class)
			im.SetConfig(configuration)
			fmt.Println("OK")
		default:
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n", text[0])
		}
	}
}
