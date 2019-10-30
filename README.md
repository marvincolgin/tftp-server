# TFTP Server

```
Author:

Vin Colgin
mcolgin@gmail.com
https://github.com/marvincolgin
https://www.linkedin.com/in/mcolgin
https://marvincolgin.com
```


## Scope

TFTP Server, written in Go, following the [RFC-1370](https://tools.ietf.org/html/rfc1350) specification.

## Limitations

* Does not support later RFC specifications
* Only octet aka "binary" mode is supported

## Building

```
go get github.com/pborman/getopt
go build -o <path to resulting binary>
```

## Parameters / Usage

Parameters follow posix standards using the Google [getopt](https://godoc.org/github.com/pborman/getopt) go package.

*TABLE: Command-line Parameters*

| param | desc | default |
| ----- | ---- | ------- |
| help  | Help for Parameters | |
| ip    | IP Address for Listener | 127.0.0.1 |
| port  | Port for Listener | 69 |
| threads | Number of Threads | 16 |
| timeout | Seconds for Timeout | 1 |

*Example*

Help on Parameters
```
tftp --help
```

Start Listener on 192.168.0.1:6969
```
tftp --ip 192.168.0.1 --port 6969
```

Restrict Listener to 4 threads
```
tftp --threads 4
```

## Testing

```
go test
```
