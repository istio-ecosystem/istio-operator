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
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"istio.io/operator/pkg/object"
	"istio.io/operator/pkg/util"
	"istio.io/pkg/log"
)

// YAMLSuffix is the suffix of a YAML file.
const YAMLSuffix = ".yaml"

type manifestDiffArgs struct {
	// compareDir indicates comparison between directory.
	compareDir bool
	// selection specifies the resources to compare, it's comma-separated of resource indicators
	// in the form of "<kind>:<namespace>:<name>,<kind>:<namespace>:<name>".
	selection string
	// ignore specifies the resources to ignore, it's comma-separated of resource indicators
	// in the form of "<kind>:<namespace>:<name>,<kind>:<namespace>:<name>".
	ignore string
}

func addManifestDiffFlags(cmd *cobra.Command, diffArgs *manifestDiffArgs) {
	cmd.PersistentFlags().BoolVarP(&diffArgs.compareDir, "directory", "r",
		false, "compare directory")
	cmd.PersistentFlags().StringVar(&diffArgs.selection, "select", "::",
		"specifies the resources to compare, it's comma-separated of resource indicators"+
			"in the form of \"<kind>:<namespace>:<name>,<kind>:<namespace>:<name>\".")
	cmd.PersistentFlags().StringVar(&diffArgs.ignore, "ignore", "",
		"specifies the resources to ignore, it's comma-separated of resource indicators"+
			"in the form of \"<kind>:<namespace>:<name>,<kind>:<namespace>:<name>\".")
}

func manifestDiffCmd(rootArgs *rootArgs, diffArgs *manifestDiffArgs) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Compare manifests and generate diff.",
		Long:  "The diff-manifest subcommand is used to compare manifest from two files or directories.",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if diffArgs.selection != "::" && diffArgs.ignore != "" {
				logAndFatalf(rootArgs, "Cannot specify both the selected and ignored resources.")
			}
			if diffArgs.compareDir {
				compareManifestsFromDirs(rootArgs, args[0], args[1], diffArgs.selection, diffArgs.ignore)
			} else {
				compareManifestsFromFiles(rootArgs, args, diffArgs.selection, diffArgs.ignore)
			}
		}}
	return cmd
}

//compareManifestsFromFiles compares two manifest files
func compareManifestsFromFiles(rootArgs *rootArgs, args []string, selection, ignore string) {
	checkLogsOrExit(rootArgs)

	a, err := ioutil.ReadFile(args[0])
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	b, err := ioutil.ReadFile(args[1])
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	var diff string
	if ignore != "" {
		diff, err = object.ManifestDiffWithIgnore(string(a), string(b), ignore)
	} else {
		diff, err = object.ManifestDiffWithSelect(string(a), string(b), selection)
	}

	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	if diff == "" {
		fmt.Println("Manifests are identical")
	} else {
		fmt.Printf("Difference of manifests are:\n%s", diff)
		os.Exit(1)
	}
}

func yamlFileFilter(path string) bool {
	return filepath.Ext(path) == YAMLSuffix
}

//compareManifestsFromDirs compares manifests from two directories
func compareManifestsFromDirs(rootArgs *rootArgs, dirName1, dirName2, selection, ignore string) {
	checkLogsOrExit(rootArgs)

	mf1, err := util.ReadFiles(dirName1, yamlFileFilter)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	mf2, err := util.ReadFiles(dirName2, yamlFileFilter)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	var diff string
	if ignore != "" {
		diff, err = object.ManifestDiffWithIgnore(mf1, mf2, ignore)
	} else {
		diff, err = object.ManifestDiffWithSelect(mf1, mf2, selection)
	}
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	if diff == "" {
		fmt.Println("Manifests are identical")
	} else {
		fmt.Printf("Difference of manifests are:\n%s", diff)
		os.Exit(1)
	}
}
