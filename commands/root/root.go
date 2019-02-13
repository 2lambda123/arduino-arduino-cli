/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package root

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/arduino/arduino-cli/output"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/mattn/go-colorable"

	"github.com/arduino/go-paths-helper"

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/commands/board"
	"github.com/arduino/arduino-cli/commands/compile"
	"github.com/arduino/arduino-cli/commands/config"
	"github.com/arduino/arduino-cli/commands/core"
	"github.com/arduino/arduino-cli/commands/generatedocs"
	"github.com/arduino/arduino-cli/commands/lib"
	"github.com/arduino/arduino-cli/commands/sketch"
	"github.com/arduino/arduino-cli/commands/upload"
	"github.com/arduino/arduino-cli/commands/version"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/arduino/arduino-cli/configs"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Init prepares the cobra root command.
func Init() *cobra.Command {
	command := &cobra.Command{
		Use:              "arduino-cli",
		Short:            "Arduino CLI.",
		Long:             "Arduino Command Line Interface (arduino-cli).",
		Example:          "  " + commands.AppName + " <command> [flags...]",
		PersistentPreRun: preRun,
	}
	command.PersistentFlags().BoolVar(&commands.GlobalFlags.Debug, "debug", false, "Enables debug output (super verbose, used to debug the CLI).")
	command.PersistentFlags().StringVar(&commands.GlobalFlags.Format, "format", "text", "The output format, can be [text|json].")
	command.PersistentFlags().StringVar(&yamlConfigFile, "config-file", "", "The custom config file (if not specified the default will be used).")
	command.AddCommand(board.InitCommand())
	command.AddCommand(compile.InitCommand())
	command.AddCommand(config.InitCommand())
	command.AddCommand(core.InitCommand())
	command.AddCommand(generatedocs.InitCommand())
	command.AddCommand(lib.InitCommand())
	// command.AddCommand(login.InitCommand())
	// command.AddCommand(logout.InitCommand())
	command.AddCommand(sketch.InitCommand())
	command.AddCommand(upload.InitCommand())
	// command.AddCommand(validate.InitCommand())
	command.AddCommand(version.InitCommand())
	return command
}

var yamlConfigFile string

func preRun(cmd *cobra.Command, args []string) {
	// Reset logrus if debug flag changed.
	if !commands.GlobalFlags.Debug {
		// Discard logrus output if no debug.
		logrus.SetOutput(ioutil.Discard)
	} else {
		// Else print on stderr.

		// Workaround to get colored output on windows
		if terminal.IsTerminal(int(os.Stdout.Fd())) {
			logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})
		}
		logrus.SetOutput(colorable.NewColorableStdout())
		commands.ErrLogrus.Out = colorable.NewColorableStderr()
		formatter.SetLogger(commands.ErrLogrus)
	}
	initConfigs()

	logrus.Info(commands.AppName + "-" + commands.Version)
	logrus.Info("Starting root command preparation (`arduino`)")
	if !formatter.IsSupported(commands.GlobalFlags.Format) {
		logrus.WithField("inserted format", commands.GlobalFlags.Format).Warn("Unsupported format, using text as default")
		commands.GlobalFlags.Format = "text"
	}
	formatter.SetFormatter(commands.GlobalFlags.Format)
	if commands.GlobalFlags.Format != "text" {
		output.SetOutputKind(output.JSON)
	}

	logrus.Info("Formatter set")
	if !formatter.IsCurrentFormat("text") {
		cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
			logrus.Warn("Calling help on JSON format")
			formatter.PrintErrorMessage("Invalid Call : should show Help, but it is available only in TEXT mode.")
			os.Exit(commands.ErrBadCall)
		})
	}
}

// initConfigs initializes the configuration from the specified file.
func initConfigs() {
	// Start with default configuration
	if conf, err := configs.NewConfiguration(); err != nil {
		logrus.WithError(err).Error("Error creating default configuration")
		formatter.PrintError(err, "Error creating default configuration")
		os.Exit(commands.ErrGeneric)
	} else {
		commands.Config = conf
	}

	// Navigate through folders
	pwd, err := filepath.Abs(".")
	if err != nil {
		logrus.WithError(err).Warn("Did not manage to find current path")
	}

	commands.Config.Navigate("/", pwd)
	commands.Config.LoadFromYAML(commands.Config.ConfigFile)

	if yamlConfigFile != "" {
		commands.Config.ConfigFile = paths.New(yamlConfigFile)
		if err := commands.Config.LoadFromYAML(commands.Config.ConfigFile); err != nil {
			logrus.WithError(err).Warn("Did not manage to get config file, using default configuration")
		}
	}

	logrus.Info("Initiating configuration")

	if commands.Config.IsBundledInDesktopIDE() {
		logrus.Info("CLI is bundled into the IDE")
		err := commands.Config.LoadFromDesktopIDEPreferences()
		if err != nil {
			logrus.WithError(err).Warn("Did not manage to get config file of IDE, using default configuration")
		}
	} else {
		logrus.Info("CLI is not bundled into the IDE")
	}
	commands.Config.LoadFromEnv()
	logrus.Info("Configuration set")
}
