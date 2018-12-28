# Cronetheus

## Overview

Cronetheus is a tool to schedule your cron jobs. It exposes a prometheus metric for cron jobs that has failed so that you can keep track of them and create alert on them. 

## Installation

```
go get -u github.com/serhatcetinkaya/cronetheus/cmd/cronetheus
```

## Usage

After installing cronetheus you can run it from the command line, it needs sudo permissions to find the UID and GID of the unix user that specified in the config.

```
# cronetheus -config config.yaml -port :9375
```

You can get the metrics from `/metrics`, health status from `/health` and config from `/config` endpoints.

To get more information use `-help`:

```
$ cronetheus -help
Usage of ./cronetheus:
  -alsologtostderr
    	log to standard error as well as files (default true)
  -config string
    	The Cronetheus config file (default "config.yaml")
  -log_backtrace_at value
    	when logging hits line file:N, emit a stack trace
  [...]
```

## Configuration

Cronetheus expects a yaml configuration in the following format:

```yaml
cron_config:
  - cron_id: "cron1"
    descriptor: "@every 3s"
    user: "unix_user1"
    binary: "echo"
    args: "test string > test_file.txt"
  - cron_id: "cron2"
    user: "unix_user2"
    binary: "rm"
    args: "-rf /tmp/test"
    schedule:
      second: "30"
      minute: "*"
      hour: "*"
      day_of_month: "*"
      month: "*"
      day_of_week: "*"
```

To schedule any given cron job you can either specify a descriptor or specify a cron schedule. The minimum resolution for any cron job schedule is 1 second. If both descriptor keyword and schedule is defined for a cron job only descriptor would be used schedule would be omitted. For schedule format please check [CRON Expression Format](https://godoc.org/github.com/robfig/cron#hdr-CRON_Expression_Format). Valid descriptor keywords are `"@yearly", "@annually", "@monthly", "@weekly", "@daily", "@midnight", "@hourly"`.

It allows live configuration reloads with `-HUP` signal:

```
# kill -HUP $PID
```


