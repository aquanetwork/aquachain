# ZIP-based File System for serving HTTP requests

[![GoDoc](https://godoc.org/github.com/spkg/zipfs?status.svg)](https://godoc.org/github.com/spkg/zipfs)
[![Build Status (Linux)](https://travis-ci.org/spkg/zipfs.svg?branch=master)](https://travis-ci.org/spkg/zipfs)
[![Build status (Windows)](https://ci.appveyor.com/api/projects/status/tko2unyo9wm172e1?svg=true)](https://ci.appveyor.com/project/jjeffery/zipfs)
[![Coverage Status](https://coveralls.io/repos/github/spkg/zipfs/badge.svg?branch=master)](https://coveralls.io/github/spkg/zipfs?branch=master)
[![GoReportCard](https://goreportcard.com/badge/github.com/spkg/zipfs)](https://goreportcard.com/report/github.com/spkg/zipfs)
[![License](https://img.shields.io/badge/license-BSD-green.svg)](https://raw.githubusercontent.com/spkg/zipfs/master/LICENSE.md)

Package `zipfs` provides a convenient way for a HTTP server to serve
static content from a ZIP file.

Usage is simple. See the example in the
[GoDoc](https://godoc.org/github.com/spkg/zipfs) documentation.

## License

Some of the code in this project is based on code in the `net/http`
package in the Go standard library. For this reason, this package has
the same license as the Go standard library.
