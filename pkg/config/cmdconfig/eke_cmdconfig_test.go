package cmdconfig

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

type testData struct {
	FakeUsrEtc string
	FakeEtc    string
	FakeHome   string
}

func setup() (td testData, err error) {
	td.FakeUsrEtc, err = ioutil.TempDir("", "fake-usr-etc")
	if err != nil {
		return
	}

	td.FakeEtc, err = ioutil.TempDir("", "fake-etc")
	if err != nil {
		return
	}

	td.FakeHome, err = ioutil.TempDir("", "fake-home")
	if err != nil {
		return
	}

	return
}

func teardown(td testData) {
	os.RemoveAll(td.FakeUsrEtc)
	os.RemoveAll(td.FakeEtc)
	os.RemoveAll(td.FakeHome)
}

func writeConfig(path, data string) error {
	return ioutil.WriteFile(
		filepath.Join(path, "eke.cmd.yaml"),
		[]byte(data),
		0644)
}

func TestOnlySystemConfigExists(t *testing.T) {

	td, err := setup()
	if err != nil {
		t.Error(err)
	}
	defer teardown(td)

	var data = `
ekeKubectlConfig:
  allowDownload: false
  systemPath: test
  timeout: 5
`
	err = writeConfig(td.FakeUsrEtc, data)
	if err != nil {
		t.Error(err)
	}

	c := ConfigLoader{
		Paths: []string{td.FakeUsrEtc, td.FakeEtc, td.FakeHome},
	}

	v, err := c.Load()
	if err != nil {
		t.Errorf("Unexpected error loading config: %v", err)
	}
	if v.EkeKubectlConfig.AllowDownload != false {
		t.Error("Expected configuration value wasn't found")
	}
}

func TestHomeConfigOverridesSystemOne(t *testing.T) {
	td, err := setup()
	if err != nil {
		t.Error(err)
	}
	defer teardown(td)

	var data = `
ekeKubectlConfig:
  allowDownload: false
`

	err = writeConfig(td.FakeUsrEtc, data)
	if err != nil {
		t.Error(err)
	}

	var data2 = `
ekeKubectlConfig:
  allowDownload: true
`
	err = writeConfig(td.FakeHome, data2)
	if err != nil {
		t.Error(err)
	}

	c := ConfigLoader{
		Paths: []string{td.FakeUsrEtc, td.FakeEtc, td.FakeHome},
	}

	v, err := c.Load()
	if err != nil {
		t.Errorf("Unexpected error loading config: %v", err)
	}
	if v.EkeKubectlConfig.AllowDownload != true {
		t.Error("Expected configuration value wasn't found")
	}
}

func TestMergeConfigs(t *testing.T) {
	td, err := setup()
	if err != nil {
		t.Error(err)
	}
	defer teardown(td)

	usrEtcCfg := `
ekeKubectlConfig:
  allowDownload: true
  systemPath: test
  timeout: 5
`
	err = writeConfig(td.FakeUsrEtc, usrEtcCfg)
	if err != nil {
		t.Error(err)
	}

	etcCfg := `
ekeKubectlConfig:
  timeout: 200
`
	err = writeConfig(td.FakeEtc, etcCfg)
	if err != nil {
		t.Error(err)
	}

	homeCfg := `
ekeKubectlConfig:
  systemPath: global
`
	err = writeConfig(td.FakeHome, homeCfg)
	if err != nil {
		t.Error(err)
	}

	c := ConfigLoader{
		Paths: []string{td.FakeUsrEtc, td.FakeEtc, td.FakeHome},
	}

	v, err := c.Load()
	if err != nil {
		t.Errorf("Unexpected error loading config: %v", err)
	}

	if v.EkeKubectlConfig.AllowDownload != true {
		t.Errorf(
			"Wrong value for AllowDownload: got %v instead of %v",
			v.EkeKubectlConfig.AllowDownload, true)
	}

	if v.EkeKubectlConfig.Timeout != 200 {
		t.Errorf(
			"Wrong value for Timeout: got %v instead of %v",
			v.EkeKubectlConfig.Timeout, 200)
	}

	if v.EkeKubectlConfig.SystemPath != "global" {
		t.Errorf(
			"Wrong value for Timeout: got %v instead of %v",
			v.EkeKubectlConfig.SystemPath, "global")
	}
}
