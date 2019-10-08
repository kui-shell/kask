# Kask: a manager for UI-based kubectl plugins

[![Build Status](https://travis-ci.org/kui-shell/kask.svg?branch=master)](https://travis-ci.org/kui-shell/kask)

The goal of `kask` is to enrich the world of
[krew](https://github.com/kubernetes-sigs/krew) with graphics. Whereas
`krew` allows users to install terminal-based extensions to `kubectl`,
`kask` aims to allow users to install graphical plugins. We consider a
graphical plugin to `kubectl` one that starts from a terminal, but,
rather than responding to commands with ASCII art, responds instead
with a graphical popup window.

# Using `kask`

Coming soon

# Architecture of `kask`

`kask` acts as a front-end to [Kui](https://github.com/IBM/kui). Kui
is a project that leverages [electron](https://electronjs.org) and
[webpack](https://webpack.js.org) to deliver a graphically enhanced
terminal experience. `kask` thusly has the following responsibilities:

- Fetch a base Kui electron build.
- Invoke Kui's plugin manager to provision a plugin (say "xxx") that
  the user has selected
- Generate and place the `kubectl-xxx` stubs such that `kubectl xxx`
  works to transparently pop up an electron window that provides the
  graphical response to the user's command line.

# Developing `kask`

`kask` is written as a [Go](http://www.golang.org) project. As such,
it has the normal development environment.  TO start, you will need to
have Go development environment installed on your machine; this
includes setting up a
[GOPATH](http://golang.org/doc/code.html#GOPATH).

## Clone and Contribute

```bash
cd $GOPATH && git clone git@github.com:kui-shell/kask.git kui-shell/kask
$ go get
$ go test ./...
```

## Build and Run

```bash
$ go build
./kask version
```
