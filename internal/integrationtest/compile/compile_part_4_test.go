// This file is part of arduino-cli.
//
// Copyright 2022 ARDUINO SA (http://www.arduino.cc/)
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

package compile_test

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

func TestCompileWithLibrary(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	_, _, err = cli.Run("core", "install", "arduino:avr@1.8.3")
	require.NoError(t, err)

	sketchName := "CompileSketchWithWiFi101Dependency"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	fqbn := "arduino:avr:uno"
	// Create new sketch and add library include
	_, _, err = cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)
	sketchFile := sketchPath.Join(sketchName + ".ino")
	data, err := sketchFile.ReadFile()
	require.NoError(t, err)
	data = append([]byte("#include <WiFi101.h>\n"), data...)
	err = sketchFile.WriteFile(data)
	require.NoError(t, err)

	// Manually installs a library
	gitUrl := "https://github.com/arduino-libraries/WiFi101.git"
	libPath := cli.SketchbookDir().Join("my-libraries", "WiFi101")
	_, err = git.PlainClone(libPath.String(), false, &git.CloneOptions{
		URL:           gitUrl,
		ReferenceName: plumbing.NewTagReferenceName("0.16.1"),
	})
	require.NoError(t, err)

	stdout, _, err := cli.Run("compile", "-b", fqbn, sketchPath.String(), "--library", libPath.String(), "-v")
	require.NoError(t, err)
	require.Contains(t, string(stdout), "WiFi101")
}

func TestCompileWithLibraryPriority(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	_, _, err = cli.Run("core", "install", "arduino:avr@1.8.3")
	require.NoError(t, err)

	sketchName := "CompileSketchWithLibraryPriority"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	fqbn := "arduino:avr:uno"

	// Manually installs a library
	gitUrl := "https://github.com/arduino-libraries/WiFi101.git"
	manuallyInstalledLibPath := cli.SketchbookDir().Join("my-libraries", "WiFi101")
	_, err = git.PlainClone(manuallyInstalledLibPath.String(), false, &git.CloneOptions{
		URL:           gitUrl,
		ReferenceName: plumbing.NewTagReferenceName("0.16.1"),
	})
	require.NoError(t, err)

	// Install the same library we installed manually
	_, _, err = cli.Run("lib", "install", "WiFi101")
	require.NoError(t, err)

	// Create new sketch and add library include
	_, _, err = cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)
	sketchFile := sketchPath.Join(sketchName + ".ino")
	lines, err := sketchFile.ReadFileAsLines()
	require.NoError(t, err)
	lines = append([]string{"#include <WiFi101.h>\n"}, lines...)
	var data []byte
	for _, l := range lines {
		data = append(data, []byte(l)...)
	}
	err = sketchFile.WriteFile(data)
	require.NoError(t, err)

	stdout, _, err := cli.Run("compile", "-b", fqbn, sketchPath.String(), "--library", manuallyInstalledLibPath.String(), "-v")
	require.NoError(t, err)
	cliInstalledLibPath := cli.SketchbookDir().Join("libraries", "WiFi101")
	expectedOutput := [3]string{
		"Multiple libraries were found for \"WiFi101.h\"",
		"  Used: " + manuallyInstalledLibPath.String(),
		"  Not used: " + cliInstalledLibPath.String(),
	}
	require.Contains(t, string(stdout), expectedOutput[0]+"\n"+expectedOutput[1]+"\n"+expectedOutput[2]+"\n")
}

