package istiocontrolplane

import (
	"path/filepath"

	"github.com/spf13/cobra"
)

// Options represents the details used to configure the controller.
type Options struct {
	// BaseChartPath is the abosolute path used as the base path when a relative path is specified in
	// IstioControlPlane.Spec.ChartPath
	BaseChartPath string
	// DefaultChartPath is the relative path used added to BaseChartPath when no value is specified in
	// IstioControlPlane.Spec.ChartPath
	DefaultChartPath string
}

// ControllerOptions represents the options used by the controller
var controllerOptions = &Options{
	// XXX: update this once we add charts to the operator
	BaseChartPath:    "/etc/istio-operator/helm",
	DefaultChartPath: "istio",
}

// AttachCobraFlags attaches a set of Cobra flags to the given Cobra command.
//
// Cobra is the command-line processor that Istio uses. This command attaches
// the set of flags used to configure the IstioControlPlane reconciler
func AttachCobraFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&controllerOptions.BaseChartPath, "base-chart-path", "",
		"The absolute path to a directory containing nested charts, e.g. /etc/istio-operator/helm.  "+
			"This will be used as the base path for any IstioControlPlane instances specifying a relative ChartPath.")
	cmd.PersistentFlags().StringVar(&controllerOptions.BaseChartPath, "default-chart-path", "",
		"A path relative to base-chart-path containing charts to be used when no ChartPath is specified by an IstioControlPlane resource, e.g. 1.1.0/istio")
}

func calculateChartPath(inputPath string) string {
	if len(inputPath) == 0 {
		return filepath.Join(controllerOptions.BaseChartPath, controllerOptions.DefaultChartPath)
	} else if filepath.IsAbs(inputPath) {
		return inputPath
	}
	return filepath.Join(controllerOptions.BaseChartPath, inputPath)
}
