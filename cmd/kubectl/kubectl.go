/*
Copyright 2021 eke authors

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
package kubectl

import (
	"eke/pkg/config"

	"github.com/spf13/cobra"
)

type CmdOpts config.CLIOptions

const LIST_BINS_CMD = "bins"

const GET_BIN_CMD = "get-bin"

func NewKubectlCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "kubectl [original kubectl commands | get-bin or bins]",
		Short: "eke kubectl",
		Long: `eke kubectl has two types of commands:
		1, "get-bin" and "bins" used to manage different versions of kubectl
		2, Other normal kubectl commands`,

		Run: func(cmd *cobra.Command, args []string) {
			c := CmdOpts(config.GetCmdOpts())

			if len(args) > 0 {
				subcmd := args[0]

				if subcmd != LIST_BINS_CMD && subcmd != GET_BIN_CMD {
					kubectlWrapperMode(c.CmdConfig.EkeKubectlConfig, args)
				} else {
					return
				}
			} else {
				cmd.Help()
			}
		},
	}
	cmd.SetUsageTemplate(USAGE_TEMPLATE)
	cmd.AddCommand(NewBinsCmd())
	cmd.AddCommand(NewGetbinCmd())
	return cmd
}
