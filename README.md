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