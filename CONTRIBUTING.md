# Cosmos SDK Tutorials repo

This repo contains the code for the `gex` in-terminal explorer.

## Contributing

Thank you for helping us to create and maintain awesome software.

- To provide feedback, file an issue and provide abundant details to help us understand how we can make it better.
- To provide feedback and a fix, you can make a direct contribution. This repo is protected. If you're not a member or maintainer, fork the repo and then submit a PR from your forked repo to master.

## Install the latest version

To install the latest version from GitHub, execute the following command in your terminal:

`go get -u github.com/cosmos/gex`

Then run the code with executing gex in the terminal:

`gex`

## Code structure

Currently the code resides in the `main.go` file. Additionally, there is an example how-to-connect-with-websocket file in `websocket.go`, which shows how the websocket connection is created to Cosmos SDK.

Websocket takes care of the blocks, transactions and validators feed. A combination of the websocket and API requests are used in order to also feature connected peers or the version of Cosmos SDK currently running.

The UI framework used for the terminal is [Termdash](https://github.com/mum4k/termdash). Visit the GitHub page to get inspiration on what is possible to do with Termdash, how to work with it, or which charting styles and paginations are available.

