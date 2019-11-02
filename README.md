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
      * [TODO](#todo)
      * [Limitations](#limitations)
      * [Building](#building)
      * [Parameters](#parameters)
      * [Sample Execution](#sample-execution)
      * [Testing](#testing)
         * [Unit Test](#unit-test)
         * [Integration Test](#integration-test)
         * [Concurrency Testing](#concurrency-testing)
      * [Scripting](#scripting)

<!-- Added by: mmc, at: Thu Oct 31 15:43:03 PDT 2019 -->

<!--te-->

## Scope

TFTP Server, written in Go, following the [RFC-1370](https://tools.ietf.org/html/rfc1350) specification.

## TODO

There are several features that I'd like to continue building:

*_TEST: Incomplete Files should not Appear_*
* Expire a Loaded File with TTL
* Implement Timeouts
* ~~Tests with Lots of Clients~~ TESTED with 5/10/20/50
* Date/Time Stamps to Messages
* Progress Indicators for Each Thread like NPM (Pie-in-the-Sky)
* Speed-Up in allocation of byte buffers on WRITE

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
cd src
go test
```

### Integration Test

The following scripts will great two large files, one with a filesize that is even 512 blocks, the other is not. Compare the two MD5 hashs to confirm that the same file generated locally, sent to the tftp-server, then pulled back down is the same.

*Test Single*

_Parameters_
```
~$ ./test-entrypoint.sh <uniqe-id, required> <filesize, default 100000>
```

_Example_
```
~$ cd test
~$ ./test-entrypoint.sh 1 100000

OK #1: Perfect Match
```

### Concurrency Testing

This testing script will spawn off X number of calls to "./test-entrypoint.sh"

_Parameters_
```
./test.sh <# of concurrent clients>
```

_Example_
```
~$ cd test
~$ ./test.sh 5
Spawning 1
Spawning 2
Spawning 3
Spawning 4
Spawning 5

OK #1: Perfect Match
OK #2: Perfect Match
OK #3: Perfect Match
OK #4: Perfect Match
OK #5: Perfect Match
```



## Scripting

This software supports exit-codes for errors resulting in abnormal execution:

| Code | Desc |
| ---- | ---- |
| 0    | No Error |
| 1    | Listener Error: IP |
| 2    | Listener Error: Port |
