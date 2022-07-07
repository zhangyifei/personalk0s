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
	"bufio"
	util "eke/internal/util/utilityFunctions"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

func kubeconfigGetCmd() *cobra.Command {

	// kubeconfigGetCmd represents the kubeconfigGet(full command: eke kubeconfig get)
	var kubeconfigGetCmd = &cobra.Command{
		Use:   "get",
		Short: "Display the kubeconfig file",
		Long:  `Display the kubeconfig file`,
		Run: func(cmd *cobra.Command, args []string) {

			// the name of the kubeconfig file is set via the flag,
			// or KUBECONFIG env variable, or the default path
			// 1. Read the flag
			var kubeconfig_path string
			kubeconfig_path, _ = cmd.Flags().GetString("kubeconfig")
			if kubeconfig_path == "" {
				// 2. Read the env variable
				kubeconfig_path = os.Getenv("KUBECONFIG")
				if kubeconfig_path == "" { // 3. The last option is to use the default path
					kubeconfig_path = util.Get_kubeconfig_path() + "config"
				}
			}

			// display the contents of the kubeconfig file
			kubeconfig, err := os.Open(kubeconfig_path)
			if err != nil {
				log.Print("could not get the kubeconfig file:", err)
			}

			scanner := bufio.NewScanner(kubeconfig)

			for scanner.Scan() { // internally, it advances token based on sperator
				fmt.Println(scanner.Text()) // token in unicode-char
			}

		},
	}

	// --kubeconfig flag
	kubeconfigGetCmd.PersistentFlags().String("kubeconfig", util.Get_kubeconfig_path()+"config", "path to the kubeconfig file")

	return kubeconfigGetCmd
}
