
[//]: # ( Copyright 2018 Turbine Labs, Inc.                                   )
[//]: # ( you may not use this file except in compliance with the License.    )
[//]: # ( You may obtain a copy of the License at                             )
[//]: # (                                                                     )
[//]: # (     http://www.apache.org/licenses/LICENSE-2.0                      )
[//]: # (                                                                     )
[//]: # ( Unless required by applicable law or agreed to in writing, software )
[//]: # ( distributed under the License is distributed on an "AS IS" BASIS,   )
[//]: # ( WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or     )
[//]: # ( implied. See the License for the specific language governing        )
[//]: # ( permissions and limitations under the License.                      )

# turbinelabs/tbnctl

**This project is no longer maintained by Turbine Labs, which has
[shut down](https://blog.turbinelabs.io/turbine-labs-is-shutting-down-and-our-team-is-joining-slack-2ad41554920c).**

[![Apache 2.0](https://img.shields.io/badge/license-apache%202.0-blue.svg)](LICENSE)
[![GoDoc](https://godoc.org/github.com/turbinelabs/tbnctl?status.svg)](https://godoc.org/github.com/turbinelabs/tbnctl)
[![CircleCI](https://circleci.com/gh/turbinelabs/tbnctl.svg?style=shield)](https://circleci.com/gh/turbinelabs/tbnctl)
[![Go Report Card](https://goreportcard.com/badge/github.com/turbinelabs/tbnctl)](https://goreportcard.com/report/github.com/turbinelabs/tbnctl)
[![codecov](https://codecov.io/gh/turbinelabs/tbnctl/branch/master/graph/badge.svg)](https://codecov.io/gh/turbinelabs/tbnctl)

The tbnctl project provides a command-line interface to the Turbine Labs public API.

## Requirements

- Go 1.10.3 or later (previous versions may work, but we don't build or test against them)

## Dependencies

The tbnctl project depends on these packages:

- [api](https://github.com/turbinelabs/api)
- [cli](https://github.com/turbinelabs/cli)
- [codec](https://github.com/turbinelabs/codec)
- [nonstdlib](https://github.com/turbinelabs/nonstdlib)

The tests depend on our [test package](https://github.com/turbinelabs/test),
and on [gomock](https://github.com/golang/mock). Some code is generated using
[codegen](http://github.com/turbinelabs/tools/codegen).

It should always be safe to use HEAD of all master branches of Turbine Labs
open source projects together, or to vendor them with the same git tag.

## Install

We plan to make `tbnctl` available via `brew`/`apt-get`/`yum`/etc in the future.
For now you can build and install it from source:

```
go get -u github.com/turbinelabs/tbnctl
go install github.com/turbinelabs/tbnctl
```

## Clone/Test

```
mkdir -p $GOPATH/src/turbinelabs
git clone https://github.com/turbinelabs/tbnctl.git > $GOPATH/src/turbinelabs/tbnctl
go test github.com/turbinelabs/tbnctl/...
```

## Generated code

Some code is generated using `go generate` and
[codegen](http://github.com/turbinelabs/tools/codegen). If you need to modify the generated
code, modify [`adaptor.template`](adaptor.template), then:

```
go get -u github.com/turbinelabs/tools/codegen
go install github.com/turbinelabs/tools/codegen
go generate github.com/turbinelabs/tbnctl
```

## Versioning

Please see [Versioning of Turbine Labs Open Source Projects](http://github.com/turbinelabs/developer/blob/master/README.md#versioning).

## Pull Requests

Patches accepted! Please see [Contributing to Turbine Labs Open Source Projects](http://github.com/turbinelabs/developer/blob/master/README.md#contributing).

## Code of Conduct

All Turbine Labs open-sourced projects are released with a
[Contributor Code of Conduct](CODE_OF_CONDUCT.md). By participating in our
projects you agree to abide by its terms, which will be carefully enforced.

# Basic Usage

Here we summarize what you can do with `tbnctl`. For more detailed help,
run `tbnctl -h`.

## CRUD Operations

`tbnctl` supports the following operations on Clusters, Domains, Proxies,
Routes, SharedRules, and Zones:

- list
- get
- create
- edit
- delete

Both `create` and `edit` will use the editor corresponding to the value of
`EDITOR` in your environment.

You can get detailed usage for each sub-command by typing `tbnctl help <cmd>`.

## Initial Environment Setup

The `init-zone` sub-command can be used to initialize a Zone with appropriate
Clusters, Domains, Proxies, Routes, and SharedRules. See `tbnctl help init-zone`
for more detail.

## A Look into... THE FUTURE

We will continue to improve and extend `tbnctl` over time. Some examples of
things we might someday add include:

- parity with the web app for our the release workflow
- access to the Stats API
- better defaults when creating more complex objects (e.g. Routes, SharedRules)
