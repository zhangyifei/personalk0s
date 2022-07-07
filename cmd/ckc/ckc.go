/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

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
package ckc

import (
	"log"

	"github.com/spf13/cobra"
)

func NewCkcCmd() *cobra.Command {

	// ckcCmd represents the ckc command
	var ckcCmd = &cobra.Command{
		Use:   "ckc",
		Short: "Get user's k8s client key and certificate from EWS",
		Long:  `Get user's k8s client key and certificate from EWS`,
		Run: func(cmd *cobra.Command, args []string) {

			if len(args) == 0 {
				if err := cmd.Help(); err != nil {
					log.Println(err)
				}
			}
		},
	}

	ckcCmd.AddCommand(ckcGetCmd())
	ckcCmd.AddCommand(ckcRenewCmd())
	return ckcCmd

}
