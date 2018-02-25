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

	"github.com/abiosoft/ishell"
	"github.com/pborman/uuid"
)

type FakeCloud struct{}

func (f *FakeCloud) SetConfig(config *config.Config) {}
func (f *FakeCloud) DeviceCreate(
	instanceID string,
	class *config.Class,
) (*cloudprovider.Device, error) {
	return &cloudprovider.Device{
		ID:   uuid.New(),
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

	// ishell
	shell := ishell.New()
	shell.Println("Rico Simulator")
	shell.SetPrompt("> ")

	// Node add
	shell.AddCmd(&ishell.Cmd{
		Name:    "node-add",
		Aliases: []string{"na"},
		Func: func(c *ishell.Context) {
			if len(c.Args) < 1 {
				c.Err(fmt.Errorf("node-add <id>"))
				return
			}
			fs.NodeAdd(&storageprovider.StorageNode{
				Name: c.Args[0],
				Metadata: storageprovider.InstanceMetadata{
					ID: c.Args[0],
				},
			})
		},
		Help: "Add a node to the storage system",
	})

	// Utilization set
	shell.AddCmd(&ishell.Cmd{
		Name:    "utilization-set",
		Aliases: []string{"us"},
		Func: func(c *ishell.Context) {
			if len(c.Args) < 2 {
				c.Err(fmt.Errorf("utilization-set <class-name> <int>"))
				return
			}
			utilization, err := strconv.Atoi(c.Args[1])
			if err != nil {
				c.Err(err)
				return
			}

			found := false
			for _, class := range configuration.Classes {
				if class.Name == c.Args[0] {
					found = true
					fs.SetUtilization(&class, utilization)
					break
				}
			}
			if found {
				c.Println("OK")
			} else {
				c.Err(fmt.Errorf("class %s not found", c.Args[0]))
			}

		},
		Help: "Set utilization of a class across the cluster",
	})

	// Show topology
	shell.AddCmd(&ishell.Cmd{
		Name:    "topology",
		Aliases: []string{"t"},
		Func: func(c *ishell.Context) {
			t, _ := fs.GetTopology()
			c.Println(t.String(im.Config()))
		},
		Help: "Show storage topoology",
	})

	// Reconcile
	shell.AddCmd(&ishell.Cmd{
		Name:    "reconcile",
		Aliases: []string{"r"},
		Func: func(c *ishell.Context) {
			if err := im.Reconcile(); err == nil {
				c.Println("OK")
			} else {
				c.Err(err)
			}
		},
		Help: "Reconcile once",
	})

	// List classes
	shell.AddCmd(&ishell.Cmd{
		Name:    "class-list",
		Aliases: []string{"c", "classes"},
		Func: func(c *ishell.Context) {
			for _, class := range configuration.Classes {
				c.Printf("%v\n", class)
			}
		},
		Help: "List classes",
	})

	// Delete class
	shell.AddCmd(&ishell.Cmd{
		Name:    "class-delete",
		Aliases: []string{"cd"},
		Func: func(c *ishell.Context) {
			if len(c.Args) < 1 {
				c.Err(fmt.Errorf("Missing class name: class-delte <name>"))
				return
			}
			className := c.Args[0]

			// This should be part of config
			found := false
			index := 0
			for i, class := range configuration.Classes {
				if class.Name == className {
					found = true
					index = i
					break
				}
			}
			if !found {
				c.Err(fmt.Errorf("Class %s not found\n", className))
				return
			}

			// Delete class
			configuration.Classes[index] = configuration.Classes[len(configuration.Classes)-1]
			configuration.Classes = configuration.Classes[:len(configuration.Classes)-1]
			im.SetConfig(configuration)
			c.Println("OK")
		},
		Help: "Delete class",
	})

	// Add class
	shell.AddCmd(&ishell.Cmd{
		Name:    "class-add",
		Aliases: []string{"ca"},
		Func: func(c *ishell.Context) {
			if len(c.Args) < 6 {
				c.Err(fmt.Errorf("Missing arguments: " +
					"ca name=<name> " +
					"wh=<watermark high> " +
					"wl=<watermark low> " +
					"size=<disk size Gi> " +
					"max=<total max size Gi> " +
					"min=<total min size Gi>"))
				return
			}
			newClass := config.Class{}
			for _, param := range c.Args {
				kv := strings.Split(strings.ToLower(param), "=")
				if len(kv) != 2 {
					c.Err(fmt.Errorf("Bad param: %s\n", param))
					return
				}
				switch kv[0] {
				case "name":
					newClass.Name = kv[1]
				case "wh":
					i, err := strconv.Atoi(kv[1])
					if err != nil {
						c.Err(err)
						return
					}
					newClass.WatermarkHigh = i
				case "wl":
					i, err := strconv.Atoi(kv[1])
					if err != nil {
						c.Err(err)
						return
					}
					newClass.WatermarkLow = i
				case "size":
					i, err := strconv.ParseInt(kv[1], 10, 64)
					if err != nil {
						c.Err(err)
						return
					}
					newClass.DiskSizeGb = i
				case "max":
					i, err := strconv.ParseInt(kv[1], 10, 64)
					if err != nil {
						c.Err(err)
						return
					}
					newClass.MaximumTotalSizeGb = i
				case "min":
					i, err := strconv.ParseInt(kv[1], 10, 64)
					if err != nil {
						c.Err(err)
						return
					}
					newClass.MinimumTotalSizeGb = i
				default:
					c.Err(fmt.Errorf("Unknown key: %s", kv[0]))
					return
				}
			}
			if len(newClass.Name) == 0 {
				c.Err(fmt.Errorf("Name missing: name=<name>"))
				return
			}
			if newClass.WatermarkHigh == 0 || newClass.WatermarkLow == 0 {
				c.Err(fmt.Errorf("Watermarks missing: wh=<int> wl=<int>"))
				return
			}
			if newClass.MinimumTotalSizeGb == 0 || newClass.MaximumTotalSizeGb == 0 {
				c.Err(fmt.Errorf("Max or min missing: max=<int> min=<int>"))
				return
			}
			if newClass.DiskSizeGb == 0 {
				c.Err(fmt.Errorf("Size missing: size=<int>"))
				return
			}
			configuration.Classes = append(configuration.Classes, newClass)
			im.SetConfig(configuration)
			c.Println("OK")
		},
		Help: "Add class",
	})

	// Run shell
	shell.Run()
	defer shell.Close()
	if false {

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
}
