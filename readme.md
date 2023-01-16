# Foundry VTT World ID Reset tool

A tool to reset the IDs of all the documents in a Foundry VTT world. One use-case for it is if a world has been duplicated and used as the base for another world.

The tool works by generating new IDs for the documents within the "data" directory of a world and then replacing those references throughout the *.db files.

## Usage instructions

```bash
Usage:
  ./foundryvtt-world-id-reset [flags]

Flags:
  -h, --help            help for foundryvtt-world-id-reset
  -p, --path string     The path to the "worlds/" folder you wish to parse. It should contain the "world.json" file. Defaults to the current directory.
  -v, --verbose         Whether to output verbose logging.
```

Either run the tool from the directory containing the world you wish to reset, or specify the path to the "worlds/" folder using the `-p` flag. You will likely want to run the tool with the `-v` flag to see what it is doing.

You **must stop** the Foundry VTT server _before_ running this tool. You will run into troubles with updating the `*.db` files if the server is running the world you want to reset.

You will be given an opportunity to manually verify the planned changed before the tool writes any changes. Just `CTRL-C` the tool if you wish to cancel.

## Installation

Download the latest release from the [releases page](https://github.com/sneat/foundryvtt-world-id-reset/releases) and extract the binary to a location on your system.
