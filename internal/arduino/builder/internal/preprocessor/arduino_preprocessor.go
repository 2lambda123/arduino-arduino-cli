// This file is part of arduino-cli.
//
// Copyright 2023 ARDUINO SA (http://www.arduino.cc/)
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

package preprocessor

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"runtime"

	"github.com/arduino/arduino-cli/internal/arduino/builder/internal/utils"
	"github.com/arduino/arduino-cli/internal/arduino/sketch"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
)

// PreprocessSketchWithArduinoPreprocessor performs preprocessing of the arduino sketch
// using arduino-preprocessor (https://github.com/arduino/arduino-preprocessor).
func PreprocessSketchWithArduinoPreprocessor(
	sk *sketch.Sketch, buildPath *paths.Path, includeFolders paths.PathList,
	lineOffset int, buildProperties *properties.Map, onlyUpdateCompilationDatabase bool,
) (*Result, error) {
	verboseOut := &bytes.Buffer{}
	normalOut := &bytes.Buffer{}
	if err := buildPath.Join("preproc").MkdirAll(); err != nil {
		return nil, err
	}

	sourceFile := buildPath.Join("sketch", sk.MainFile.Base()+".cpp")
	targetFile := buildPath.Join("preproc", "sketch_merged.cpp")
	gccResult, err := GCC(sourceFile, targetFile, includeFolders, buildProperties)
	verboseOut.Write(gccResult.Stdout())
	verboseOut.Write(gccResult.Stderr())
	if err != nil {
		return nil, err
	}

	arduiniPreprocessorProperties := properties.NewMap()
	arduiniPreprocessorProperties.Set("tools.arduino-preprocessor.path", "{runtime.tools.arduino-preprocessor.path}")
	arduiniPreprocessorProperties.Set("tools.arduino-preprocessor.cmd.path", "{path}/arduino-preprocessor")
	arduiniPreprocessorProperties.Set("tools.arduino-preprocessor.pattern", `"{cmd.path}" "{source_file}" -- -std=gnu++11`)
	arduiniPreprocessorProperties.Set("preproc.macros.flags", "-w -x c++ -E -CC")
	arduiniPreprocessorProperties.Merge(buildProperties)
	arduiniPreprocessorProperties.Merge(arduiniPreprocessorProperties.SubTree("tools").SubTree("arduino-preprocessor"))
	arduiniPreprocessorProperties.SetPath("source_file", targetFile)
	pattern := arduiniPreprocessorProperties.Get("pattern")
	if pattern == "" {
		return nil, errors.New(tr("arduino-preprocessor pattern is missing"))
	}

	commandLine := arduiniPreprocessorProperties.ExpandPropsInString(pattern)
	parts, err := properties.SplitQuotedString(commandLine, `"'`, false)
	if err != nil {
		return nil, err
	}

	command, err := paths.NewProcess(nil, parts...)
	if err != nil {
		return nil, err
	}
	if runtime.GOOS == "windows" {
		// chdir in the uppermost directory to avoid UTF-8 bug in clang (https://github.com/arduino/arduino-preprocessor/issues/2)
		command.SetDir(filepath.VolumeName(parts[0]) + "/")
	}

	verboseOut.WriteString(commandLine)
	commandStdOut, commandStdErr, err := command.RunAndCaptureOutput(context.Background())
	verboseOut.Write(commandStdErr)
	if err != nil {
		return &Result{args: gccResult.Args(), stdout: verboseOut.Bytes(), stderr: normalOut.Bytes()}, err
	}
	result := utils.NormalizeUTF8(commandStdOut)

	destFile := buildPath.Join(sk.MainFile.Base() + ".cpp")
	if err := destFile.WriteFile(result); err != nil {
		return &Result{args: gccResult.Args(), stdout: verboseOut.Bytes(), stderr: normalOut.Bytes()}, err
	}
	return &Result{args: gccResult.Args(), stdout: verboseOut.Bytes(), stderr: normalOut.Bytes()}, err
}
