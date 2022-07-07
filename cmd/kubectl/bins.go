package kubectl

import (
	"fmt"

	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"

	"eke/internal/kubectlcmd/finder"
)

// NewBinsCmd creates a new `kuberlr bins` cobra command
func NewBinsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "bins",
		Short: "Print information about the kubectl binaries found",
		Run: func(cmd *cobra.Command, args []string) {
			kFinder := finder.NewKubectlFinder("", "")
			systemBins, err := kFinder.SystemKubectlBinaries()

			fmt.Printf("%s\n", text.FgGreen.Sprint("system-wide kubectl binaries"))
			if err != nil {
				fmt.Printf("Error retrieving binaries: %v\n", err)
			} else if len(systemBins) == 0 {
				fmt.Println("No binaries found.")
			} else {
				printBinTable(systemBins)
			}

			fmt.Printf("\n\n")
			localBins, err := kFinder.LocalKubectlBinaries()

			fmt.Printf("%s\n", text.FgGreen.Sprint("local kubectl binaries"))
			if err != nil {
				fmt.Printf("Error retrieving binaries: %v\n", err)
			} else if len(localBins) == 0 {
				fmt.Println("No binaries found.")
			} else {
				printBinTable(localBins)
			}
		},
	}
}
