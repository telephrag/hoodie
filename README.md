# Hoodie

Hoodie is a markup language that compiles into Valve KeyValues used in various Source 1 games. Hoodie implements ability to utilise reusable code blocks called traits.

# Install

## Prebuilt executable

Download executable from releases section on this github page.

## Build from source

NOTE: hoodie is yet to be tested on Windows

Dependecies:
 - go 1.21.3

How to:
 - download this repository
 - `go build -o=hoo main.go` inside downloaded directory on Linux
 - `go build -o=hoo.exe main.go` on Windows
 - on Linux executable will be `hoo` and `hoo.exe` on Windows

# Usage

See folder `example` and its contents. 

Windows: `.\hoo.exe -d=path\to\project`

Linux: `./hoo -d=path/to/project`
