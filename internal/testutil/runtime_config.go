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
package testutil

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"eke/internal/pkg/file"
	"eke/pkg/apis/eke/clientset/fake"

	ekev1beta1 "eke/pkg/apis/eke/clientset/typed/eke/v1beta1"
	"eke/pkg/apis/eke/v1beta1"
	"eke/pkg/config"
	"eke/pkg/constant"
)

const RuntimeFakePath = "/tmp/eke.yaml"

var resourceType = v1.TypeMeta{APIVersion: "eke/v1beta1", Kind: "clusterconfigs"}

type ConfigGetter struct {
	NodeConfig bool
	YamlData   string

	ekeVars     constant.CfgVars
	cfgFilePath string
}

// NewConfigGetter sets the parameters required to fetch a fake config for testing
func NewConfigGetter(yamlData string, isNodeConfig bool, ekeVars constant.CfgVars) *ConfigGetter {
	return &ConfigGetter{
		YamlData:   yamlData,
		NodeConfig: isNodeConfig,
		ekeVars:    ekeVars,
	}
}

// FakeRuntimeConfig takes a yaml construct and returns a config object from a fake runtime config path
func (c *ConfigGetter) FakeConfigFromFile() (*v1beta1.ClusterConfig, error) {
	err := c.initRuntimeConfig()
	if err != nil {
		return nil, err
	}
	loadingRules := config.ClientConfigLoadingRules{
		RuntimeConfigPath: RuntimeFakePath,
		Nodeconfig:        c.NodeConfig,
		EkeVars:           c.ekeVars,
	}
	return loadingRules.Load()
}

func (c *ConfigGetter) FakeAPIConfig() (*v1beta1.ClusterConfig, error) {
	err := c.initRuntimeConfig()
	if err != nil {
		return nil, err
	}

	// create the API config using a fake client
	client := fake.NewSimpleClientset()

	err = c.createFakeAPIConfig(client.EkeV1beta1())
	if err != nil {
		return nil, fmt.Errorf("failed to get fake API client: %v", err)
	}

	loadingRules := config.ClientConfigLoadingRules{
		RuntimeConfigPath: RuntimeFakePath,
		Nodeconfig:        c.NodeConfig,
		APIClient:         client.EkeV1beta1(),
		EkeVars:           c.ekeVars,
	}

	return loadingRules.Load()
}

func (c *ConfigGetter) initRuntimeConfig() error {
	// write the yaml string into a temporary config file path
	cfgFilePath, err := file.WriteTmpFile(c.YamlData, "eke-config")
	if err != nil {
		return fmt.Errorf("error creating tempfile: %v", err)
	}

	c.cfgFilePath = cfgFilePath

	logrus.Infof("using config path: %s", cfgFilePath)
	f, err := os.Open(c.cfgFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	mergedConfig, err := v1beta1.ConfigFromReader(f, c.getStorageSpec())
	if err != nil {
		return fmt.Errorf("unable to parse config from %s: %v", cfgFilePath, err)
	}
	data, err := yaml.Marshal(&mergedConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}
	err = os.WriteFile(RuntimeFakePath, data, 0755)
	if err != nil {
		return fmt.Errorf("failed to write runtime config to %s: %v", RuntimeFakePath, err)
	}
	return nil
}

func (c *ConfigGetter) createFakeAPIConfig(client ekev1beta1.EkeV1beta1Interface) error {
	clusterConfigs := client.ClusterConfigs(constant.ClusterConfigNamespace)
	ctxWithTimeout, cancelFunction := context.WithTimeout(context.Background(), time.Duration(10)*time.Second)
	defer cancelFunction()

	cfg, err := v1beta1.ConfigFromString(c.YamlData, c.getStorageSpec())
	if err != nil {
		return fmt.Errorf("failed to parse config yaml: %s", err.Error())
	}

	_, err = clusterConfigs.Create(ctxWithTimeout, cfg.GetClusterWideConfig().StripDefaults(), v1.CreateOptions{TypeMeta: resourceType})
	if err != nil {
		return fmt.Errorf("failed to create clusterConfig in the API: %s", err.Error())
	}
	return nil
}

func (c *ConfigGetter) getStorageSpec() *v1beta1.StorageSpec {
	var storage *v1beta1.StorageSpec

	if c.ekeVars.DefaultStorageType == "kine" {
		storage = &v1beta1.StorageSpec{
			Type: v1beta1.KineStorageType,
			Kine: v1beta1.DefaultKineConfig(c.ekeVars.DataDir),
		}
	}
	return storage
}
