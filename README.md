# duckdns

![Build Status](https://travis-ci.org/invisiblethreat/duckdns.svg?branch=master)

Golang client for updating DNS entries at https://duckdns.org

## Usage

```bash

Usage of ./duckdns:
  -c, --config string       Config file location (default "duckdns.yaml")
  -d, --debug               Use debug mode
  -n, --names stringSlice   Names to update with DuckDNS. Just the subdomain section. Use the flag multiple times to set multiple values.
  -t, --token string        Token for updating DuckDNS
  -l, --log string          Log file path. If unset default to `stderr`
  ```

## Modes

Preference for items is in the following order:

* CLI
* Environmental,
* Configuration File

If items are only partially complete, extra methods will be used to try and complete the needed values. If you only set the token on the CLI the name values
will attempt be be filled from the environment variables, and finally the
configuration file. This is also a useful strategy for overriding items that are
set lower in the order of priority. Use the CLI for your names, and rely on the
token from the configuration file.

### CLI Only

Pass all of the arguments in via CLI

```bash

duckdns -t <your token> -n name1 -n name2

```

### Environment Variables

```bash

export DUCK_TOKEN="<your token>"
export DUCK_NAMES="name1 name2" #use space delimited names
duckdns

```

### Configuration File

`duckdns.yaml`

```yaml

---
token: feedfeed-feed-feed-feed-feedfeedfeed
domains:
  - testdomain
  - test-domain

  ```

```bash

duckdns # uses duckdns.yaml in the same directory as the default
# or
duckdns -c /path/to/duckdns.yaml

```

Note: This does not currently allow for specification of an IP address. The
address that is observed by DuckDNS is what is used.
