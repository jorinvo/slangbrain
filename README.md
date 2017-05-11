# Slangbrain

[![GoDoc](https://godoc.org/qvl.io/slangbrain?status.svg)](https://godoc.org/qvl.io/slangbrain)
[![Build Status](https://travis-ci.org/qvl/slangbrain.svg?branch=master)](https://travis-ci.org/qvl/slangbrain)
[![Go Report Card](https://goreportcard.com/badge/qvl.io/slangbrain)](https://goreportcard.com/report/qvl.io/slangbrain)


## Install

- With [Go](https://golang.org/):
```
go get qvl.io/slangbrain
```

- With [Homebrew](http://brew.sh/):
```
  brew install qvl/tap/slangbrain
```

- Download from https://github.com/qvl/slangbrain/releases


## Development

1. `go get github.com/jorinvo/slangbrain` to get project
2. `cd $GOPATH/src/github.com/jorinvo/slangbrain`
3. `go run cmd/slangbrain-telegram/main.go` to run project
4. `gvt update -all` to update dependencies (requires [gvt](https://github.com/FiloSottile/gvt))
5. For local development use `ngrok` (https://ngrok.com) as webhook


### Contributing

Make sure to use `gofmt` and create a [Pull Request](https://github.com/qvl/slangbrain/pulls).

When changing external dependencies please use [dep](https://github.com/golang/dep/) to vendor them.


### Releasing

Push a new Git tag and [GoReleaser](https://github.com/goreleaser/releaser) will automatically create a release.


## Thank you

To these helpful open source projects Slangbrain is built on top of:



## License

[MIT](./license)
