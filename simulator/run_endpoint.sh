#!/bin/bash

# Set up the routing needed for the simulation
/setup.sh

# The following variables are available for use:
# - ROLE contains the role of this execution context, client or server
# - SERVER_PARAMS contains user-supplied command line parameters
# - CLIENT_PARAMS contains user-supplied command line parameters

# echo "193.167.0.100 client" >> /etc/hosts
# echo "193.167.100.100 server" >> /etc/hosts

if [ "$ROLE" == "client" ]; then
    echo "Wait for the simulator to start up."
    /wait-for-it.sh sim:57832 -s -t 30

    echo "Run the client HTTP3 requesting to 193.167.100.100:4433"
    go run client/client.go -url 193.167.100.100:4433
elif [ "$ROLE" == "server" ]; then
    # It is recommended to increase the maximum buffer size (https://github.com/quic-go/quic-go/wiki/UDP-Receive-Buffer-Size)
    # sysctl -w net.core.rmem_max=2500000

    echo "Run the server HTTP3 on 0.0.0.0:4433"
    go run server.go -addr 0.0.0.0:4433
fi