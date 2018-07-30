package main

import (
	"fmt"
	"github.com/IBM-Cloud/ibm-cloud-cli-sdk/bluemix/terminal"
	"github.com/IBM-Cloud/ibm-cloud-cli-sdk/bluemix/trace"
	"github.com/IBM-Cloud/ibm-cloud-cli-sdk/common/downloader"
	"github.com/IBM-Cloud/ibm-cloud-cli-sdk/common/file_helpers"
	"github.com/IBM-Cloud/ibm-cloud-cli-sdk/plugin"
	"github.com/mholt/archiver"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"strconv"
)

type CloudShellPlugin struct{}

// THE PLUGIN_VERSION CONSTANT SHOULD BE LEFT EXACTLY AS-IS SINCE IT CAN BE PROGRAMMATICALLY SUBSTITUTED
const PLUGIN_VERSION = "1.6.1"

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

	command := shellPlugin.DownloadDistIfNecessary(context)

	shellArgs := args[1:]
	trace.Logger.Println("Running command " + command + " " + strings.Join(shellArgs, " "))

	cmd := exec.Command(command, shellArgs...)

	exitImmediately := len(shellArgs) > 0 && shellArgs[0] == "shell"
	if exitImmediately {
		if err := cmd.Start(); err != nil {
			fmt.Println("command failed!")
		}
	} else {
		stdoutStderr, _ := cmd.CombinedOutput()
		if err := cmd.Start(); err != nil {
		}
		fmt.Printf("%s", stdoutStderr)
		if err := cmd.Wait(); err != nil {
		}
	}
}

func (shellPlugin *CloudShellPlugin) DownloadDistIfNecessary(context plugin.PluginContext) string {
	metadata := shellPlugin.GetMetadata()
	version := metadata.Version.String()

	host := "https://s3-api.us-geo.objectstorage.softlayer.net/shelldist/"
	archivePath := "shell-" + version + "-darwin.tar.gz"

	url := host + archivePath
	targetDir := filepath.Join(context.PluginDirectory(), "/cache-"+version)
	successFile := filepath.Join(targetDir, "success")
	extractedDir := filepath.Join(targetDir, "extract")
	command := filepath.Join(extractedDir, "shell/bin/fsh")
	if !file_helpers.FileExists(successFile) {
		downloadedFile := filepath.Join(targetDir, "downloaded.tar.gz")
		extractedDir := filepath.Join(targetDir, "extract")

		os.MkdirAll(extractedDir, 0700)

		ui := terminal.NewStdUI()
		fileDownloader := new(downloader.FileDownloader)
		fileDownloader.ProxyReader = downloader.NewProgressBar(ui.Writer())
		trace.Logger.Println("Downloading shell to " + downloadedFile)
		fileDownloader.DownloadTo(url, downloadedFile)
		trace.Logger.Println("Downloaded shell to " + downloadedFile)

		trace.Logger.Println("Extracting shell to " + extractedDir)
		archiver.TarGz.Open(downloadedFile, extractedDir)
		trace.Logger.Println("Extracted shell to " + extractedDir)

		MakeExecutable(command)
		distDir := filepath.Join(extractedDir, "shell/dist")
		MakeExecutable(distDir)

		os.OpenFile(successFile, os.O_RDONLY|os.O_CREATE, 0666)
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
		Name:    "shell",
		Version: GetVersion(),
		Commands: []plugin.Command{
			{
				Name:        "shell",
				Alias:       "fsh",
				Description: "Cloud shell",
				Usage:       "ibmcloud shell",
			},
		},
	}
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
