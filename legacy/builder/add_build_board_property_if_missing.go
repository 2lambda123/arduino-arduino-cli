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

package builder

import (
	"strings"

	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/types"
)

type AddBuildBoardPropertyIfMissing struct{}

func (*AddBuildBoardPropertyIfMissing) Run(ctx *types.Context) error {
	packages := ctx.Hardware

	for _, aPackage := range packages {
		for _, platform := range aPackage.Platforms {
			for _, platformRelease := range platform.Releases {
				for _, board := range platformRelease.Boards {
					if board.Properties.Get("build.board") == "" {
						board.Properties.Set("build.board", strings.ToUpper(platform.Architecture+"_"+board.BoardID))
						ctx.Info(tr("Warning: Board %[1]s doesn't define a %[2]s preference. Auto-set to: %[3]s",
							aPackage.Name+":"+platform.Architecture+":"+board.BoardID,
							"'build.board'",
							board.Properties.Get(constants.BUILD_PROPERTIES_BUILD_BOARD)))
					}
				}
			}
		}
	}

	return nil
}