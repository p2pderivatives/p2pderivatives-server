[![CircleCI](https://circleci.com/gh/cryptogarageinc/p2pderivatives-server.svg?style=svg&circle-token=54264d31d871e4b527f2c942d40a821199ef45c4)](https://circleci.com/gh/cryptogarageinc/p2pderivatives-server)

# p2pderivatives-server
Repository for the P2PDerivatives server

## Requirements
Confirmed working with Go v1.16.6.
To install protoc, use the script in `scripts/install_protoc.sh`.

## Getting started

The protobuf files are included as submodules.
When cloning the repository, use `git clone --recursive <url>` to clone the submodule as well.
Alternatively, run `git submodule update --init --recursive` within the repository root folder.

Run `make setup` to setup the repository.
You will need to setup a `postgresql` database connection in the configuration file (or via environment variables).  
You can easily setup a running database using `make gen-ssl-certs` to generate the db certificates and then `docker-compose up db`.
Once that is done, the server can be run locally using `make run-local-server`.
A cli client tool can be used to interact with the server.
Build it using `make client`.
Then run `./bin/p2pdclient` to see the list of available commands.

## Running using Docker

### Docker Compose
You can easily start and build the docker environment using `docker-compose`  
To build from scratch the server use: `docker-compose up --build`
Note that you will need to run `make gen-ssl-certs` to generate a certificate for the database.
 
### Building the image

In the root of the repository:

`docker build -t p2pd-server .`

The name `p2pd-server` can be changed to any other docker compliant name.

### Running the container
Once built, you can start running the server (a database connection is necessary):

`docker run -p 8080:8080 p2pd-server`

If your image name is different, please use the one specified before. 
The port the internal server is mapped to can be specified by changing the initial number of the pair, e.g.: `-p 5000:8080` to map to local port 5000.
To run the container in the background use the `-d` flag.

### Docker configuration
You can override any variable from the configuration file `.yaml` using `environment variables` with this format `APPNAME_MY_PATH_TO_VARIABLE`  
Exemple:  
- To override the `database.host` property in configuration file : `P2PD_DATABASE_HOST=mynewhost`