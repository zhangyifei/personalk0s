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
	util "eke/internal/util/utilityFunctions"
	"log"

	"github.com/spf13/cobra"
)

// to store flag values
var signum, pass string

func ckcRenewCmd() *cobra.Command {

	// renewCmd represents the renew command
	var renewCmd = &cobra.Command{
		Use:   "renew",
		Short: "Renew EWS client key and certificate",
		Long:  `Renew EWS client key and certificate`,
		Run: func(cmd *cobra.Command, args []string) {

			// Get signum and password from the user if not already set by corresponding flags
			var err error
			//signum, _ := cmd.Flags().GetString("userid")
			if signum == "" {
				signum, err = util.GetUserSignum()
				if err != nil {
					log.Fatalln("error occured while prompting for credentials:", err)
				}
			}
			//pass, _ := cmd.Flags().GetString("password")
			if pass == "" {
				pass, err = util.GetUserPassword()
				if err != nil {
					log.Fatalln("error occured while prompting for credentials:", err)
				}
			}

			eke_cache := util.Get_eke_path()

			util.RequestCertAndKeyFromEWS(eke_cache, signum, pass, true)
		},
	}

	// --userid flag
	renewCmd.PersistentFlags().StringVarP(&signum, "userid", "u", "", "ericsson signum")
	// --password flag
	renewCmd.PersistentFlags().StringVarP(&pass, "password", "p", "", "user password")
	return renewCmd
}
