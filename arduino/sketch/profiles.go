// This file is part of arduino-cli.
//
// Copyright 2020-2022 ARDUINO SA (http://www.arduino.cc/)
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

package sketch

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/arduino/go-paths-helper"
	semver "go.bug.st/relaxed-semver"
	"gopkg.in/yaml.v2"
)

// Project represents all the
type Project struct {
	Profiles       map[string]*Profile `yaml:"profiles"`
	DefaultProfile string              `yaml:"default_profile"`
}

// AsYaml outputs the project file as Yaml
func (p *Project) AsYaml() string {
	res := "profiles:\n"
	for name, profile := range p.Profiles {
		res += fmt.Sprintf("  %s:\n", name)
		res += profile.AsYaml()
		res += "\n"
	}
	if p.DefaultProfile != "" {
		res += fmt.Sprintf("default_profile: %s\n", p.DefaultProfile)
	}
	return res
}

// Profile is a sketch profile, it contains a reference to all the resources
// needed to build and upload a sketch
type Profile struct {
	Notes     string                   `yaml:"notes"`
	FQBN      string                   `yaml:"fqbn"`
	Platforms ProfileRequiredPlatforms `yaml:"platforms"`
	Libraries ProfileRequiredLibraries `yaml:"libraries"`
}

// AsYaml outputs the profile as Yaml
func (p *Profile) AsYaml() string {
	res := ""
	if p.Notes != "" {
		res += fmt.Sprintf("    notes: %s\n", p.Notes)
	}
	res += fmt.Sprintf("    fqbn: %s\n", p.FQBN)
	res += p.Platforms.AsYaml()
	res += p.Libraries.AsYaml()
	return res
}

// ProfileRequiredPlatforms is a list of ProfilePlatformReference (platforms
// required to build the sketch using this profile)
type ProfileRequiredPlatforms []*ProfilePlatformReference

// AsYaml outputs the required platforms as Yaml
func (p *ProfileRequiredPlatforms) AsYaml() string {
	res := ""
	for _, platform := range *p {
		res += platform.AsYaml()
	}
	return res
}

// ProfileRequiredLibraries is a list of ProfileLibraryReference (libraries
// required to build the sketch using this profile)
type ProfileRequiredLibraries []*ProfileLibraryReference

// AsYaml outputs the required libraries as Yaml
func (p *ProfileRequiredLibraries) AsYaml() string {
	return ""
}

// ProfilePlatformReference is a reference to a platform
type ProfilePlatformReference struct {
	Packager         string
	Architecture     string
	Version          *semver.Version
	PlatformIndexURL *url.URL
}

func (p *ProfilePlatformReference) String() string {
	res := fmt.Sprintf("%s:%s@%s", p.Packager, p.Architecture, p.Version)
	if p.PlatformIndexURL != nil {
		res += fmt.Sprintf(" (%s)", p.PlatformIndexURL)
	}
	return res
}

// AsYaml outputs the platform reference as Yaml
func (p *ProfilePlatformReference) AsYaml() string {
	res := fmt.Sprintf("      - platform: %s:%s (%s)\n", p.Packager, p.Architecture, p.Version)
	if p.PlatformIndexURL != nil {
		res += fmt.Sprintf("        platform_index_url: %s\n", p.PlatformIndexURL)
	}
	return res
}

func parseNameAndVersion(in string) (string, string, bool) {
	re := regexp.MustCompile(`^([a-zA-Z0-9.\-_ :]+) \((.+)\)$`)
	split := re.FindAllStringSubmatch(in, -1)
	if len(split) != 1 || len(split[0]) != 3 {
		return "", "", false
	}
	return split[0][1], split[0][2], true
}

func (p *ProfilePlatformReference) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var data map[string]string
	if err := unmarshal(&data); err != nil {
		return err
	}
	if platformId, ok := data["platform"]; !ok {
		return fmt.Errorf(tr("missing 'platform' directive"))
	} else if platformId, platformVersion, ok := parseNameAndVersion(platformId); !ok {
		return fmt.Errorf(tr("invalid 'platform' directive"))
	} else if c, err := semver.Parse(platformVersion); err != nil {
		return fmt.Errorf("%s: %w", tr("error parsing version constraints"), err)
	} else if split := strings.SplitN(platformId, ":", 2); len(split) != 2 {
		return fmt.Errorf("%s: %s", tr("invalid platform identifier"), platformId)
	} else {
		p.Packager = split[0]
		p.Architecture = split[1]
		p.Version = c
	}

	if rawIndexURL, ok := data["platform_index_url"]; ok {
		if indexURL, err := url.Parse(rawIndexURL); err != nil {
			return fmt.Errorf("%s: %w", tr("invlid platform index URL:"), err)
		} else {
			p.PlatformIndexURL = indexURL
		}
	}
	return nil
}

// ProfileLibraryReference is a reference to a library
type ProfileLibraryReference struct {
	Library string
	Version *semver.Version
}

func (l *ProfileLibraryReference) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var data string
	if err := unmarshal(&data); err != nil {
		return err
	}
	if libName, libVersion, ok := parseNameAndVersion(data); !ok {
		return fmt.Errorf("%s %s", tr("invalid library directive:"), data)
	} else if v, err := semver.Parse(libVersion); err != nil {
		return fmt.Errorf("%s %w", tr("invalid version:"), err)
	} else {
		l.Library = libName
		l.Version = v
	}
	return nil
}

func (l *ProfileLibraryReference) String() string {
	return fmt.Sprintf("%s@%s", l.Library, l.Version)
}

// LoadProjectFile reads a sketch project file
func LoadProjectFile(file *paths.Path) (*Project, error) {
	data, err := file.ReadFile()
	if err != nil {
		return nil, err
	}
	res := &Project{}
	if err := yaml.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	return res, nil
}