func TestRecompileWithDifferentLibrary(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	_, _, err = cli.Run("core", "install", "arduino:avr@1.8.3")
	require.NoError(t, err)

	sketchName := "RecompileCompileSketchWithDifferentLibrary"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	fqbn := "arduino:avr:uno"

	// Install library
	_, _, err = cli.Run("lib", "install", "WiFi101")
	require.NoError(t, err)

	// Manually installs a library
	gitUrl := "https://github.com/arduino-libraries/WiFi101.git"
	manuallyInstalledLibPath := cli.SketchbookDir().Join("my-libraries", "WiFi101")
	_, err = git.PlainClone(manuallyInstalledLibPath.String(), false, &git.CloneOptions{
		URL:           gitUrl,
		ReferenceName: plumbing.NewTagReferenceName("0.16.1"),
	})
	require.NoError(t, err)

	// Create new sketch and add library include
	_, _, err = cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)
	sketchFile := sketchPath.Join(sketchName + ".ino")
	lines, err := sketchFile.ReadFileAsLines()
	require.NoError(t, err)
	lines = append([]string{"#include <WiFi101.h>\n"}, lines...)
	var data []byte
	for _, l := range lines {
		data = append(data, []byte(l)...)
	}
	err = sketchFile.WriteFile(data)
	require.NoError(t, err)

	md5 := md5.Sum(([]byte(sketchPath.String())))
	sketchPathMd5 := strings.ToUpper(hex.EncodeToString(md5[:]))
	require.NotEmpty(t, sketchPathMd5)
	buildDir := paths.TempDir().Join("arduino-sketch-" + sketchPathMd5)

	// Compile sketch using library not managed by CLI
	stdout, _, err := cli.Run("compile", "-b", fqbn, "--library", manuallyInstalledLibPath.String(), sketchPath.String(), "-v")
	require.NoError(t, err)
	objPath := buildDir.Join("libraries", "WiFi101", "WiFi.cpp.o")
	require.NotContains(t, string(stdout), "Using previously compiled file: "+objPath.String())

	// Compile again using library installed from CLI
	stdout, _, err = cli.Run("compile", "-b", fqbn, sketchPath.String(), "-v")
	require.NoError(t, err)
	require.NotContains(t, string(stdout), "Using previously compiled file: "+objPath.String())
}

func TestCompileWithConflictingLibrariesInclude(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	_, _, err = cli.Run("core", "install", "arduino:avr@1.8.3")
	require.NoError(t, err)

	// Installs conflicting libraries
	gitUrl := "https://github.com/pstolarz/OneWireNg.git"
	oneWireNgLibPath := cli.SketchbookDir().Join("libraries", "onewireng_0_8_1")
	_, err = git.PlainClone(oneWireNgLibPath.String(), false, &git.CloneOptions{
		URL:           gitUrl,
		ReferenceName: plumbing.NewTagReferenceName("0.8.1"),
	})
	require.NoError(t, err)

	gitUrl = "https://github.com/PaulStoffregen/OneWire.git"
	oneWireLibPath := cli.SketchbookDir().Join("libraries", "onewire_2_3_5")
	_, err = git.PlainClone(oneWireLibPath.String(), false, &git.CloneOptions{
		URL:           gitUrl,
		ReferenceName: plumbing.NewTagReferenceName("v2.3.5"),
	})
	require.NoError(t, err)

	sketchPath := cli.CopySketch("sketch_with_conflicting_libraries_include")
	fqbn := "arduino:avr:uno"

	stdout, _, err := cli.Run("compile", "-b", fqbn, sketchPath.String(), "--verbose")
	require.NoError(t, err)
	expectedOutput := [3]string{
		"Multiple libraries were found for \"OneWire.h\"",
		"  Used: " + oneWireLibPath.String(),
		"  Not used: " + oneWireNgLibPath.String(),
	}
	require.Contains(t, string(stdout), expectedOutput[0]+"\n"+expectedOutput[1]+"\n"+expectedOutput[2]+"\n")
}

func TestCompileWithInvalidBuildOptionJson(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	_, _, err = cli.Run("core", "install", "arduino:avr@1.8.3")
	require.NoError(t, err)

	sketchName := "CompileInvalidBuildOptionsJson"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	fqbn := "arduino:avr:uno"

	// Create a test sketch
	_, _, err = cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)

	// Get the build directory
	md5 := md5.Sum(([]byte(sketchPath.String())))
	sketchPathMd5 := strings.ToUpper(hex.EncodeToString(md5[:]))
	require.NotEmpty(t, sketchPathMd5)
	buildDir := paths.TempDir().Join("arduino-sketch-" + sketchPathMd5)

	_, _, err = cli.Run("compile", "-b", fqbn, sketchPath.String(), "--verbose")
	require.NoError(t, err)

	// Breaks the build.options.json file
	buildOptionsJson := buildDir.Join("build.options.json")
	err = buildOptionsJson.WriteFile([]byte("invalid json"))
	require.NoError(t, err)

	_, _, err = cli.Run("compile", "-b", fqbn, sketchPath.String(), "--verbose")
	require.NoError(t, err)
}