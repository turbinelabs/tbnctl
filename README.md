
[//]: # ( Copyright 2017 Turbine Labs, Inc.                                   )
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

[![Apache 2.0](https://img.shields.io/hexpm/l/plug.svg)](LICENSE)
[![GoDoc](https://https://godoc.org/github.com/turbinelabs/tbnctl?status.svg)](https://https://godoc.org/github.com/turbinelabs/tbnctl)
[![CircleCI](https://circleci.com/gh/turbinelabs/tbnctl.svg?style=shield)](https://circleci.com/gh/turbinelabs/tbnctl)

The tbnctl project provides a command-line interface to the Turbine Labs public API.

## Requirements

- Go 1.7.4 or later (previous versions may work, but we don't build or test against them)

## Dependencies

The tbnctl project depends on these packages:

- [api](https://github.com/turbinelabs/api)
- [cli](https://github.com/turbinelabs/cli)
- [codec](https://github.com/turbinelabs/codec)
- [nonstdlib](https://github.com/turbinelabs/nonstdlib)
- [tbnproxy](https://github.com/turbinelabs/tbnproxy)

The tests depend on our [test package](https://github.com/turbinelabs/test),
and on [gomock](https://github.com/golang/mock). Some code is generated using
[genny](http://github.com/falun/genny).

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
[genny](http://github.com/falun/genny). If you need to modify the generated
code, modify [`adaptor.genny`](adaptor.genny), then:

```
go get -u github.com/falun/genny
go install github.com/falun/genny
go generate github.com/turbinelabs/tbnctl
```

## Versioning

Please see [Versioning of Turbine Labs Open Source Projects](http://github.com/turbinelabs/developer/blob/master/README.md#versioning).

## Pull Requests

Patches accepted! Please see [Contributing to Turbine Labs Open Source Projects](http://github.com/turbinelabs/developer/blob/master/README.md#contributing).

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

## Proxy Configuration Dump

The `proxy-config` sub-command can be used to produce the NGINX configuration for
a named Proxy. See `tbnctl help proxy-config` for more detail.

## A Look into... THE FUTURE

We will continue to improve and extend `tbnctl` over time. Some examples of
things we might someday add include:

- parity with the web app for our the release workflow
- access to the Stats API
- better defaults when creating more complex objects (eg Routes, SharedRules)

