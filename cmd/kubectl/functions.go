package kubectl

import (
	"eke/internal/kubectlcmd/finder"
	"eke/internal/kubectlcmd/osexec"
	"eke/pkg/config/cmdconfig"
	"log"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
)

func printBinTable(bins finder.KubectlBinaries) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Version", "Binary"})
	for i, b := range bins {
		t.AppendRow([]interface{}{i + 1, b.Version, b.Path})
	}
	t.Render()
}

func kubectlWrapperMode(config cmdconfig.EkeKubectlConfig, args []string) {

	kFinder := finder.NewKubectlFinder("", config.SystemPath)
	versioner := finder.NewVersioner(kFinder)
	version, err := versioner.KubectlVersionToUse(int64(config.Timeout))
	if err != nil {
		log.Fatal(err)
	}

	kubectlBin, err := versioner.EnsureCompatibleKubectlAvailable(
		version,
		config.AllowDownload)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(os.Args)
	childArgs := append([]string{kubectlBin}, args...)
	err = osexec.Exec(kubectlBin, childArgs, os.Environ())
	log.Fatal(err)
}
