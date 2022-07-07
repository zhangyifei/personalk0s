package showconfig

import (
	"eke/pkg/config"
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type CmdOpts config.CLIOptions

func NewShowconfigCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "showconfig",
		Short: "Show Cmd Config",

		Run: func(cmd *cobra.Command, args []string) {
			c := CmdOpts(config.GetCmdOpts())
			configdata, _ := yaml.Marshal(c.CmdConfig)
			fmt.Println(string(configdata))
		},
	}
	return cmd

}
