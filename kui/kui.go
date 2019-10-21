package kui

import (
	"fmt"
	"github.com/mholt/archiver"
	. "github.com/kui-shell/kask/i18n"
	log "go.uber.org/zap"
	baselog "log"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

type KrewComponent interface {
	init()
}

type KuiComponent struct {
}

type Context interface {
	PluginDirectory() (string, error)
	logger() *log.SugaredLogger
}

// THE PLUGIN_VERSION CONSTANT SHOULD BE LEFT EXACTLY AS-IS SINCE IT CAN BE PROGRAMMATICALLY SUBSTITUTED
const PLUGIN_VERSION = "dev"

// the command context to use for Kui command execution; this will be
// the default, if we cannot infer one from e.g. the executable name
// (whereby "kubectl-foo" implies a command context of "foo")
const defaultCommandContext = "plugin"

type MainContext struct {
	_logger *log.SugaredLogger
}
func (context MainContext) PluginDirectory() (string, error) {
	home, err := os.UserHomeDir()
	if err == nil {
		return filepath.Join(home, ".kask"), nil
	} else {
		handleError(context, err)
		return "", err
	}
}
func (context MainContext) logger() *log.SugaredLogger {
	return context._logger
}
func initLogger()(*log.Logger, error) {
	_, debugSet := os.LookupEnv("DEBUG")
	if debugSet {
		return log.NewDevelopment()
	} else {
		return log.NewProduction()
	}
}
func (context *MainContext) initDefault()(*MainContext) {
	logger, err := initLogger()
	if err != nil {
		baselog.Fatalf("can't initialize zap logger: %v", err)
	}
	context._logger = logger.Sugar()
	return context
}

func Start() {
	runner := KuiComponent{}
	context := MainContext{}
	context.initDefault()
	runner.Run(context, os.Args)
}

func (component *KuiComponent) init() {
}

func blue(str string) string {
	return fmt.Sprintf("\033[0;34m%s\033[0m", str)
}
func cyan(str string) string {
	return fmt.Sprintf("\033[0;36m%s\033[0m", str)
}
func yellow(str string) string {
	return fmt.Sprintf("\033[0;33m%s\033[0m", str)
}

func (component *KuiComponent) Run(context Context, args []string) {
	component.init()

	if len(args) == 1 {
		fmt.Printf("Usage: %v\n\n%v\n%v\tUpdate the local UI code\n%v\tPrint the version of the UI code", cyan("kask <command>"), yellow("Commands:"), blue("refresh"), blue("version"))
		os.Exit(1)
		return
	}

	refreshRequested := args[1] == "refresh"
	context.logger().Debugf("refreshRequested? %v", refreshRequested)

	cmd, err := component.DownloadDistIfNecessary(context, refreshRequested)
	if err != nil {
		os.Exit(1)
		return
	}

	var kaskArgs []string
	if refreshRequested {
		context.logger().Debug("refresh done")
		kaskArgs = []string{"version"}
	} else {
		kaskArgs = args[1:]
	}

	base := path.Base(args[0])
	r := regexp.MustCompile("^kubectl-")
	var kuiCommandContext string
	if r.MatchString(base) {
		kuiCommandContext = r.ReplaceAllString(base, "")
	} else {
		kuiCommandContext = defaultCommandContext
	}
	if kuiCommandContext == "kask" {
		kuiCommandContext = defaultCommandContext
	}
	context.logger().Debugf("command context: %s", kuiCommandContext)
	cmd.Env = append(cmd.Env, "KUI_COMMAND_CONTEXT=" + kuiCommandContext)

	component.invokeRun(context, cmd, kaskArgs)
}

func (component *KuiComponent) invokeRun(context Context, cmd *exec.Cmd, kaskArgs []string) {
	cmd.Args = append(cmd.Args, kaskArgs...)
	context.logger().Debugf("args %s", cmd.Args)

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err := cmd.Start(); err != nil {
		fmt.Println("command failed!")
	}
}

func GetDistOSSuffix() string {
	switch runtime.GOOS {
	case "windows":
		return "-base-win32-x64.zip"
	case "darwin":
		return "-base-darwin-x64.tar.bz2"
	default:
		return "-base-linux-x64.zip"
	}
}

func GetRootCommand(extractedDir string) *exec.Cmd {
	switch runtime.GOOS {
	case "windows":
		// TODO verify
		return exec.Command(filepath.Join(extractedDir, "Kui-base-win32-x64\\Kui.exe"))
	case "darwin":
		return exec.Command(filepath.Join(extractedDir, "Kui-base-darwin-x64/Kui.app/Contents/MacOS/Kui"))
	default:
		// TODO verify
		return exec.Command(filepath.Join(extractedDir, "Kui-base-linux-x64/Kui"))
	}
}

func GetDistLocation(version string) string {
	host := "https://s3-api.us-geo.objectstorage.softlayer.net/kui-" + version
	DEV_OVERRIDE_HOST, overrideSet := os.LookupEnv("KUI_DIST")
	if overrideSet {
		host = DEV_OVERRIDE_HOST
	}
	if !strings.HasSuffix(host, "/") {
		host += "/"
	}
	return host + "Kui" + GetDistOSSuffix()
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filepath string, url string) error {

    // Get the data
    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // Create the file
    out, err := os.Create(filepath)
    if err != nil {
        return err
    }
    defer out.Close()

    // Write the body to file
    _, err = io.Copy(out, resp.Body)
    return err
}

func (p *KuiComponent) DownloadDistIfNecessary(context Context, force bool) (*exec.Cmd, error) {
	Debug := context.logger().Debug
	Debugf := context.logger().Debugf

	Debugf("force refetch? %v", force)

	metadata := p.GetMetadata()
	version := metadata.Version.String()

	url := GetDistLocation(version)

	pluginDir, err := context.PluginDirectory()
	if err != nil {
		handleError(context, err)
		return nil, err
	}

	binDir := filepath.Join(pluginDir, "bin")
	targetDir := filepath.Join(pluginDir, "/cache-"+version)
	successFile := filepath.Join(targetDir, "success")
	extractedDir := filepath.Join(targetDir, "extract")
	Debugf("targetDir %s", targetDir)

	executable, err := os.Executable()
	if err != nil {
		handleError(context, err)
		return nil, err
	}
	basenameOfSelf := filepath.Base(executable)

	command := GetRootCommand(extractedDir)
	command.Env = append(os.Environ(), "KUI_BIN_DIR=" + binDir, "KUI_BIN_PREFIX=kubectl-", "KUI_BIN=" + executable, "KUI_DEFAULT_PRETTY_TYPE=" + basenameOfSelf)

	if force {
		err := os.Remove(successFile)
		if err != nil {
			Debugf("error removing lock file %v", err)
		}
		err2 := os.RemoveAll(extractedDir)
		if err2 != nil {
			Debugf("error removing unpack %v", err2)
		}
	}

	if _, err := os.Stat(successFile); err != nil {
		downloadedFile := filepath.Join(targetDir, "downloaded.zip")
		extractedDir := filepath.Join(targetDir, "extract")

		os.MkdirAll(extractedDir, 0700)

		if err := DownloadFile(downloadedFile, url); err != nil {
			handleError(context, err)
			return nil, err
		}

		// link ourselves to kubectl-<basename>
		targetOfSymlink := filepath.Join(binDir, "kubectl-" + basenameOfSelf)
		os.Remove(targetOfSymlink)
		os.Symlink(executable, targetOfSymlink)
		Debugf("Symlinked ourselves to %s", targetOfSymlink)

		Debugf("Downloaded kui-base %s", downloadedFile)
		Debugf("Extracting kui-base %s", extractedDir)

		if strings.HasSuffix(url, ".tar.bz2") {
			if err := archiver.DefaultTarBz2.Unarchive(downloadedFile, extractedDir); err != nil {
				handleError(context, err)
				return nil, err
			}
		} else {
			if err := archiver.DefaultZip.Unarchive(downloadedFile, extractedDir); err != nil {
				handleError(context, err)
				return nil, err
			}
		}

		Debugf("Extracted kui-base %s", extractedDir)

		if _, err := os.OpenFile(successFile, os.O_RDONLY|os.O_CREATE, 0666); err != nil {
			handleError(context, err)
			return nil, err
		}
	} else {
		Debug("Using cached download")
	}

	return command, nil
}

func MakeExecutable(path string) error {
	return filepath.Walk(path, func(name string, info os.FileInfo, err error) error {
		if err == nil {
			err = os.Chmod(name, 0700)
		}
		return err
	})
}

type Command struct {
	Name string
	Alias string
	Description string
	Usage string
}
type Metadata struct {
	Name string
	Version VersionType
	Commands []Command
}

func (component *KuiComponent) GetMetadata() Metadata {
	return Metadata{
		Name:    "kask",
		Version: GetVersion(),
		Commands: []Command{
			{
				Name:        "kask",
				Alias:       "kask",
				Description: "Kask for Krew",
				Usage:       "krew kask install ...",
			},
		},
	}
}

func handleError(context Context, err error) {
	switch err {
	case nil:
		return
	default:
		context.logger().Errorw("msg", T("An error has occurred:\n{{.Error}}\n", map[string]interface{}{"Error": err.Error()}))
	}

	return
}

func toInt(in string) int {
	outValue, _ := strconv.Atoi(in)
	return outValue
}

type VersionType interface {
	String() string
}

type DevVer struct {}

type SemVer struct {
	Major int
	Minor int
	Build int
}

func (version SemVer) String() string {
	return fmt.Sprintf("%d.%d.%d", version.Major, version.Minor, version.Build)
}

func (DevVer) String() string {
	return PLUGIN_VERSION
}

func GetVersion() VersionType {
	if PLUGIN_VERSION == "dev" {
		return DevVer{}
	} else {
		s := strings.Split(PLUGIN_VERSION, ".")
		return SemVer{
			Major: toInt(s[0]),
			Minor: toInt(s[1]),
			Build: toInt(s[2]),
		}
	}
}
