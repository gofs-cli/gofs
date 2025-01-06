# gofs (Golang Full Stack)

`gofs` is a 10100 CLI tool to generate opinionated secure projects

## Installation

```bash
git clone git@github.com:gofs-cli/go-fs.git
cd go-fs
go mod tidy
go install
```

## Usage

```bash
gofs
```

## Current Status

In development but used in production at one of europe's largest tech companies.

## Features + Roadmap

Full docs site coming soon

## Using generated templates

The template includes several modules that are optional and should be deleted to reduce build size. For example we include a postgres connector and a cloudsql connector for conveniance, but you should likely only need one of them.
