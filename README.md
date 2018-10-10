ndraddrgen
====

[![Build Status](https://travis-ci.org/endurio/ndraddrgen.png?branch=master)](https://travis-ci.org/endurio/ndraddrgen)
[![GoDoc](https://godoc.org/gitlab.com/endurio/ndraddrgen?status.png)](http://godoc.org/gitlab.com/endurio/ndraddrgen)


ndraddrgen is a simple offline address generator for [endurio](https://endurio.org/).

It allows one to generate an address (along with either the private
key or a wallet seed) without a running wallet or daemon.

## Installation and updating

### Windows/Linux/BSD/POSIX - Build from source

Building or updating from source requires the following build dependencies:

- **Go 1.8 or 1.9**

  Installation instructions can be found here: http://golang.org/doc/install.
  It is recommended to add `$GOPATH/bin` to your `PATH` at this point.

- **Dep**

  Dep is used to manage project dependencies and provide reproducible builds.
  It is recommended to use the latest Dep release, unless a bug prevents doing
  so.  The latest releases (for both binary and source) can be found
  [here](https://github.com/golang/dep/releases).

Unfortunately, the use of `dep` prevents a handy tool such as `go get` from
automatically downloading, building, and installing the source in a single
command.  Instead, the latest project and dependency sources must be first
obtained manually with `git` and `dep`, and then `go` is used to build and
install the project.

**Getting the source**:

For a first time installation, the project and dependency sources can be
obtained manually with `git` and `dep` (create directories as needed):

```
git clone https://gitlab.com/endurio/ndraddrgen $GOPATH/src/gitlab.com/endurio/ndraddrgen
cd $GOPATH/src/gitlab.com/endurio/ndraddrgen
dep ensure
```

To update an existing source tree, pull the latest changes and install the
matching dependencies:

```
cd $GOPATH/src/gitlab.com/endurio/ndraddrgen
git pull
dep ensure
```

**Building/Installing**:

The `go` tool is used to build or install (to `GOPATH`) the project.  Some
example build instructions are provided below (all must run from the `ndraddrgen`
project directory).

To build a `ndraddrgen` executable and install it to `$GOPATH/bin/`:

```
go install
```

To build a `ndraddrgen` executable and place it in the current directory:

```
go build
```

## Usage

```
Usage: ndraddrgen [-testnet] [-simnet] [-regtest] [-noseed] [-h] filename
Generate a endurio private and public key or wallet seed.
These are output to the file 'filename'.

  -h 		    Print this message
  -testnet 	    Generate a testnet key instead of mainnet
  -simnet       Generate a simnet key instead of mainnet
  -regtest      Generate a regtest key instead of mainnet
  -noseed       Generate a single keypair instead of a seed
  -verify 	    Verify a seed by generating the first address
```
