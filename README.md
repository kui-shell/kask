# Kask: a Krew-based manager for UI plugins to kubectl

# Get started

You firstly need [Go](http://www.golang.org) installed on your machine, and set up a [GOPATH](http://golang.org/doc/code.html#GOPATH). 

```bash
cd $GOPATH && git clone git@github.com:kui-shell/kask.git kui-shell/kask
```

Pull in the required dependencies (at some point, may switch to govendor)
```bash
$ go get
```

and then run tests:

```bash
$ go test ./...
```

# Build and run plugin

Compile the plugin source code with `go build` command, for example

```bash
$ go build
```

# Try it Out
```bash
./kask kui version
```
