# Slangbrain

## Development

1. See `go run main.go -help` for more
2. `./run` to start development version locally
3. `dep ensure -update` to update dependencies (requires [dep](https://github.com/golang/dep))
4. `./deploy` to push latest version to server

## Misc

- Get the leo.org exporter ready to use as bookmarklet with:

```
npm i -g uglify
uglify -s utils/leoToCSV.js -o dist/leoToCSV.min.js && echo "javascript:$(cat dist/leoToCSV.min.js)" | copy
```
