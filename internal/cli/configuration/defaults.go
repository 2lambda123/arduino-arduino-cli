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

package configuration

import (
	"os"
	"time"
)

// SetDefaults sets the default values for certain keys
func SetDefaults(settings *Settings) {
	// logging
	settings.Defaults.SetString("logging.level", "info")
	settings.Defaults.SetString("logging.format", "text")

	// Libraries
	settings.Defaults.SetBool("library.enable_unsafe_install", false)

	// Boards Manager
	settings.Defaults.Set("board_manager.additional_urls", []string{})

	// arduino directories
	settings.Defaults.SetString("directories.data", getDefaultArduinoDataDir())
	settings.Defaults.SetString("directories.downloads", "")
	settings.Defaults.SetString("directories.user", getDefaultUserDir())

	// Sketch compilation
	settings.Defaults.SetBool("sketch.always_export_binaries", false)
	settings.Defaults.SetUint("build_cache.ttl", uint(time.Hour*24*30))
	settings.Defaults.SetUint("build_cache.compilations_before_purge", 10)

	// daemon settings
	settings.Defaults.SetString("daemon.port", "50051")

	// metrics settings
	settings.Defaults.SetBool("metrics.enabled", true)
	settings.Defaults.SetString("metrics.addr", ":9090")

	// output settings
	settings.Defaults.SetBool("output.no_color", false)

	// updater settings
	settings.Defaults.SetBool("updater.enable_notification", true)

	// Bind env vars
	settings.Defaults.InjectEnvVars(os.Environ(), "ARDUINO")

	// Bind env aliases to keep backward compatibility
	// settings.defaults.BindEnv("library.enable_unsafe_install", "ARDUINO_ENABLE_UNSAFE_LIBRARY_INSTALL")
	// settings.defaults.BindEnv("directories.User", "ARDUINO_SKETCHBOOK_DIR")
	// settings.defaults.BindEnv("directories.Downloads", "ARDUINO_DOWNLOADS_DIR")
	// settings.defaults.BindEnv("directories.Data", "ARDUINO_DATA_DIR")
	// settings.defaults.BindEnv("sketch.always_export_binaries", "ARDUINO_SKETCH_ALWAYS_EXPORT_BINARIES")
}
