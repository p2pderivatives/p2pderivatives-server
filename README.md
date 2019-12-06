[![CircleCI](https://circleci.com/gh/cryptogarageinc/p2pderivatives-server.svg?style=svg&circle-token=54264d31d871e4b527f2c942d40a821199ef45c4)](https://circleci.com/gh/cryptogarageinc/p2pderivatives-server)

# p2pderivatives-server
Repository for the P2PDerivatives server

## Requirements
Confirmed working with Go v1.12.7.
To install protoc, use the script in `scripts/install_protoc.sh`.

## Getting started

The protobuf files are included as submodules.
When cloning the repository, use `git clone --recursive <url>` to clone the submodule as well.
Alternatively, run `git submodule update --init --recursive` within the repository root folder.

Run `make setup` to setup the repository.
Once that is done, the server can be run locally using `make run-local-server`.
A cli client tool can be used to interact with the server.
Build it using `make client`.
Then run `./bin/p2pdclient` to see the list of available commands.
