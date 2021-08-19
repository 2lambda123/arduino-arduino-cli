// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package arguments

import (
	"fmt"
	"net/url"
	"time"

	"github.com/arduino/arduino-cli/arduino/discovery"
	"github.com/arduino/arduino-cli/arduino/sketch"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Port contains the port arguments result.
// This is useful so all flags used by commands that need
// this information are consistent with each other.
type Port struct {
	address  string
	protocol string
	timeout  time.Duration
}

// AddToCommand adds the flags used to set port and protocol to the specified Command
func (p *Port) AddToCommand(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&p.address, "port", "p", "", tr("Upload port address, e.g.: COM3 or /dev/ttyACM2"))
	cmd.Flags().StringVarP(&p.protocol, "protocol", "l", "", tr("Upload port protocol, e.g: serial"))
	cmd.Flags().DurationVar(&p.timeout, "discovery-timeout", 5*time.Second, tr("Max time to wait for port discovery, e.g.: 30s, 1m"))
}

// GetPort returns the Port obtained by parsing command line arguments.
// The extra metadata for the ports is obtained using the pluggable discoveries.
func (p *Port) GetPort(instance *rpc.Instance, sk *sketch.Sketch) (*discovery.Port, error) {
	address := p.address
	protocol := p.protocol

	if address == "" && sk != nil && sk.Metadata != nil {
		deviceURI, err := url.Parse(sk.Metadata.CPU.Port)
		if err != nil {
			return nil, errors.Errorf("invalid Device URL format: %s", err)
		}
		if deviceURI.Scheme == "serial" {
			address = deviceURI.Host + deviceURI.Path
		}
	}
	if address == "" {
		// If no address is provided we assume the user is trying to upload
		// to a board that supports a tool that automatically detects
		// the attached board without specifying explictly a port.
		// Tools that work this way must be specified using the property
		// "BOARD_ID.upload.tool.default" in the platform's boards.txt.
		return &discovery.Port{
			Protocol: "default",
		}, nil
	}
	logrus.WithField("port", address).Tracef("Upload port")

	pm := commands.GetPackageManager(instance.Id)
	if pm == nil {
		return nil, errors.New("invalid instance")
	}
	dm := pm.DiscoveryManager()
	if errs := dm.RunAll(); len(errs) == len(dm.IDs()) {
		// All discoveries failed to run, we can't do anything
		return nil, fmt.Errorf("%v", errs)
	} else if len(errs) > 0 {
		// If only some discoveries failed to run just tell the user and go on
		for _, err := range errs {
			feedback.Error(err)
		}
	}
	eventChan, errs := dm.StartSyncAll()
	if len(errs) > 0 {
		return nil, fmt.Errorf("%v", errs)
	}

	defer func() {
		// Quit all discoveries at the end.
		if errs := dm.QuitAll(); len(errs) > 0 {
			logrus.Errorf("quitting discoveries when getting port metadata: %v", errs)
		}
	}()

	deadline := time.After(p.timeout)
	for {
		select {
		case portEvent := <-eventChan:
			if portEvent.Type != "add" {
				continue
			}
			port := portEvent.Port
			if (protocol == "" || protocol == port.Protocol) && address == port.Address {
				return port, nil
			}

		case <-deadline:
			// No matching port found
			if protocol == "" {
				return &discovery.Port{
					Address:  address,
					Protocol: "serial",
				}, nil
			}
			return nil, fmt.Errorf("port not found: %s %s", address, protocol)
		}
	}
}
