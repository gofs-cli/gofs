![gofs logo](/docs/gofs.svg)

# gofs (Go Full Stack)

Lightweight templates for secure full-stack go apps.

## Documentation

Developer documentation is available here [https://gofs.dev](https://gofs.dev).

## Installation

With Go 1.23 or greater installed, run:

```bash
go install github.com/gofs-cli/gofs@latest
```

Or from source:

```bash
git clone git@github.com:gofs-cli/gofs.git
cd gofs
go mod tidy
go install
```

## Usage

```bash
gofs
```

## Current Status

In development but used in production at one of europe's largest tech companies.

## Using generated templates

The template includes several modules that are optional and should be deleted to reduce build size. For example we include a postgres connector and a cloudsql connector for conveniance, but you should likely only need one of them.
