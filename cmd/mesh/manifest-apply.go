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
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"istio.io/operator/pkg/manifest"
	opversion "istio.io/operator/version"
)

type manifestApplyArgs struct {
	// inFilename is the path to the input IstioControlPlane CR.
	inFilename string
	// kubeConfigPath is the path to kube config file.
	kubeConfigPath string
	// readinessTimeout is maximum time to wait for all Istio resources to be ready.
	readinessTimeout time.Duration
	// wait is flag that indicates whether to wait resources ready before exiting.
	wait bool
	// set is a string with element format "path=value" where path is an IstioControlPlane path and the value is a
	// value to set the node at that path to.
	set []string
}

func addManifestApplyFlags(cmd *cobra.Command, args *manifestApplyArgs) {
	cmd.PersistentFlags().StringVarP(&args.inFilename, "filename", "f", "", filenameFlagHelpStr)
	cmd.PersistentFlags().StringVarP(&args.kubeConfigPath, "kubeconfig", "c", "", "Path to kube config.")
	cmd.PersistentFlags().DurationVar(&args.readinessTimeout, "readiness-timeout", 300*time.Second, "Maximum time to wait for all Istio resources to be ready."+
		"--wait must be set for this flag to apply.")
	cmd.PersistentFlags().BoolVarP(&args.wait, "wait", "w", false, "Wait, if set will wait until all Pods, Services, and minimum number of Pods "+
		"of a Deployment are in a ready state before the command exits. It will wait for a maximum duration of --readiness-timeout seconds.")
	cmd.PersistentFlags().StringSliceVarP(&args.set, "set", "s", nil, setFlagHelpStr)
}

func manifestApplyCmd(rootArgs *rootArgs, maArgs *manifestApplyArgs) *cobra.Command {
	return &cobra.Command{
		Use:   "apply",
		Short: "Generates and applies Istio install manifest.",
		Long:  "The apply subcommand is used to generate an Istio install manifest and apply it to a cluster.",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			l := newLogger(rootArgs.logToStdErr, cmd.OutOrStdout(), cmd.OutOrStderr())
			manifestApply(rootArgs, maArgs, l)
		}}
}

func manifestApply(args *rootArgs, maArgs *manifestApplyArgs, l *logger) {
	if err := configLogs(args); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Could not configure logs: %s", err)
		os.Exit(1)
	}

	overlayFromSet, err := makeTreeFromSetList(maArgs.set)
	if err != nil {
		l.lfatal(err.Error())
	}
	manifests, err := genManifests(maArgs.inFilename, overlayFromSet)
	if err != nil {
		l.lfatal("Could not generate manifest: ", err)
	}

	out, err := manifest.ApplyAll(manifests, opversion.OperatorBinaryVersion, args.dryRun, args.verbose, maArgs.wait, maArgs.readinessTimeout)
	if err != nil {
		l.lfatal("Failed to apply manifest with kubectl client: ", err)
	}

	for cn := range manifests {
		cs := fmt.Sprintf("Output for component %s:", cn)
		l.lprintf("\n%s\n%s\n", cs, strings.Repeat("=", len(cs)))
		if out[cn].Err != nil {
			l.lprint("Error: ", out[cn].Err, "\n")
		}
		if strings.TrimSpace(out[cn].Stderr) != "" {
			l.lprint("Error detail:\n", out[cn].Stderr, "\n")
		}
		if strings.TrimSpace(out[cn].Stdout) != "" {
			l.lprint("Stdout:\n", out[cn].Stdout, "\n")
		}
		if args.verbose {
			l.lprint("Manifest:\n\n", out[cn].Manifest, "\n")
		}
	}
}
