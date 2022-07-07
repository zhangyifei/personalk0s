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
package kubeconfig

import (
	util "eke/internal/util/utilityFunctions"
	"fmt"

	"github.com/spf13/cobra"
)

func kubeconfigAuthCmd() *cobra.Command {
	// authCmd represents the auth command
	var kubeconfigAuthCmd = &cobra.Command{
		Use:   "auth",
		Short: "Gets user .crt and .key from EWS",
		Long:  `Checks for cached .crt and .key files, if not it will request them from EWS`,
		Run: func(cmd *cobra.Command, args []string) {
			eke_cache := util.Get_eke_path()
			userCert, userKey := util.GetCertAndKey(eke_cache)
			fmt.Println(util.CreateOutput(userCert, userKey))
		},
	}
	return kubeconfigAuthCmd
}
