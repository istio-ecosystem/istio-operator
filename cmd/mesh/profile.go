package mesh

import (
	"github.com/spf13/cobra"
)

func profileCmd(args *rootArgs) *cobra.Command {
	pc := &cobra.Command{
		Use:   "profile",
		Short: "Commands related to Istio configuration profiles.",
		Long:  "The profile subcommand is list, dump or diff Istio configuration profiles.",
	}

	plArgs := &profileListArgs{}
	pdArgs := &profileDumpArgs{}

	plc := profileListCmd(args, plArgs)
	pdc := profileDumpCmd(args, pdArgs)
	pdfc := profileDiffCmd(args)

	addFlags(plc, args)
	addFlags(pdc, args)
	addFlags(pdfc, args)

	addProfileListFlags(plc, plArgs)
	addProfileDumpFlags(pdc, pdArgs)

	pc.AddCommand(plc)
	pc.AddCommand(pdc)
	pc.AddCommand(pdfc)

	return pc
}
