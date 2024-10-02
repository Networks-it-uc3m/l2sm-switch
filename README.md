# L2S-M Switch

An open virtual switch implementation written in Go, designed for use within the L2S-M main component as a Docker image that constitutes either the network overlay or the network edge devices. It provides a gRPC server and CLI tools to enable the creation of OVS bridges, dynamic port attachment, and dynamic VXLAN creation, utilizing OpenFlow 1.3.

## Table of Contents

- [Features](#features)
- [Project Structure](#project-structure)
- [Installation](#installation)
  - [Prerequisites](#prerequisites)
  - [Building from Source](#building-from-source)
  - [Building the Docker Image](#building-the-docker-image)
- [Usage](#usage)
  - [CLI Tools](#cli-tools)
  - [Running the gRPC Server](#running-the-grpc-server)
  - [Sample Configuration](#sample-configuration)
- [Configuration](#configuration)
- [Makefile Targets](#makefile-targets)
- [Setup Scripts](#setup-scripts)
- [Contributing](#contributing)
- [License](#license)
- [Acknowledgments](#acknowledgments)

## Features

- **Open Virtual Switch Implementation**: Built with Go for high performance.
- **Dockerized Deployment**: Simplifies deployment and management via Docker.
- **gRPC Server**: Provides a remote management interface.
- **CLI Tools**: Command-line utilities for initializing the switch, adding ports, and managing VXLANs.
- **Dynamic Configuration**: Supports dynamic creation of OVS bridges, ports, and VXLANs.
- **OpenFlow 1.3 Support**: Compatible with OpenFlow 1.3 protocol.
- **Integration with L2SM**: Designed to function within the L2SM main component.
## Installation

### Prerequisites

- **Go**: Version 1.21 or higher.
- **Docker**: Installed and running.
- **Protobuf Compiler**: For generating gRPC code.
- **Open vSwitch**: Installed if running outside Docker.

### Building from Source

Clone the repository and navigate to the project directory:

```bash
git clone https://github.com/Networks-it-uc3m/l2sm-switch.git
cd l2sm-switch
```

#### Generating gRPC Code

Modify the contents in api/v1/ned.proto, and later generate the gRPC code from the `.proto` file:

```bash
make generate-proto
```

#### Building the Go Binaries

Use the provided build script:

```bash
chmod +x ./build/build-go.sh
./build/build-go.sh
```

This script compiles the CLI tools and places the binaries in `/usr/local/bin/`.

Alternatively, build each component individually:

```bash
go build -o /usr/local/bin/l2sm-init ./cmd/l2sm-init
go build -o /usr/local/bin/l2sm-add-port ./cmd/l2sm-add-port
go build -o /usr/local/bin/l2sm-vxlans ./cmd/l2sm-vxlans
go build -o /usr/local/bin/ned-server ./cmd/ned-server
```

### Building the Docker Image

Build the Docker image using the provided Makefile, after setting up the image name and repository:

```bash
make docker-build
```


## Usage

### CLI Tools

The project includes several CLI tools located in the `cmd` directory:

- **l2sm-init**: Initializes the OVS bridge.
- **l2sm-add-port**: Adds ports to the OVS bridge.
- **l2sm-vxlans**: Manages VXLANs.
- **ned-server**: Starts the gRPC server.


### Running the gRPC Server

Start the gRPC server using the `ned-server` command:

```bash
ned-server --config_dir ./config/config.json --neighbors_dir ./config/neighbors.json
```

### Sample Configuration



This creates sample configuration files in the `config` directory, for development purposes. 
The Makefile uses environment variables defined in a `.env` file:

- `CONTROLLERIP`: IP address of the controller.
- `NODENAME`: Name of the node.
- `NEDNAME`: Name of the network edge device (default is `brtun`).

So, first create a `.env` file with the required variables:

```bash
CONTROLLERIP=192.168.1.1
NODENAME=node1
NEDNAME=brtun
```
And then, generate the sample `config.json` and `neighbors.json` files:

```bash
make sample-config
```

The neighbors.json contains an array that must be filled manually with the ip addresses of every neighbour you want to initially attach the ned to. 



### Setup Scripts

- **setup_switch.sh**: Sets up the OVS switch upon container startup.
- **setup_ned.sh**: Configures the network edge device.

These scripts are copied into the Docker image and executed as part of the container initialization.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request with your improvements.


Before contributing, please ensure you have read the [CONTRIBUTING](CONTRIBUTING.md) guidelines. (pending)

For any questions feel free to contact the us via 100383348@alumnos.uc3m.es 

## License

This project is licensed under the terms of the [Apache License 2.0](LICENSE).

