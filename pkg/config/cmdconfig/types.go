package cmdconfig

// check how cluster config build
type EkeCmdConfig struct {
	EkeKubectlConfig EkeKubectlConfig `mapstructure:"ekeKubectlConfig"`
}

type EkeKubectlConfig struct {
	AllowDownload bool   `mapstructure:"allowDownload"`
	SystemPath    string `mapstructure:"systemPath"`
	Timeout       int    `mapstructure:"timeout"`
}
