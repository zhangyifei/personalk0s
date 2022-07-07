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
	//"fmt"
	util "eke/internal/util/utilityFunctions"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func kubeconfigResetCmd() *cobra.Command {
	// resetCmd represents the reset command
	var resetCmd = &cobra.Command{
		Use:   "reset",
		Short: "Clean up kubeconfig file and ews config dir ~/.eke/",
		Long:  `Clean up kubeconfig file and ews config dir ~/.eke/`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := RunReset(cmd)
			if err != nil {
				log.Fatal(err)
			}
			return nil
		},
	}
	// --kubeconfig flag
	resetCmd.PersistentFlags().String("kubeconfig", util.Get_kubeconfig_path()+"config", "path to assign to the created kubeconfig file")

	return resetCmd
}

func RunReset(cmd *cobra.Command) error {

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

	// remove the kubeconfig file
	err := os.Remove(kubeconfig_path)
	if err != nil {
		// This error should not be fatal as kubeconfig file might not exist yet
		log.Println("error while clearing kubeconfig:", err)
	}

	// clear the eke cache
	eke_cache := util.Get_eke_path()
	d, err := os.Open(eke_cache)
	if err != nil {
		//log.Fatal("error while clearing the eke cache:", err)
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		//log.Fatal("error while clearing the eke cache:", err)
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(eke_cache, name))
		if err != nil {
			//log.Fatal("error while clearing the eke cache:", err)
			return err
		}
	}

	fmt.Fprintf(cmd.OutOrStdout(), "kubeconfig and eke cache cleared successfully\n")
	return nil
}
