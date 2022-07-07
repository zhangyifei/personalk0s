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
package cmd

import ( //"eke/cmd/auth"
	// "log"
	"eke/cmd/ckc"
	"eke/cmd/kubeconfig"
	"eke/cmd/kubectl"
	"eke/cmd/showconfig"
	"eke/cmd/version"
	"eke/pkg/build"
	"eke/pkg/config"
	"log"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type cliOpts config.CLIOptions

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "eke",
		Short: "EWS Kubernetes Engine",
		Long:  `EWS Kubernetes Engine`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			c := cliOpts(config.GetCmdOpts())

			if c.Verbose {
				logrus.SetLevel(logrus.InfoLevel)
			}

			// set DEBUG from env, or from command flag
			if viper.GetString("debug") != "" || c.Debug {
				logrus.SetLevel(logrus.DebugLevel)
				go func() {
					log.Println("starting debug server under", c.DebugListenOn)
					log.Println(http.ListenAndServe(c.DebugListenOn, nil))
				}()
			}
		},
	}
	// This is to remove the help command from list of Available commands
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	// Disable auto completion
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.DisableAutoGenTag = true

	longDesc := "EWS Kubernetes Engine"
	if build.EulaNotice != "" {
		longDesc = longDesc + "\n" + build.EulaNotice
	}
	rootCmd.Long = longDesc

	// workaround for the data-dir location input for the kubectl command
	rootCmd.PersistentFlags().AddFlagSet(config.GetKubeCtlFlagSet())
	rootCmd.PersistentFlags().AddFlagSet(config.GetPersistentFlagSet())

	rootCmd.AddCommand(ckc.NewCkcCmd())
	rootCmd.AddCommand(kubeconfig.NewKubeconfigCmd())
	rootCmd.AddCommand(version.NewVersionCmd())
	rootCmd.AddCommand(showconfig.NewShowconfigCmd())
	rootCmd.AddCommand(kubectl.NewKubectlCmd())

	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(newRootCmd().Execute())
}
