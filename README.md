# TFTP Server

```
Author:

Vin Colgin
mcolgin@gmail.com
https://github.com/marvincolgin
https://www.linkedin.com/in/mcolgin
https://marvincolgin.com
```

<!--ts-->
   * [TFTP Server](#tftp-server)
      * [Scope](#scope)
      * [Limitations](#limitations)
      * [Building](#building)
      * [Parameters](#parameters)
      * [Sample Execution](#sample-execution)
      * [Testing](#testing)
         * [Unit Test](#unit-test)
         * [Integration Test](#integration-test)
      * [Scripting](#scripting)

<!-- Added by: mmc, at: Wed Oct 30 21:46:55 PDT 2019 -->

<!--te-->

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

## Parameters

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

## Sample Execution
```
~$ tftp

Listener: 127.0.0.1:69
Threads: 16 Starting...Done
Listener: Loop Running
READ: REQUEST file:[test-even.dat], client:[127.0.0.1:61073]
READ: SUCCESS file:[test-even.dat], client:[127.0.0.1:61073]
READ: REQUEST file:[test-odd.dat], client:[127.0.0.1:61075]
READ: SUCCESS file:[test-odd.dat], client:[127.0.0.1:61075]
```

## Testing

### Unit Test
```
go test
```

### Integration Test

The following scripts will great two large files, one with a filesize that is even 512 blocks, the other is not. Compare the two MD5 hashs to confirm that the same file generated locally, sent to the tftp-server, then pulled back down is the same.

*Example*
```
cd test
./test.sh

ae2797df6804b7653d697f9a869e4007  test-even.dat
37d0ee31bb8827f75c1c69dca888d83c  test-odd.dat
ae2797df6804b7653d697f9a869e4007  test-even.dat
37d0ee31bb8827f75c1c69dca888d83c  test-odd.dat
```

## Scripting

This software supports exit-codes for errors resulting in abnormal execution:

| Code | Desc |
| ---- | ---- |
| 0    | No Error |
| 1    | Listener Error: IP |
| 2    | Listener Error: Port |
