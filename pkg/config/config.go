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

	"eke/internal/pkg/file"

	"eke/pkg/apis/eke/v1beta1"
	"eke/pkg/constant"

	cfgClient "eke/pkg/apis/eke/clientset"

	ekev1beta1 "eke/pkg/apis/eke/clientset/typed/eke/v1beta1"

	"k8s.io/client-go/tools/clientcmd"
)

// general interface for config related methods
type Loader interface {
	BootstrapConfig() (*v1beta1.ClusterConfig, error)
	ClusterConfig() (*v1beta1.ClusterConfig, error)
	IsAPIConfig() bool
	IsDefaultConfig() bool
	Load() (*v1beta1.ClusterConfig, error)
}

type EkeConfigGetter struct {
	ekeConfigGetter Getter
}

func (g *EkeConfigGetter) IsAPIConfig() bool {
	return false
}

func (g *EkeConfigGetter) IsDefaultConfig() bool {
	return false
}

func (g *EkeConfigGetter) BootstrapConfig() (*v1beta1.ClusterConfig, error) {
	return g.ekeConfigGetter()
}

func (g *EkeConfigGetter) Load() (*v1beta1.ClusterConfig, error) {
	return g.ekeConfigGetter()
}

type Getter func() (*v1beta1.ClusterConfig, error)

var _ Loader = &ClientConfigLoadingRules{}

type ClientConfigLoadingRules struct {
	// APIClient is an optional field for passing a kubernetes API client, to fetch the API config
	// mostly used by tests, to pass a fake client
	APIClient ekev1beta1.EkeV1beta1Interface

	// Nodeconfig is an optional field indicating if provided config-file is a node-config or a standard cluster-config file.
	Nodeconfig bool

	// RuntimeConfigPath is an optional field indicating the location of the runtime config file (default: /run/eke/eke.yaml)
	// this parameter is mainly used for testing purposes, to override the default location on local dev system
	RuntimeConfigPath string

	// EkeVars is needed for fetching the right config from the API
	EkeVars constant.CfgVars
}

func (rules *ClientConfigLoadingRules) BootstrapConfig() (*v1beta1.ClusterConfig, error) {
	return rules.fetchNodeConfig()
}

// ClusterConfig generates a client and queries the API for the cluster config
func (rules *ClientConfigLoadingRules) ClusterConfig() (*v1beta1.ClusterConfig, error) {
	if rules.APIClient == nil {
		// generate a kubernetes client from AdminKubeConfigPath
		config, err := clientcmd.BuildConfigFromFlags("", EkeVars.AdminKubeConfigPath)
		if err != nil {
			return nil, fmt.Errorf("can't read kubeconfig: %v", err)
		}
		client, err := cfgClient.NewForConfig(config)
		if err != nil {
			return nil, fmt.Errorf("can't create kubernetes typed client for cluster config: %v", err)
		}

		rules.APIClient = client.EkeV1beta1()
	}
	return rules.getConfigFromAPI(rules.APIClient)
}

func (rules *ClientConfigLoadingRules) IsAPIConfig() bool {
	return controllerOpts.EnableDynamicConfig
}

func (rules *ClientConfigLoadingRules) IsDefaultConfig() bool {
	// if no custom-value is provided as a config file, and no config-file exists in the default location
	// we assume we need to generate configuration defaults
	return CfgFile == constant.EkeConfigPathDefault && !file.Exists(constant.EkeConfigPathDefault)
}

func (rules *ClientConfigLoadingRules) Load() (*v1beta1.ClusterConfig, error) {
	if rules.Nodeconfig {
		return rules.fetchNodeConfig()
	}
	if !rules.IsAPIConfig() {
		return rules.readRuntimeConfig()
	}
	if rules.IsAPIConfig() {
		nodeConfig, err := rules.BootstrapConfig()
		if err != nil {
			return nil, err
		}
		apiConfig, err := rules.ClusterConfig()
		if err != nil {
			return nil, err
		}
		// get node config from the config-file and cluster-wide settings from the API and return a combined result
		return rules.mergeNodeAndClusterconfig(nodeConfig, apiConfig)
	}
	return nil, nil
}
