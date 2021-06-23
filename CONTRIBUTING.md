# Cosmos SDK Tutorials Repo

This repo contains the code for the GEX in-terminal explorer. 

## Contributing

Thank you for helping us to create and maintain awesome software.

- To provide feedback, file an issue and provide abundant details to help us understand how we can make it better.
- To provide feedback and a fix, you can make a direct contribution. This repo is protected. If you're not a member or maintainer, fork the repo and then submit a PR from your forked repo to master.

## Install the Latest Version of GEX

The GEX installation requires [Go](https://golang.org/dl). If you don't already have Go installed, download the binary release that is suitable for your system and follow the installation instructions.

To install the latest version of GEX from GitHub, run the following command in a terminal window:

`go get -u github.com/cosmos/gex`

To launch the GEX block explorer in the terminal window, type:

`gex`

## GEX Code Structure

The GEX code resides in the `main.go` file. 

### Websocket 

An example how-to-connect-with-websocket file is provided in `websocket.go`. This example shows how to create the websocket connection to the Cosmos SDK.

Websocket takes care of the blocks, transactions and validators feed. A combination of the websocket and API requests are used in order to also feature connected peers or the version of Cosmos SDK currently running.

### UI Framework 

The UI framework for the terminal is Termdash. To see what is possible with Termdash, see [Termdash](https://github.com/mum4k/termdash) on GitHub. Learn how to work with Termdash and see the available charting styles and paginations.

