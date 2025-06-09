# go-ogn-client

This repository provides a minimal Go implementation of the
[python-ogn-client](https://github.com/glidernet/python-ogn-client)
library.  It contains a small APRS client, a very lightweight APRS
parser and a helper to download the OGN device database.

The implementation focuses on the basic functionality:

* `client` – connect to an OGN APRS server and read raw packets
* `parser` – parse position reports into Go structs
* `ddb` – download and decode the OGN device database

The code aims to be simple and self contained so it can serve as a
starting point for further development.
