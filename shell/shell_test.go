package shell

import (
	"testing"
	"log"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.ibm.com/composer/cloud-shell-cli/i18n"
	"github.com/IBM-Cloud/ibm-cloud-cli-sdk/plugin/pluginfakes"
	"github.com/IBM-Cloud/ibm-cloud-cli-sdk/testhelpers/terminal"
	"github.com/stretchr/testify/suite"
	"github.com/IBM-Cloud/ibm-cloud-cli-sdk/common/file_helpers"
)

type ShellCmdTestSuite struct {
	suite.Suite
	ui            *terminal.FakeUI
	pluginContext *pluginfakes.FakePluginContext
	cmd           *CloudShellPlugin
	SaveDir       string
	version       string
}

func TestShellCmdTestSuite(t *testing.T) {
	suite.Run(t, new(ShellCmdTestSuite))
}

func (suite *ShellCmdTestSuite) SetupSuit() {
	i18n.T = i18n.InitWithLocale(i18n.DEFAULT_LOCALE)
}


func (suite *ShellCmdTestSuite) SetupSuite() {
	suite.SaveDir, _ = ioutil.TempDir("", "testfiledownload")
	suite.pluginContext = createDefaultFakePluginContext(suite.SaveDir)
	suite.version = suite.cmd.GetMetadata().Version.String()
	suite.ui = terminal.NewFakeUI()
	suite.cmd =  &CloudShellPlugin{
		ui: suite.ui,
	}
	suite.cmd.init(suite.ui)
}

func (suite *ShellCmdTestSuite) TearDownSuite() {
	if suite.SaveDir != "" {
		os.RemoveAll(suite.SaveDir)
	}
}

func (suite *ShellCmdTestSuite) TestGetIsCmdHeadless() {
	isHeadless := IsCommandHeadless([]string{"shell"})
	suite.Equal(false, isHeadless)
}


func (suite *ShellCmdTestSuite) TestRunDownloadDistNonHeadless() {
	suite.cmd.DownloadDistIfNecessary(suite.pluginContext, false)
	log.Println(suite.ui.Outputs())
	path := suite.pluginContext.PluginDirectory()
	successFile := filepath.Join(path, "/cache-" + suite.version, "success")
	suite.FileExists(successFile)

	// Test duplicate download fails
	os.Remove(filepath.Join(path, "/cache-" + suite.version, "success"))
	_, err := suite.cmd.DownloadDistIfNecessary(suite.pluginContext, false)
	suite.Equal(file_helpers.FileExists(successFile), false)
	suite.NotNil(err)
}


func (suite *ShellCmdTestSuite) TestRunDownloadDistHeadless() {
	if MinimalNodeVersionSupported() {
		suite.cmd.DownloadDistIfNecessary(suite.pluginContext, true)
		log.Println(suite.ui.Outputs())
		path := suite.pluginContext.PluginDirectory()
		successFile := filepath.Join(path, "/cache-headless" + suite.version, "success")
		suite.FileExists(successFile)

		// Test duplicate download
		os.Remove(filepath.Join(path, "/cache-headless" + suite.version, "success"))
		_, err := suite.cmd.DownloadDistIfNecessary(suite.pluginContext, true)
		suite.Equal(file_helpers.FileExists(successFile), false)
		suite.NotNil(err)
	}
}

func createDefaultFakePluginContext(saveDir string) *pluginfakes.FakePluginContext {
	context := new(pluginfakes.FakePluginContext)
	cfContext := new(pluginfakes.FakeCFContext)
	context.PluginConfigReturns(new(pluginfakes.FakePluginConfig))
	context.HasAPIEndpointReturns(true)
	context.CFReturns(cfContext)
	cfContext.IsLoggedInReturns(true)
	//context.APIEndpointReturns("https://api-endpoint")
	context.PluginDirectoryReturns(saveDir)
	context.TraceReturns("true")
	log.Println(context.PluginDirectory())
	return context
}
