# kf

`kf` is like `cf` for Knative

## How to build

**Dependencies:**

[go mod](https://github.com/golang/go/wiki/Modules#quick-start) is used and required for dependencies

**Requirements:**

  - Golang `1.11.4`

**Building:**

```sh
$ go build ./cmd/...
```

**Notes:**

- `kf` CLI must be built outside of the `$GOPATH` folder unless you explicitly use `export GO111MODULE=on`.


## Development and releasing

We use [ko](https://github.com/google/ko) for rapid development and during the release process to build a full set of `kf` images and installation YAML.
