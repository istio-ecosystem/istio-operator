/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kubectlcmd

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"istio.io/operator/pkg/util"
	"istio.io/pkg/log"
)

// New creates a Client that runs kubectl available on the path with default authentication
func New() *Client {
	return &Client{cmdSite: &console{}}
}

// Client provides an interface to kubectl
type Client struct {
	cmdSite commandSite
}

// commandSite allows for tests to mock cmd.Run() events
type commandSite interface {
	Run(*exec.Cmd) error
}
type console struct {
}

func (console) Run(c *exec.Cmd) error {
	return c.Run()
}

// kubectlParams is a set of params passed to kubectl.
type kubectlParams struct {
	dryRun     bool
	verbose    bool
	kubeconfig string
	context    string
	namespace  string
	stdin      string
	output     string
	extraArgs  []string
}

// Apply runs the `kubectl apply` command with parameters:
// dryRun to not actually run the cmd, verbose to not print manifest
// kubeconfig, context to identify a cluster,
// namespace to locate a k8s namespace, and
// returns stdout, stderr, error getting from running this `kubectl` command.
func (c *Client) Apply(dryRun, verbose bool, kubeconfig, context, namespace string,
	manifest string, extraArgs ...string) (string, string, error) {
	if strings.TrimSpace(manifest) == "" {
		log.Infof("Empty manifest, not running kubectl apply.")
		return "", "", nil
	}
	subcmds := []string{"apply"}
	params := &kubectlParams{
		dryRun:     dryRun,
		verbose:    verbose,
		kubeconfig: kubeconfig,
		context:    context,
		namespace:  namespace,
		stdin:      manifest,
		output:     "",
		extraArgs:  extraArgs,
	}
	return c.kubectl(subcmds, params)
}

// Delete runs the `kubectl delete` command with parameters:
// dryRun to not actually run the cmd, verbose to not print manifest
// kubeconfig, context to identify a cluster,
// namespace to locate a k8s namespace, and
// returns stdout, stderr, error getting from running this `kubectl` command.
func (c *Client) Delete(dryRun, verbose bool, kubeconfig, context, namespace string,
	manifest string, extraArgs ...string) (string, string, error) {
	if strings.TrimSpace(manifest) == "" {
		log.Infof("Empty manifest, not running kubectl delete.")
		return "", "", nil
	}
	subcmds := []string{"delete"}
	params := &kubectlParams{
		dryRun:     dryRun,
		verbose:    verbose,
		kubeconfig: kubeconfig,
		context:    context,
		namespace:  namespace,
		stdin:      manifest,
		output:     "",
		extraArgs:  extraArgs,
	}
	return c.kubectl(subcmds, params)
}

// GetAll runs the `kubectl get all` with with parameters:
// kubeconfig, context to identify a cluster,
// namespace to locate a k8s namespace, and
// returns stdout, stderr, error getting from running this `kubectl` command.
func (c *Client) GetAll(kubeconfig, context, namespace, output string,
	extraArgs ...string) (string, string, error) {
	subcmds := []string{"get", "all"}
	params := &kubectlParams{
		dryRun:     false,
		verbose:    false,
		kubeconfig: kubeconfig,
		context:    context,
		namespace:  namespace,
		stdin:      "",
		output:     output,
		extraArgs:  extraArgs,
	}
	return c.kubectl(subcmds, params)
}

// GetConfig runs the `kubectl get cm` command with parameters:
// dryRun to not actually run the cmd, verbose to not print stdin
// kubeconfig, context to identify a cluster,
// namespace to locate a k8s namespace, and
// returns stdout, stderr, error getting from running this `kubectl` command.
func (c *Client) GetConfig(kubeconfig, context, name, namespace, output string,
	extraArgs ...string) (string, string, error) {
	subcmds := []string{"get", "cm", name}
	params := &kubectlParams{
		dryRun:     false,
		verbose:    false,
		kubeconfig: kubeconfig,
		context:    context,
		namespace:  namespace,
		stdin:      "",
		output:     output,
		extraArgs:  extraArgs,
	}
	return c.kubectl(subcmds, params)
}

func logAndPrint(v ...interface{}) {
	s := fmt.Sprintf(v[0].(string), v[1:]...)
	log.Infof(s)
	fmt.Println(s)
}

// Apply runs the `kubectl` command by specifying subcommands in subcmds with parameters:
// dryRun to not actually run the cmd, verbose to not print stdin
// kubeconfig, context to identify a cluster,
// namespace to locate a k8s namespace, and
// returns stdout, stderr, error getting from running this `kubectl` command.
func (c *Client) kubectl(subcmds []string, params *kubectlParams) (string, string, error) {
	hasStdin := strings.TrimSpace(params.stdin) != ""
	args := subcmds
	if params.kubeconfig != "" {
		args = append(args, "--kubeconfig", params.kubeconfig)
	}
	if params.context != "" {
		args = append(args, "--context", params.context)
	}
	if params.namespace != "" {
		args = append(args, "-n", params.namespace)
	}
	if params.output != "" {
		args = append(args, "-o", params.output)
	}
	args = append(args, params.extraArgs...)

	if hasStdin {
		args = append(args, "-f", "-")
	}

	cmd := exec.Command("kubectl", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	cmdStr := strings.Join(args, " ")
	if hasStdin {
		cmd.Stdin = strings.NewReader(params.stdin)
		if params.verbose {
			cmdStr += "\n" + params.stdin
		} else {
			cmdStr += " <use --verbose to see stdin string> \n"
		}
	}

	if params.dryRun {
		logAndPrint("dry run mode: would be running this cmd:\n%s\n", cmdStr)
		return "", "", nil
	}

	log.Infof("running command:\n%s\n", cmdStr)
	err := c.cmdSite.Run(cmd)
	csError := util.ConsolidateLog(stderr.String())

	if err != nil {
		logAndPrint("error running kubectl: %s", err)
		return stdout.String(), csError, fmt.Errorf("error running kubectl: %s", err)
	}

	logAndPrint("kubectl success")

	return stdout.String(), csError, nil
}
