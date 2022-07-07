/*
Copyright 2022 eke authors

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
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	k8s "k8s.io/client-go/kubernetes"
	cloudprovider "k8s.io/cloud-provider"

	"eke/pkg/apis/eke/v1beta1"

	"eke/pkg/constant"

	cmdconfig "eke/pkg/config/cmdconfig"
)

var (
	CfgFile        string
	CmdCfgFile     string
	DataDir        string
	Debug          bool
	DebugListenOn  string
	StatusSocket   string
	EkeVars        constant.CfgVars
	workerOpts     WorkerOptions
	Verbose        bool
	controllerOpts ControllerOptions
)

// This struct holds all the CLI options & settings required by the
// different eke sub-commands
type CLIOptions struct {
	WorkerOptions
	ControllerOptions
	CfgFile          string
	ClusterConfig    *v1beta1.ClusterConfig
	NodeConfig       *v1beta1.ClusterConfig
	CmdConfig        *cmdconfig.EkeCmdConfig
	Debug            bool
	DebugListenOn    string
	DefaultLogLevels map[string]string
	EkeVars          constant.CfgVars
	KubeClient       k8s.Interface
	Logging          map[string]string // merged outcome of default log levels and cmdLoglevels
	Verbose          bool
}

// Shared controller cli flags
type ControllerOptions struct {
	EnableWorker      bool
	SingleNode        bool
	NoTaints          bool
	DisableComponents []string

	EnableEkeCloudProvider          bool
	EkeCloudProviderPort            int
	EkeCloudProviderUpdateFrequency time.Duration
	EnableDynamicConfig             bool
	EnableMetricsScraper            bool
}

// Shared worker cli flags
type WorkerOptions struct {
	APIServer        string
	CIDRRange        string
	CloudProvider    bool
	ClusterDNS       string
	CmdLogLevels     map[string]string
	CriSocket        string
	KubeletExtraArgs string
	Labels           []string
	Taints           []string
	TokenFile        string
	TokenArg         string
	WorkerProfile    string
}

func DefaultLogLevels() map[string]string {
	return map[string]string{
		"etcd":                    "info",
		"containerd":              "info",
		"konnectivity-server":     "1",
		"kube-apiserver":          "1",
		"kube-controller-manager": "1",
		"kube-scheduler":          "1",
		"kubelet":                 "1",
		"kube-proxy":              "1",
	}
}

func GetPersistentFlagSet() *pflag.FlagSet {
	flagset := &pflag.FlagSet{}
	flagset.BoolVarP(&Debug, "debug", "d", false, "Debug logging (default: false)")
	flagset.BoolVarP(&Verbose, "verbose", "v", false, "Verbose logging (default: false)")
	flagset.StringVar(&DataDir, "data-dir", "", "Data Directory for eke (default: /var/lib/eke). DO NOT CHANGE for an existing setup, things will break!")
	flagset.StringVar(&CmdCfgFile, "cmd-config", "", "the directory for eke commands config file eke.cmd.yaml")
	flagset.StringVar(&StatusSocket, "status-socket", filepath.Join(EkeVars.RunDir, "status.sock"), "Full file path to the socket file.")
	flagset.StringVar(&DebugListenOn, "debugListenOn", ":6060", "Http listenOn for Debug pprof handler")
	return flagset
}

// XX: not a pretty hack, but we need the data-dir flag for the kubectl subcommand
// XX: when other global flags cannot be used (specifically -d and -c)
func GetKubeCtlFlagSet() *pflag.FlagSet {
	flagset := &pflag.FlagSet{}
	flagset.StringVar(&DataDir, "data-dir", "", "Data Directory for eke (default: /var/lib/eke). DO NOT CHANGE for an existing setup, things will break!")
	flagset.BoolVar(&Debug, "debug", false, "Debug logging (default: false)")
	return flagset
}

func GetCriSocketFlag() *pflag.FlagSet {
	flagset := &pflag.FlagSet{}
	flagset.StringVar(&workerOpts.CriSocket, "cri-socket", "", "container runtime socket to use, default to internal containerd. Format: [remote|docker]:[path-to-socket]")
	return flagset
}

func GetWorkerFlags() *pflag.FlagSet {
	flagset := &pflag.FlagSet{}

	flagset.StringVar(&workerOpts.WorkerProfile, "profile", "default", "worker profile to use on the node")
	flagset.StringVar(&workerOpts.APIServer, "api-server", "", "HACK: api-server for the windows worker node")
	flagset.StringVar(&workerOpts.CIDRRange, "cidr-range", "10.96.0.0/12", "HACK: cidr range for the windows worker node")
	flagset.StringVar(&workerOpts.ClusterDNS, "cluster-dns", "10.96.0.10", "HACK: cluster dns for the windows worker node")
	flagset.BoolVar(&workerOpts.CloudProvider, "enable-cloud-provider", false, "Whether or not to enable cloud provider support in kubelet")
	flagset.StringVar(&workerOpts.TokenFile, "token-file", "", "Path to the file containing token.")
	flagset.StringToStringVarP(&workerOpts.CmdLogLevels, "logging", "l", DefaultLogLevels(), "Logging Levels for the different components")
	flagset.StringSliceVarP(&workerOpts.Labels, "labels", "", []string{}, "Node labels, list of key=value pairs")
	flagset.StringSliceVarP(&workerOpts.Taints, "taints", "", []string{}, "Node taints, list of key=value:effect strings")
	flagset.StringVar(&workerOpts.KubeletExtraArgs, "kubelet-extra-args", "", "extra args for kubelet")
	flagset.AddFlagSet(GetCriSocketFlag())

	return flagset
}

func AvailableComponents() []string {
	return []string{
		constant.KonnectivityServerComponentName,
		constant.KubeSchedulerComponentName,
		constant.KubeControllerManagerComponentName,
		constant.ControlAPIComponentName,
		constant.CsrApproverComponentName,
		constant.DefaultPspComponentName,
		constant.KubeProxyComponentName,
		constant.CoreDNSComponentname,
		constant.NetworkProviderComponentName,
		constant.HelmComponentName,
		constant.MetricsServerComponentName,
		constant.KubeletConfigComponentName,
		constant.SystemRbacComponentName,
	}
}

func GetControllerFlags() *pflag.FlagSet {
	flagset := &pflag.FlagSet{}

	flagset.StringVar(&workerOpts.WorkerProfile, "profile", "default", "worker profile to use on the node")
	flagset.BoolVar(&controllerOpts.EnableWorker, "enable-worker", false, "enable worker (default false)")
	flagset.StringSliceVar(&controllerOpts.DisableComponents, "disable-components", []string{}, "disable components (valid items: "+strings.Join(AvailableComponents()[:], ",")+")")
	flagset.StringVar(&workerOpts.TokenFile, "token-file", "", "Path to the file containing join-token.")
	flagset.StringToStringVarP(&workerOpts.CmdLogLevels, "logging", "l", DefaultLogLevels(), "Logging Levels for the different components")
	flagset.BoolVar(&controllerOpts.SingleNode, "single", false, "enable single node (implies --enable-worker, default false)")
	flagset.BoolVar(&controllerOpts.NoTaints, "no-taints", false, "disable default taints for controller node")
	flagset.BoolVar(&controllerOpts.EnableEkeCloudProvider, "enable-eke-cloud-provider", false, "enables the eke-cloud-provider (default false)")
	flagset.DurationVar(&controllerOpts.EkeCloudProviderUpdateFrequency, "eke-cloud-provider-update-frequency", 2*time.Minute, "the frequency of eke-cloud-provider node updates")
	flagset.IntVar(&controllerOpts.EkeCloudProviderPort, "eke-cloud-provider-port", cloudprovider.CloudControllerManagerPort, "the port that eke-cloud-provider binds on")
	flagset.AddFlagSet(GetCriSocketFlag())
	flagset.BoolVar(&controllerOpts.EnableDynamicConfig, "enable-dynamic-config", false, "enable cluster-wide dynamic config based on custom resource")
	flagset.BoolVar(&controllerOpts.EnableMetricsScraper, "enable-metrics-scraper", false, "enable scraping metrics from the controller components (kube-scheduler, kube-controller-manager)")
	flagset.AddFlagSet(FileInputFlag())
	return flagset
}

// The config flag used to be a persistent, joint flag to all commands
// now only a few commands use it. This function helps to share the flag with multiple commands without needing to define
// it in multiple places
func FileInputFlag() *pflag.FlagSet {
	flagset := &pflag.FlagSet{}
	descString := fmt.Sprintf("config file, use '-' to read the config from stdin (default \"%s\")", constant.EkeConfigPathDefault)
	flagset.StringVarP(&CfgFile, "config", "c", "", descString)

	return flagset
}

func GetCmdOpts() CLIOptions {
	EkeVars = constant.GetConfig(DataDir)

	if controllerOpts.SingleNode {
		controllerOpts.EnableWorker = true
		EkeVars.DefaultStorageType = "kine"
	}

	// When CfgFile is set, verify the file can be opened
	if CfgFile != "" {
		_, err := os.Open(CfgFile)
		if err != nil {
			logrus.Fatalf("failed to load config file (%s): %v", CfgFile, err)
		}
	}

	opts := CLIOptions{
		ControllerOptions: controllerOpts,
		WorkerOptions:     workerOpts,

		CfgFile:          CfgFile,
		ClusterConfig:    getClusterConfig(EkeVars),
		NodeConfig:       getNodeConfig(EkeVars),
		CmdConfig:        getEkeCmdConfig(),
		Debug:            Debug,
		Verbose:          Verbose,
		DefaultLogLevels: DefaultLogLevels(),
		EkeVars:          EkeVars,
		DebugListenOn:    DebugListenOn,
	}
	return opts
}

func PreRunValidateConfig(ekeVars constant.CfgVars) error {
	loadingRules := ClientConfigLoadingRules{EkeVars: ekeVars}
	_, err := loadingRules.ParseRuntimeConfig()
	if err != nil {
		return fmt.Errorf("failed to get config: %v", err)
	}
	return nil
}
func getNodeConfig(ekeVars constant.CfgVars) *v1beta1.ClusterConfig {
	loadingRules := ClientConfigLoadingRules{Nodeconfig: true, EkeVars: ekeVars}
	cfg, err := loadingRules.Load()
	if err != nil {
		return nil
	}
	return cfg
}

func getClusterConfig(ekeVars constant.CfgVars) *v1beta1.ClusterConfig {
	loadingRules := ClientConfigLoadingRules{EkeVars: ekeVars}
	cfg, err := loadingRules.Load()
	if err != nil {
		return nil
	}
	return cfg
}

func getEkeCmdConfig() *cmdconfig.EkeCmdConfig {
	configLoader := cmdconfig.NewConfigLoader()

	if CmdCfgFile != "" {
		configLoader.Paths = append(configLoader.Paths, CmdCfgFile)
	}

	cfg, err := configLoader.Load()
	if err != nil {
		return nil
	}
	return cfg
}
