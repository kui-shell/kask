# IBMCloud CLI Plugin for Cloud Shell

# Get started

You firstly need [Go](http://www.golang.org) installed on your machine, and set up a [GOPATH](http://golang.org/doc/code.html#GOPATH). Then clone this repository into `$GOPATH/src/github.ibm.com/composer/cloud-shell-cli`. 

Pull in the required dependencies (at some point, may switch to govendor)
```bash
$ go get
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
$ bluemix plugin install ./cloud-shell-cli
```

# Use the plugin

For example:
```
ibmcloud shell host get
```

```
ibmcloud fsh shell
```

# CI Process

See https://github.ibm.com/composer/cloud-shell-cli/wiki/Travis-Build-&-Deployment

When PR is merged into master, the CLI "repo" will be available at `https://s3-api.us-geo.objectstorage.softlayer.net/shelldist/dev`

To test, add the repo:
```
bluemix plugin repo-add shell-dev-repo https://s3-api.us-geo.objectstorage.softlayer.net/shelldist/dev
```

Now you'll see the the plugins show up when you issue:
```
bluemix plugin repo-plugins
```

To uninstall the repo simply issue:
```
bluemix plugin repo-remove shell-dev-repo
```

# Publish to Bluemix Plugin Repo on YS1
After running `bin/build-all.sh`, using your IBM ID (assuming that you've been registered - ask on #bluemix-cli if not), run the following script to create a PR that the CLI team will merge in manually
```
node gen-and-update-ys1.js "version" "userName" "password"
```
TODO: we should work this in with a functional ID as part of the automated master build

# Promote to Bluemix Plugin Repo in Production

Use the Jenkins job here to manually promote:

https://wcp-cloud-foundry-jenkins.swg-devops.com/job/Promote%20from%20staging%20to%20production/build?delay=0sec