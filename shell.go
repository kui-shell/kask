package main

import (
	"fmt"
	"github.com/IBM-Cloud/ibm-cloud-cli-sdk/bluemix/terminal"
	"github.com/IBM-Cloud/ibm-cloud-cli-sdk/bluemix/trace"
	"github.com/IBM-Cloud/ibm-cloud-cli-sdk/common/downloader"
	"github.com/IBM-Cloud/ibm-cloud-cli-sdk/common/file_helpers"
	"github.com/IBM-Cloud/ibm-cloud-cli-sdk/plugin"
	"github.com/mholt/archiver"
	. "github.ibm.com/composer/cloud-shell-cli/i18n"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

type CloudShellPlugin struct{}

// THE PLUGIN_VERSION CONSTANT SHOULD BE LEFT EXACTLY AS-IS SINCE IT CAN BE PROGRAMMATICALLY SUBSTITUTED
const PLUGIN_VERSION = "1.6.1"

const MINIMAL_NODE_VERSION = 8

func main() {
	argsWithoutProg := os.Args[1:]
	if len(argsWithoutProg) > 0 && argsWithoutProg[0] == "version" {
		version := GetVersion()
		fmt.Println(version.String())
		return
	}
	plugin.Start(new(CloudShellPlugin))
}

func (shellPlugin *CloudShellPlugin) Run(context plugin.PluginContext, args []string) {
	trace.Logger = trace.NewLogger(context.Trace())

	shellArgs := args[1:]
	headless := IsCommandHeadless(shellArgs)

	if headless {
		trace.Logger.Println("Executing headless command")
	}

	cmd := shellPlugin.DownloadDistIfNecessary(context, headless)
	cmd.Args = append(cmd.Args, shellArgs...)

	trace.Logger.Println(cmd)

	if !headless {
		if err := cmd.Start(); err != nil {
			fmt.Println("command failed!")
		}
	} else {
		stdoutStderr, err := cmd.CombinedOutput()
		cmd.Run()
		if err != nil {
			fmt.Println("headless command failed!")
			fmt.Println(err)
		}
		fmt.Printf("%s", stdoutStderr)
	}
}

func GetDistOSSuffix(headless bool) string {
	if headless {
		return "-headless.zip"
	}
	switch runtime.GOOS {
	case "windows":
		return "-win32-x64.zip"
	case "darwin":
		return "-darwin-x64.tar.bz2"
	default:
		return "-linux-x64.zip"
	}
}

func GetRootCommand(extractedDir string, headless bool) *exec.Cmd {
	if headless {
		return exec.Command("node", filepath.Join(extractedDir, "cloudshell/bin/fsh.js"))
	}
	switch runtime.GOOS {
	case "windows":
		// TODO verify
		return exec.Command(filepath.Join(extractedDir, "IBM Cloud Shell-win32-x64\\IBM Cloud Shell.exe"))
	case "darwin":
		return exec.Command(filepath.Join(extractedDir, "IBM Cloud Shell-darwin-x64/IBM Cloud Shell.app/Contents/MacOS/IBM Cloud Shell"))
	default:
		// TODO verify
		return exec.Command(filepath.Join(extractedDir, "IBM Cloud Shell-linux-x64/IBM Cloud Shell"))
	}
}

func GetDistLocation(version string, headless bool) string {
	/*
		production distributions:
			https://s3-api.us-geo.objectstorage.softlayer.net/ibm-cloud-shell-v1.6.1/IBM%20Cloud%20Shell-darwin-x64.zip
			https://s3-api.us-geo.objectstorage.softlayer.net/ibm-cloud-shell-v1.6.1/IBM%20Cloud%20Shell-win32-x64.zip
			https://s3-api.us-geo.objectstorage.softlayer.net/ibm-cloud-shell-v1.6.1/IBM%20Cloud%20Shell-linux-x64.zip

		dev distributions:
			win32: https://s3-api.us-geo.objectstorage.softlayer.net/ibm-cloud-shell-dev/IBM%20Cloud%20Shell-win32-x64.zip
			macOS zip: https://s3-api.us-geo.objectstorage.softlayer.net/ibm-cloud-shell-dev/IBM%20Cloud%20Shell-darwin-x64.zip
			macOS tar.bz2: https://s3-api.us-geo.objectstorage.softlayer.net/ibm-cloud-shell-dev/IBM%20Cloud%20Shell-darwin-x64.tar.bz2
			linux-zip: https://s3-api.us-geo.objectstorage.softlayer.net/ibm-cloud-shell-dev/IBM%20Cloud%20Shell-linux-x64.zip
			headless: https://s3-api.us-geo.objectstorage.softlayer.net/ibm-cloud-shell-dev/IBM%20Cloud%20Shell-headless.zip
	*/

	host := "https://s3-api.us-geo.objectstorage.softlayer.net/ibm-cloud-shell-v" + version
	DEV_OVERRIDE_HOST, overrideSet := os.LookupEnv("CLOUDSHELL_DIST")
	if overrideSet {
		host = DEV_OVERRIDE_HOST
	}
	if !strings.HasSuffix(host, "/") {
		host += "/"
	}
	return host + "IBM%20Cloud%20Shell" + GetDistOSSuffix(headless)
}

func IsCommandHeadless(shellArgs []string) bool {
	isShell := len(shellArgs) > 0 && (shellArgs[0] == "shell" || shellArgs[0] == "preview")
	return !isShell
}

func MinimalNodeVersionSupported() bool {
	cmd := exec.Command("node", "-v")
	stdout, err := cmd.CombinedOutput()
	cmd.Run()
	if err != nil {
		return false
	}
	versionRegEx := regexp.MustCompile(`^v([0-9]*)`)
	version := string(stdout[:])
	trace.Logger.Println("Node version is " + version)
	result := versionRegEx.FindStringSubmatch(version)
	return len(result) > 1 && ToInt(result[1]) >= MINIMAL_NODE_VERSION
}

func (shellPlugin *CloudShellPlugin) DownloadDistIfNecessary(context plugin.PluginContext, headless bool) *exec.Cmd {
	ui := terminal.NewStdUI()
	metadata := shellPlugin.GetMetadata()
	version := metadata.Version.String()

	// we can only support headless execution using the nodejs that's installed on the user's machine
	if headless && !MinimalNodeVersionSupported() {
		trace.Logger.Println("Can't use headless since minimal node version is not supported")
		headless = false
		// TODO get full shell working with headless commands - Nick arg position
	}

	url := GetDistLocation(version, headless)

	targetDir := filepath.Join(context.PluginDirectory(), "/cache-"+version)

	if headless {
		targetDir = filepath.Join(context.PluginDirectory(), "/cache-headless-"+version)
	}

	successFile := filepath.Join(targetDir, "success")
	extractedDir := filepath.Join(targetDir, "extract")
	command := GetRootCommand(extractedDir, headless)
	if !file_helpers.FileExists(successFile) {
		downloadedFile := filepath.Join(targetDir, "downloaded.zip")
		extractedDir := filepath.Join(targetDir, "extract")

		os.MkdirAll(extractedDir, 0700)

		fileDownloader := new(downloader.FileDownloader)
		// we don't want headless mode to include anything extra in the output
		if !headless {
			fileDownloader.ProxyReader = downloader.NewProgressBar(ui.Writer())
		}
		trace.Logger.Println("Downloading shell from " + url + " to " + downloadedFile)
		if _, _, err := fileDownloader.DownloadTo(url, downloadedFile); err != nil {
			handleError(err, ui)
		}
		trace.Logger.Println("Downloaded shell to " + downloadedFile)

		trace.Logger.Println("Extracting shell to " + extractedDir)
		if strings.HasSuffix(url, ".tar.bz2") {
			if err := archiver.TarBz2.Open(downloadedFile, extractedDir); err != nil {
				handleError(err, ui)
			}
		} else {
			if err := archiver.Zip.Open(downloadedFile, extractedDir); err != nil {
				handleError(err, ui)
			}
		}

		trace.Logger.Println("Extracted shell to " + extractedDir)

		if _, err := os.OpenFile(successFile, os.O_RDONLY|os.O_CREATE, 0666); err != nil {
			handleError(err, ui)
		}
	} else {
		trace.Logger.Println("Using cached download")
	}

	return command
}

func MakeExecutable(path string) error {
	return filepath.Walk(path, func(name string, info os.FileInfo, err error) error {
		if err == nil {
			err = os.Chmod(name, 0700)
		}
		return err
	})
}

func (shellPlugin *CloudShellPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name:    "function-composer",
		Version: GetVersion(),
		Commands: []plugin.Command{
			{
				Name:        "shell",
				Alias:       "fsh",
				Description: "Function composer",
				Usage:       "ibmcloud shell",
			},
		},
	}
}

func handleError(err error, ui terminal.UI) string {
	switch err {
	case nil:
		return ""
	default:
		ui.Failed(T("An error has occurred:\n{{.Error}}\n", map[string]interface{}{"Error": err.Error()}))
		os.Exit(1)
	}

	return ""
}

func ToInt(in string) int {
	outValue, _ := strconv.Atoi(in)
	return outValue
}

func GetVersion() plugin.VersionType {
	s := strings.Split(PLUGIN_VERSION, ".")
	return plugin.VersionType{
		Major: ToInt(s[0]),
		Minor: ToInt(s[1]),
		Build: ToInt(s[2]),
	}
}
