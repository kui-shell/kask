package kask

import (
	"fmt"
	"github.com/mholt/archiver"
	. "github.ibm.com/composer/cloud-shell-cli/i18n"
	"log"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

type KrewComponent struct {
}

type Context interface {
	PluginDirectory() string
}

// THE PLUGIN_VERSION CONSTANT SHOULD BE LEFT EXACTLY AS-IS SINCE IT CAN BE PROGRAMMATICALLY SUBSTITUTED
const PLUGIN_VERSION = "dev"

type MainContext struct {}
func (MainContext) PluginDirectory() string {
	return "/tmp/nicko"
}

func Start() {
	argsWithoutProg := os.Args[1:]
	if len(argsWithoutProg) > 0 && argsWithoutProg[0] == "version" {
		version := GetVersion()
		fmt.Println(version.String())
		return
	}

	runner := KrewComponent{}
	context := MainContext{}
	runner.Run(context, os.Args)
}

func (p *KrewComponent) init() {
}

func (component *KrewComponent) Run(context Context, args []string) {
	component.init()
	kaskArgs := args[1:]

	cmd, err := component.DownloadDistIfNecessary(context)
	if err != nil {
		os.Exit(1)
		return
	}
	component.invokeRun(context, cmd, kaskArgs)
}

func (component *KrewComponent) invokeRun(context Context, cmd *exec.Cmd, kaskArgs []string) {
	cmd.Args = append(cmd.Args, kaskArgs...)

	log.Println(cmd.Args)

	if err := cmd.Start(); err != nil {
		fmt.Println("command failed!")
	}
}

func GetDistOSSuffix() string {
	switch runtime.GOOS {
	case "windows":
		return "-win32-x64.zip"
	case "darwin":
		return "-darwin-x64.tar.bz2"
	default:
		return "-linux-x64.zip"
	}
}

func GetRootCommand(extractedDir string) *exec.Cmd {
	switch runtime.GOOS {
	case "windows":
		// TODO verify
		return exec.Command(filepath.Join(extractedDir, "Kui-win32-x64\\Kui.exe"))
	case "darwin":
		return exec.Command(filepath.Join(extractedDir, "Kui-darwin-x64/Kui.app/Contents/MacOS/Kui"))
	default:
		// TODO verify
		return exec.Command(filepath.Join(extractedDir, "Kui-linux-x64/Kui"))
	}
}

func GetDistLocation(version string) string {
	/*
		production distributions:
			https://s3-api.us-geo.objectstorage.softlayer.net/ibm-cloud-shell-1.14.0/IBM%20Cloud%20Shell-darwin-x64.zip
			https://s3-api.us-geo.objectstorage.softlayer.net/ibm-cloud-shell-1.14.0/IBM%20Cloud%20Shell-win32-x64.zip
			https://s3-api.us-geo.objectstorage.softlayer.net/ibm-cloud-shell-1.14.0/IBM%20Cloud%20Shell-linux-x64.zip

		dev distributions:
			win32: https://s3-api.us-geo.objectstorage.softlayer.net/ibm-cloud-shell-dev/IBM%20Cloud%20Shell-win32-x64.zip
			macOS zip: https://s3-api.us-geo.objectstorage.softlayer.net/ibm-cloud-shell-dev/IBM%20Cloud%20Shell-darwin-x64.zip
			macOS tar.bz2: https://s3-api.us-geo.objectstorage.softlayer.net/ibm-cloud-shell-dev/IBM%20Cloud%20Shell-darwin-x64.tar.bz2
			linux-zip: https://s3-api.us-geo.objectstorage.softlayer.net/ibm-cloud-shell-dev/IBM%20Cloud%20Shell-linux-x64.zip
			headless: https://s3-api.us-geo.objectstorage.softlayer.net/ibm-cloud-shell-dev/IBM%20Cloud%20Shell-headless.zip
	*/

	host := "https://s3-api.us-geo.objectstorage.softlayer.net/kui-" + version
	DEV_OVERRIDE_HOST, overrideSet := os.LookupEnv("CLOUDSHELL_DIST")
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

func (p *KrewComponent) DownloadDistIfNecessary(context Context) (*exec.Cmd, error) {

	metadata := p.GetMetadata()
	version := metadata.Version.String()

	url := GetDistLocation(version)

	targetDir := filepath.Join(context.PluginDirectory(), "/cache-"+version)
	successFile := filepath.Join(targetDir, "success")
	extractedDir := filepath.Join(targetDir, "extract")
	log.Printf("successFile %s", successFile)

	command := GetRootCommand(extractedDir)
	command.Env = os.Environ()

	if _, err := os.Stat(successFile); err != nil {
		downloadedFile := filepath.Join(targetDir, "downloaded.zip")
		extractedDir := filepath.Join(targetDir, "extract")

		os.MkdirAll(extractedDir, 0700)

		if err := DownloadFile(downloadedFile, url); err != nil {
			handleError(err)
			return nil, err
		}
		log.Println("Downloaded kui-base to " + downloadedFile)

		log.Println("Extracting kui-base to " + extractedDir)
		if strings.HasSuffix(url, ".tar.bz2") {
			if err := archiver.DefaultTarBz2.Unarchive(downloadedFile, extractedDir); err != nil {
				handleError(err)
				return nil, err
			}
		} else {
			if err := archiver.DefaultZip.Unarchive(downloadedFile, extractedDir); err != nil {
				handleError(err)
				return nil, err
			}
		}

		log.Println("Extracted kui-base to " + extractedDir)

		if _, err := os.OpenFile(successFile, os.O_RDONLY|os.O_CREATE, 0666); err != nil {
			handleError(err)
			return nil, err
		}
	} else {
		log.Println("Using cached download")
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

func (component *KrewComponent) GetMetadata() Metadata {
	return Metadata{
		Name:    "function-composer",
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

func handleError(err error) {
	switch err {
	case nil:
		return
	default:
		log.Fatal(T("An error has occurred:\n{{.Error}}\n", map[string]interface{}{"Error": err.Error()}))
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
