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

package cache

import (
	"context"
	"os"

	"github.com/arduino/arduino-cli/commands/cache"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initCleanCommand() *cobra.Command {
	cleanCommand := &cobra.Command{
		Use:     "clean",
		Short:   tr("Delete Boards/Library Manager download cache."),
		Long:    tr("Delete contents of the `directories.downloads` folder, where archive files are staged during installation of libraries and boards platforms."),
		Example: "  " + os.Args[0] + " cache clean",
		Args:    cobra.NoArgs,
		Run:     runCleanCommand,
	}
	return cleanCommand
}

func runCleanCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino-cli cache clean`")

	_, err := cache.CleanDownloadCacheDirectory(context.Background(), &rpc.CleanDownloadCacheDirectoryRequest{})
	if err != nil {
		feedback.Fatal(tr("Error cleaning caches: %v", err), feedback.ErrGeneric)
	}
}
