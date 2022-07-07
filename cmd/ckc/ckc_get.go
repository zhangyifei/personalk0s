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
	"fmt"

	"github.com/spf13/cobra"
)

func ckcGetCmd() *cobra.Command {
	// getCmd represents the get command
	var getCmd = &cobra.Command{
		Use:   "get",
		Short: "Get EWS client key and certificate",
		Long: `If user has client key/certificate cached in local already, it will just display that stored in local.
	otherwise, it will fetch it from remote and user's signum and password will be asked.`,
		Run: func(cmd *cobra.Command, args []string) {
			eke_cache := util.Get_eke_path()
			userCert, userKey := util.GetCertAndKey(eke_cache)
			fmt.Println(util.CreateOutput(userCert, userKey))
		},
	}

	return getCmd
}
