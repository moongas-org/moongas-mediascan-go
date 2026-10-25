<img src="https://raw.githubusercontent.com/moongas-org/moongas-mediatunes-web-vue/refs/heads/main/client/public/moongas.svg" width="128" height="128">

# moongas-mediascan-go

> 🚧 **Status: Work in Progress (WIP)**  
> This project is currently under active development. Features, APIs, and documentation are subject to change.

---

## Overview

**Simple and fast Golang command-line utility for scanning media library files and metadata files, extracting metadata (e.g. ID3v2 tags) and saving it into an SQLite database.**

### A component of the `moongas` ecosystem of media library tools

- [moongas-collection-demo](https://github.com/moongas-org/moongas-collection-demo) [![CI](https://github.com/moongas-org/moongas-collection-demo/actions/workflows/ci.yml/badge.svg)](https://github.com/moongas-org/moongas-collection-demo/actions/workflows/ci.yml) - Example Moongas media collection (metadata only)
- [moongas-mediatunes-web-vue](https://github.com/moongas-org/moongas-mediatunes-web-vue) [![CI](https://github.com/moongas-org/moongas-mediatunes-web-vue/actions/workflows/ci.yml/badge.svg)](https://github.com/moongas-org/moongas-mediatunes-web-vue/actions/workflows/ci.yml) - A Deno-tooled TypeScript/Vue SPA for Moongas hybrid media collections, pairing with the separate moongas-mediatunes-svc-python-blacksheep backend to seemlessly blend offline and streaming playback
- [moongas-mediatunes-svc-python-blacksheep](https://github.com/moongas-org/moongas-mediatunes-svc-python-blacksheep) [![CI](https://github.com/moongas-org/moongas-mediatunes-svc-python-blacksheep/actions/workflows/ci.yml/badge.svg)](https://github.com/moongas-org/moongas-mediatunes-svc-python-blacksheep/actions/workflows/ci.yml) - Python+BlackSheep implementation of API service for Moongas hybrid media collections—backend for Moongas mediatunes web application (moongas-mediatunes-web-vue)
- [moongas-mediatunes-svc-java-javalin](https://github.com/moongas-org/moongas-mediatunes-svc-java-javalin) [![CI](https://github.com/moongas-org/moongas-mediatunes-svc-python-blacksheep/actions/workflows/ci.yml/badge.svg)](https://github.com/moongas-org/moongas-mediatunes-svc-java-javalin/actions/workflows/ci.yml) - Java+Javalin implementation of API service for Moongas hybrid media collections—backend for Moongas mediatunes web application (moongas-mediatunes-web-vue)
- [moongas-mediascan-go](https://github.com/moongas-org/moongas-mediascan-go) [![CI](https://github.com/moongas-org/moongas-mediascan-go/actions/workflows/ci.yml/badge.svg)](https://github.com/moongas-org/moongas-mediascan-go/actions/workflows/ci.yml) - Golang module to scan media collections and Moongas YAML metatadata, outputs Moongas database
- [moongas-mediascan-python](https://github.com/moongas-org/moongas-mediascan-python) [![CI](https://github.com/moongas-org/moongas-mediascan-python/actions/workflows/ci.yml/badge.svg)](https://github.com/moongas-org/moongas-mediascan-python/actions/workflows/ci.yml) - Python library with data classes for loading Moongas mediascan databases and YAML metadata files
- [moongas-mediascripts-python](https://github.com/moongas-org/moongas-mediascripts-python) [![CI](https://github.com/moongas-org/moongas-mediascripts-python/actions/workflows/ci.yml/badge.svg)](https://github.com/moongas-org/moongas-mediascripts-python/actions/workflows/ci.yml) - Python scripts for working with Moongas media collections.
- [moongas-mediatest-python-pytest](https://github.com/moongas-org/moongas-mediatest-python-pytest) [![CI](https://github.com/moongas-org/moongas-mediatest-python-pytest/actions/workflows/ci.yml/badge.svg)](https://github.com/moongas-org/moongas-mediatest-python-pytest/actions/workflows/ci.yml) - Python tool for enforcing media collection rules (implemented with `pytest`)

```mermaid
graph TD;
    A[Start] --- B(Choose Frontend and Backend);
    B --- C{Choose Backend};
    C ---|Python| D[mediatunes-svc-python-blacksheep];
    C ---|Java| E[mediatunes-svc-java-javalin];
    B --- F{Choose Frontend};
    F ---|Deno+Vue| G[mediatunes-web-vue];
    F ---|TBD| H[tbd];
    D ---|has dependency| I[mediascan-python];
    I ---|loads| J[mediascan.db];
    J ---|generates| K[mediascan-go];
    E ---|loads| J[mediascan.db];
    J ---|validates| L[mediatest-python-pytest];
    J ---|reads readonly| M[mediascripts-python];
```

## Features

- Suports reading metadata Moongas metadata file formats such as `artist.yml` and `album.yml` files
- Output format (SQLite database) is compatible with the Moongas `mediascan` python package

## Database Schema

The output database contains three tables:

| Table | Contents |
| ----- | -------- |
| mediafile | Track metadata from individual media files (ID3 tag information read from `mp3` files) |
| artist | Metadata from `artist.yml` files |
| album | One row per unique media-file album path, with `path`, `name`, `year`, and `explicitcover` fields |

Album metadata comes from `album.yml` in the album directory: `albumData.name`,
`albumData.dateReleased.y`, and `albumData.explicitCover`. If the file is
missing, the name and year come from the first media file associated with that
album path in scan order, and `explicitcover` defaults to `false`.

## Usage

### Scan to Database

- [`cmd/scan-to-db`](./cmd/scan-to-db) - Scans media libraries for mediafiles and moongas `artist.yml` and `album.yml` files, outputs an SQLite database (`.db`) file
```bash
go run cmd/scan-to-db mediascan-config.yml mediascan.db
```

### Generate Album YAML

- [`cmd/scan-to-album-yaml`](./cmd/scan-to-album-yaml) - Creates an `album.yml`
  in each unique album directory found by the media-file scan. Existing files
  are always preserved; overwrite flags are not supported.
- Uses the first media file in scan order for `albumData.name` (album tag) and
  `albumData.dateReleased.y` (year tag). Sets `albumData.explicitCover` to
  `false` and omits release month and day.

```bash
go run ./cmd/scan-to-album-yaml mediascan-config.yml [media-root]
```

### Update Track Fingerprints

- [`cmd/fpcalc-to-track-yaml`](./cmd/fpcalc-to-track-yaml) runs `fpcalc --json`
  for media files that have an existing matching track `.yml` file but lack
  fpcalc duration or fingerprint data. It only updates the `fpcalcData` fields
  and creates a `.yml.bak` backup before modifying a track YAML file.
- Requires `fpcalc` to be installed and available on `PATH`.
- Accepts the same overwrite flags as the other scan commands:
  `--overwrite-existing` and `--overwrite-newer`.

```bash
go run ./cmd/fpcalc-to-track-yaml mediascan-config.yml [media-root]
```

### Scan to YAML (deprecated)

These commands scan media files and `artist.yml` files and output massive combined YAML (.yml) files. This was the original implementation before switching to sqlite. These may be removed in the future.

- [`cmd/scan-to-files-yaml`](./cmd/scan-to-files-yaml-yaml) - Recursively scan a directory for media files, extract metadata (including ID3v2 tags from both MP3 and M4A files), and save the output in a mediafiles YAML file (`files.yml`). 
- Reads configuration from YAML file e.g. [mediascan-config.yml](./mediascan-config.yml)
- Has only two required command-line arguments: `<config-yaml-filepath>` and `<output-yaml-filepath>`
- Created specifically to run fast on a Raspberry Pi single-board computer as part of another project of mine.
- Usage:
```bash
go run cmd/scan-to-files-yaml mediascan-config.yml files.yml
```
- [`cmd/scan-to-artists-yaml`](./cmd/scan-to-artists-yaml) - Scans artist directories for `artist.yml` files and aggregates them into a combined artists YAML file (`artists.yml`).
- Usage:
```bash
go run cmd/scan-to-artists-yaml mediascan-config.yml artists.yml
```

## mediascan YAML Configuration file Reference

The  YAML configuration file (example: [mediascan-config.yml](./mediascan-config.yml)) supports the following parameters:

| Property | Description |
| -------- | ----------- |
| `mediadir` | the path to the directory to be scanned |
| `mediaexts` | the file extensions to be included in the scan |
| `excludepath` | path substrings to exclude from scan (case-insensitive) |
| `excludetitle` | id3 title substrings to exclude from scan results (case-insensitive) |
| `excludetitle` | id3 artist substrings to exclude from scan results (case-insensitive) |
| `excludealbum` | id3 album substrings to exclude from scan results (case-insensitive) |
| `excludegenre` | id3 genre substrings to exclude from scan results (case-insensitive) |
| `sortby` | (`year`, `artist`, `none`) media file sort options |
| `getmp3duration` | (`true`, `false`) whether to calculate mp3 duration using `tcolgate/mp3` (**warning: slow*) |

## Problem: `package moongas-mediascan-go/cmd/scan-to-db is not in std (/usr/lib/go-1.26/src/moongas-mediascan-go/cmd/scan-to-db)`

Solution:

```bash
cd moongas-mediascan-go
go build ./cmd/scan-to-db
```

### Dependencies
- [gopkg.in/yaml.v3](https://pkg.go.dev/gopkg.in/yaml.v3) (Used for generating files.yml)
- [dhowden/tag](https://github.com/dhowden/tag) (Used for reading ID3 tags)
    - `dhowden/tag` is a nice little Go library for ID3, MP4 and OGG/FLAC metadata parsing. I added support for [genre codes in the 148-191 range](https://en.wikipedia.org/wiki/List_of_ID3v1_genres#Extension_by_Winamp) and David was kind enough to merge my [PR](https://github.com/dhowden/tag/pull/103).
- [tcolgate/mp3](https://github.com/tcolgate/mp3) (Used for calculating MP3 duration)
    - Note: Mp3 duration calculation can be disabed by setting `getmp3duration: false` in `mediascan-config.yml`. 
    - It's currently disabled because this library is suddenly unexpectedly slow for me.



## Performance Demo 1.5 - desktop PC with decade old i7 CPU
Much slower on first run. Why?

First run

```bash
2022/12/14 06:55:32 Successfully loaded 9124 media files
2022/12/14 06:55:32 Skipped 179 excluded media files

real	3m6.427s
user	0m4.687s
sys	0m5.567s

Second run

```bash
2022/12/14 07:09:53 Successfully loaded 9124 media files
2022/12/14 07:09:53 Skipped 179 excluded media files

real	0m3.637s
user	0m2.410s
sys	0m1.299s

```

## Performance Demo 2 - Raspberry Pi 4 model B
```bash
$ time go run cmd/ scan-to-files-yaml-yaml mediascan-config.yml files.yml
2022/07/04 15:06:39 Successfully loaded 8376 media files

real	0m18.652s
user	0m8.251s
sys	0m11.406s
```

## Performance Demo 3 - Raspberry Pi Zero 2W
```bash
$ time go run cmd/ scan-to-files-yaml-yaml mediascan-config.yml files.yml
2022/07/04 15:21:14 Successfully loaded 8376 media files

real	1m18.217s
user	0m18.136s
sys	0m14.305s
```

## Installing Dependencies

> 🚧 **IIVO Warning: Information Is Very Outdated**  
> The information below is outdated and retained for development purposes only.

Install Go programming language compiler, linker, compiled stdlib and supplementary Go tools
```bash
sudo apt update
sudo apt upgrade
sudo apt install golang-go golang-src golang-doc golang-golang-x-tools
```

Edit `~/.bashrc` by adding the following lines:

```bash
export GOPATH=~/go
export GOROOT=/usr/local/go
export PATH=$PATH:$GOPATH/bin
export PATH=$PATH:$GOROOT/bin
```
Then reload it: `source ~/.bashrc`

Note: `GOROOT` directory may vary. To confirm, check the `go` symlink, e.g.:
```bash
$ which go
/usr/bin/go
$ ls -l /usr/bin/go
lrwxrwxrwx 1 root root 21 Sep 16  2020 /usr/bin/go -> ../lib/go-1.15/bin/go
```
So based on the above output on the rpi we need: `export GOROOT=/usr/lib/go-1.15`

To test it (Raspberry Pi):
```bash
$ $GOROOT/bin/go version
go version go1.15.15 linux/arm64
```

To test it (Linux desktop):
```bash
$ $GOROOT/bin/go version
go version go1.18.1 linux/amd64
```

Install Go dependency 1 of 2: `tag`
```bash
go get github.com/dhowden/tag/cmd/tag
```

Install Go dependency 2 of 2: `yaml.v3`
```bash
go get gopkg.in/yaml.v3
```

Now we can see that these packages were installed into `$GOPATH/src`:
```bash
$ ls $GOPATH/src
github.com  gopkg.in
```

Now install mediascan:
```bash
go get github.com/moongas-org/moongas-mediascan-go
cd ~/go/src/github.com/moongas-org/moongas-mediascan-go/ && go install
```

Edit [mediascan-config.yml](mediascan-config.yml) and set the `mediadir` to the desired directory path.

Now you should be all set to run mediascan.

You can run it from the mediascan directory like this:
```bash
go run cmd/ scan-to-files-yaml-yaml mediascan-config.yml files.yml
```

Or you can run it from any directory like this:
```bash
go run github.com/moongas-org/moongas-mediascan-go mediascan-config.yml files.yml
```

Happy scanning!
