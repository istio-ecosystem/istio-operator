// Copyright 2019 Istio Authors
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

package mesh

import (
	"fmt"

	"github.com/spf13/cobra"
)

type profileListArgs struct {
	// set is a string with element format "path=value" where path is an IstioControlPlane path and the value is a
	// value to set the node at that path to.
	set string
}

func addProfileListFlags(cmd *cobra.Command, args *profileListArgs) {
	cmd.PersistentFlags().StringVarP(&args.set, "set", "s", "", setFlagHelpStr)
}

func profileListCmd(rootArgs *rootArgs, plArgs *profileListArgs) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Lists available Istio configuration profiles.",
		Long:  "The list subcommand is used to list available Istio configuration profiles.",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			profileList(rootArgs, plArgs)
		}}

}

func profileList(args *rootArgs, _ *profileListArgs) {
	checkLogsOrExit(args)
	fmt.Println("This command is not yet implemented.")
}
