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

package cmd

import (
	"flag"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	"istio.io/pkg/collateral"
	"istio.io/pkg/version"
)

type rootArgs struct {
	// cRPath is the path to the input IstioIstall CR.
	cRPath string
}

func addFlags(cmd *cobra.Command, rootArgs *rootArgs) {
	cmd.PersistentFlags().StringVarP(&rootArgs.cRPath, "crpath", "p", "localhost:9091",
		"The path to the input IstioIstall CR. Uses in cluster value with kubectl if unset.")
}

// GetRootCmd returns the root of the cobra command-tree.
func GetRootCmd(args []string, printf, fatalf FormatFn) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "iop",
		Short: "Command line Istio install utility.",
		Long: "This command uses the Istio operator code to generate templates, query configurations and perform " +
			"utility operations.",
	}
	rootCmd.SetArgs(args)
	rootCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	rootArgs := &rootArgs{}

	ic := installCmd(rootArgs, printf, fatalf)

	addFlags(ic, rootArgs)

	rootCmd.AddCommand(ic)
	rootCmd.AddCommand(version.CobraCommand())
	rootCmd.AddCommand(collateral.CobraCommand(rootCmd, &doc.GenManHeader{
		Title:   "Istio Operator CLI",
		Section: "operator CLI",
		Manual:  "Istio Operator CLI",
	}))

	return rootCmd
}
