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

package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/arduino/arduino-cli/commands/cmderrors"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

// SettingsGetValue returns a settings value given its key. If the key is not present
// an error will be returned, so that we distinguish empty settings from missing
// ones.
func (s *arduinoCoreServerImpl) ConfigurationGet(ctx context.Context, req *rpc.ConfigurationGetRequest) (*rpc.ConfigurationGetResponse, error) {
	// TODO
	return &rpc.ConfigurationGetResponse{}, nil
}

func (s *arduinoCoreServerImpl) SettingsSetValue(ctx context.Context, req *rpc.SettingsSetValueRequest) (*rpc.SettingsSetValueResponse, error) {
	// Determine the existence and the kind of the value
	key := req.GetKey()
	defaultValue, ok := s.settings.Defaults.GetOk(key)
	if !ok {
		return nil, &cmderrors.InvalidArgumentError{Message: fmt.Sprintf("key %s not found", key)}
	}
	expectedType := reflect.TypeOf(defaultValue)

	// Extract the value from the request
	var newValue any
	if err := json.Unmarshal([]byte(req.GetValueJson()), &newValue); err != nil {
		return nil, &cmderrors.InvalidArgumentError{Message: fmt.Sprintf("invalid value: %v", err)}
	}
	newValueType := reflect.TypeOf(newValue)

	// Check if the value is of the same type of the default value
	if newValueType != expectedType {
		return nil, &cmderrors.InvalidArgumentError{Message: fmt.Sprintf("value type mismatch: expected %T, got %T", defaultValue, newValue)}
	}

	// Set the value
	s.settings.Set(key, newValue)
	return &rpc.SettingsSetValueResponse{}, nil
}

func (s *arduinoCoreServerImpl) SettingsGetValue(ctx context.Context, req *rpc.SettingsGetValueRequest) (*rpc.SettingsGetValueResponse, error) {
	key := req.GetKey()
	value, ok := s.settings.GetOk(key)
	if !ok {
		value, ok = s.settings.Defaults.GetOk(key)
	}
	if !ok {
		return nil, &cmderrors.InvalidArgumentError{Message: fmt.Sprintf("key %s not found", key)}
	}
	valueJson, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("error marshalling value: %v", err)
	}
	return &rpc.SettingsGetValueResponse{
		ValueJson: string(valueJson),
	}, nil
}

// ConfigurationSave encodes the current configuration in the specified format
func (s *arduinoCoreServerImpl) ConfigurationSave(ctx context.Context, req *rpc.ConfigurationSaveRequest) (*rpc.ConfigurationSaveResponse, error) {
	// TODO
	return &rpc.ConfigurationSaveResponse{}, nil
}

// SettingsReadFromFile read settings from a YAML file and replace the settings currently stored in memory.
func (s *arduinoCoreServerImpl) ConfigurationOpen(ctx context.Context, req *rpc.ConfigurationOpenRequest) (*rpc.ConfigurationOpenResponse, error) {
	// TODO
	return &rpc.ConfigurationOpenResponse{}, nil
}

// SettingsEnumerate returns the list of all the settings keys.
func (s *arduinoCoreServerImpl) SettingsEnumerate(ctx context.Context, req *rpc.SettingsEnumerateRequest) (*rpc.SettingsEnumerateResponse, error) {
	var entries []*rpc.SettingsEnumerateResponse_Entry
	for _, k := range s.settings.Defaults.AllKeys() {
		v := s.settings.Defaults.Get(k)
		entries = append(entries, &rpc.SettingsEnumerateResponse_Entry{
			Key:  k,
			Type: reflect.TypeOf(v).String(),
		})
	}
	return &rpc.SettingsEnumerateResponse{
		Entries: entries,
	}, nil
}
