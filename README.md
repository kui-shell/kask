# IBMCloud CLI Plugin for Cloud Shell

# Get started

You firstly need [Go](http://www.golang.org) installed on your machine, and set up a [GOPATH](http://golang.org/doc/code.html#GOPATH). Then clone this repository into `$GOPATH/src/github.ibm.com/shell-cli`. 

This project uses [govendor](https://github.com/kardianos/govendor) to manage dependencies. Go to the project directory and run the following command to restore the dependencies into vendor folder:

```bash
$ govendor sync
```

and then run tests:

```bash
$ go test ./...
```

# Build and run plugin

Download and install the Bluemix CLI. See instructions [here](https://clis.ng.bluemix.net).

Compile the plugin source code with `go build` command, for example

```bash
$ go build shell.go
```

Install the plugin:

```bash
$ bluemix plugin install ./shell
```

# Use the plugin

For example:
```
ibmcloud shell host get
```

```
ibmcloud fsh shell
```

# Create a plugin repo in object storage for testing:

First, create a file called list that looks like this:
```
{
  "plugins": [{
    "name": "shell-test",
    "aliases": null,
    "description": "Shell test",
    "created": "2016-01-14T00:00:00Z",
    "updated": "2018-07-05T00:00:00Z",
    "company": "IBM",
    "homepage": "https://plugins.ng.bluemix.net",
    "authors": [],
    "versions": [
    {
      "version": "1.6.1",
      "updated": "2016-01-28T00:00:00Z",
      "doc_url": "",
      "min_cli_version": "",
      "binaries": [
        {
          "platform": "osx",
          "url": "https://plugins.ng.bluemix.net/downloads/bluemix-plugins/auto-scaling/auto-scaling-darwin-amd64-0.2.1",
          "checksum": "52947857a431afafb88882d6683e86927871d738"
        },
        {
          "platform": "win64",
          "url": "https://plugins.ng.bluemix.net/downloads/bluemix-plugins/auto-scaling/auto-scaling-windows-amd64-0.2.1.exe",
          "checksum": "79ba07cadedc32ec1eb25c954a9b29f870b25e49"
        },
        {
          "platform": "linux64",
          "url": "https://plugins.ng.bluemix.net/downloads/bluemix-plugins/auto-scaling/auto-scaling-linux-amd64-0.2.1",
          "checksum": "fa3529990ca233f20cba6d89b3d871d56bef41d9"
        }
      ],
      "api_versions": null,
      "releaseNotesLink": ""
    }
  ]
  }]
}
```

Then upload it to a COS bucket, making it public using this command:
```
curl -X "PUT" "https://s3-api.us-geo.objectstorage.softlayer.net/shelldist/bx/list" -H "x-amz-acl: public-read" -H "Authorization: $IAM_TOKEN" -H "Content-Type: application/json" -d@"list"
```

Then add the repo:
```
bluemix plugin repo-add shell-dev-repo https://s3-api.us-geo.objectstorage.softlayer.net/shelldist/
```

Now you'll see the the plugins show up when you issue:
```
bluemix plugin repo-plugins
```

To uninstall the repo simply issue:
```
bluemix plugin repo-remove shell-dev-repo
```
