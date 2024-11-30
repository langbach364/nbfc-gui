# NBFC Interface

A user-friendly interface for interacting with NBFC (Notebook FanControl) on Linux systems.

## Prerequisites

- [NBFC-Linux](https://github.com/nbfc-linux/nbfc-linux) must be installed and properly configured on your system
- This interface follows the [NBFC Protocol Specifications](https://github.com/nbfc-linux/nbfc-linux/blob/main/PROTOCOL.md)

## Installation

### NixOS Users
All required configurations and dependencies are pre-configured in the module file.

### Other Linux Distributions
Please ensure you have:
1. Installed NBFC-Linux and test and confirm it runs on your device
2. Configured necessary data packages
3. Set up proper system permissions

## Customization

You can customize the interface by modifying `main.go` according to your needs.

## Building & Running

To build and apply your changes, run:

```bash
go build -ldflags="-s -w" -o ./bin/nbfc-gui main.go
```
