# magic4linux

Allows you to use the magic remote on your webOS LG TV as a keyboard/mouse for your ~~PC~~ Linux machine.

This is a Linux implementation of the [Wouterdek/magic4pc](https://github.com/Wouterdek/magic4pc) client.

A virtual keyboard and mouse is created via the `/dev/uinput` interface, as provided by the [bendahl/uinput](https://github.com/bendahl/uinput) library. For non-root usage, please add udev rules as instructed in the [`uinput`](https://github.com/bendahl/uinput#uinput-----) documentation.

## Installation

```shell
go install github.com/mafredri/magic4linux/cmd/magic4linux
```

## Usage

There are no options yet.

```shell
magic4linux
```

## Building for other platforms

```shell
git clone https://github.com/mafredri/magic4linux
cd magic4linux/cmd/magic4linux
GOOS=linux GOARCH=arm64 go build
```
