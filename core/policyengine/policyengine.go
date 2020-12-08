//
// Copyright (C) 2020 IBM Corporation.
//
// Authors:
// Frederico Araujo <frederico.araujo@ibm.com>
// Teryl Taylor <terylt@ibm.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
package policyengine

import (
	"errors"
	"sync"

	"github.com/sysflow-telemetry/sf-apis/go/ioutils"
	"github.com/sysflow-telemetry/sf-apis/go/logger"
	"github.com/sysflow-telemetry/sf-apis/go/plugins"
	"github.com/sysflow-telemetry/sf-processor/core/cache"
	"github.com/sysflow-telemetry/sf-processor/core/flattener"
	"github.com/sysflow-telemetry/sf-processor/core/policyengine/engine"
)

const (
	pluginName  string = "policyengine"
	channelName string = "eventchan"
)

// PolicyEngine defines a driver for the Policy Engine plugin.
type PolicyEngine struct {
	pi         engine.PolicyInterpreter
	tables     *cache.SFTables
	outCh      chan *engine.Record
	filterOnly bool
	bypass     bool
	config     engine.Config
}

// NewPolicyEngine constructs a new Policy Engine plugin.
func NewPolicyEngine() plugins.SFProcessor {
	return new(PolicyEngine)
}

// GetName returns the plugin name.
func (s *PolicyEngine) GetName() string {
	return pluginName
}

// NewEventChan creates a new event record channel instance.
func NewEventChan(size int) interface{} {
	return &engine.RecordChannel{In: make(chan *engine.Record, size)}
}

// Register registers plugin to plugin cache.
func (s *PolicyEngine) Register(pc plugins.SFPluginCache) {
	pc.AddProcessor(pluginName, NewPolicyEngine)
	pc.AddChannel(channelName, NewEventChan)
}

// Init initializes the plugin.
func (s *PolicyEngine) Init(conf map[string]string) error {
	config, err := engine.CreateConfig(conf)
	if err != nil {
		return err
	}
	s.config = config
	s.pi = engine.NewPolicyInterpreter(s.config)
	s.tables = cache.GetInstance()
	if s.config.Mode == engine.FilterMode {
		logger.Trace.Println("Setting policy engine in filter mode")
		s.filterOnly = true
	} else if s.config.Mode == engine.BypassMode {
		logger.Trace.Println("Setting policy engine in bypass mode")
		s.bypass = true
		return nil
	}
	logger.Trace.Println("Loading policies from: ", config.PoliciesPath)
	paths, err := ioutils.ListFilePaths(config.PoliciesPath, ".yaml")
	if err == nil {
		if len(paths) == 0 {
			return errors.New("No policy files with extension .yaml found in path: " + config.PoliciesPath)
		}
		return s.pi.Compile(paths...)
	}
	return errors.New("Error while listing policies: " + err.Error())
}

// Process implements the main loop of the plugin.
func (s *PolicyEngine) Process(ch interface{}, wg *sync.WaitGroup) {
	in := ch.(*flattener.FlatChannel).In
	defer wg.Done()
	logger.Trace.Println("Starting policy engine with capacity: ", cap(in))
	out := func(r *engine.Record) { s.outCh <- r }
	for {
		if fc, ok := <-in; ok {
			if s.bypass {
				out(engine.NewRecord(*fc, s.tables))
			} else {
				s.pi.ProcessAsync(true, s.filterOnly, engine.NewRecord(*fc, s.tables), out)
			}
		} else {
			logger.Trace.Println("Input channel closed. Shutting down.")
			break
		}
	}
}

// SetOutChan sets the output channel of the plugin.
func (s *PolicyEngine) SetOutChan(ch interface{}) {
	s.outCh = (ch.(*engine.RecordChannel)).In
}

// Cleanup clean up the plugin resources.
func (s *PolicyEngine) Cleanup() {
	logger.Trace.Println("Exiting ", pluginName)
	if s.outCh != nil {
		close(s.outCh)
	}
}
