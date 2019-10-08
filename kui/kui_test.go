package kui

import (
	"testing"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/kui-shell/kask/i18n"
	"github.com/stretchr/testify/suite"
)

type KaskTestSuite struct {
	suite.Suite
	pluginContext *MainContext
	cmd           *KuiComponent
	SaveDir       string
	version       string
}

func TestKaskTestSuite(t *testing.T) {
	suite.Run(t, new(KaskTestSuite))
}

func (suite *KaskTestSuite) SetupSuite() {
	i18n.T = i18n.InitWithLocale(i18n.DEFAULT_LOCALE)
	suite.SaveDir, _ = ioutil.TempDir("", "testfiledownload")
	suite.pluginContext = createDefaultFakePluginContext(suite.SaveDir)
	suite.version = suite.cmd.GetMetadata().Version.String()
	suite.cmd = &KuiComponent{
	}
	suite.cmd.init()
}

func (suite *KaskTestSuite) TearDownSuite() {
	if suite.SaveDir != "" {
		os.RemoveAll(suite.SaveDir)
	}
}

func (suite *KaskTestSuite) TestRunDownload() {
	kui, err := suite.cmd.DownloadDistIfNecessary(suite.pluginContext, false)
	path, err := suite.pluginContext.PluginDirectory()
	suite.Nil(err)
	successFile := filepath.Join(path, "/cache-" + suite.version, "success")
	suite.FileExists(successFile)

	// Run kui command
	kui.Args = append(kui.Args, "kui version")
	kui.Env = append(os.Environ(), "KUI_TEE_TO_FILE=/tmp/kask-kui-test", "KUI_TEE_TO_FILE_EXIT_ON_END_MARKER=true")
	stdout, err := kui.CombinedOutput()
	kui.Run()
	version := string(stdout[:])
	suite.NotNil(version)
}

func createDefaultFakePluginContext(saveDir string) *MainContext {
	return new(MainContext).initDefault()
}
