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
 - `go build -o=hoodie main.go` inside downloaded directory on Linux
 - `go build -o=hoodie.exe main.go` on Windows
 - on Linux executable will be `main` and `main.exe` on Windows

# Usage

Windows: `hoodie.exe -d=path\to\project`

Linux: `hoodie -d=path/to/project`
