<h1 align="center">
	<br>
	<br>
	<img width="320" src="media/logo.svg" alt="Arion">
	<br>
	<br>
	<br>
</h1>

[![Build Status](https://travis-ci.org/straightdave/arion.svg?branch=master)](https://travis-ci.org/straightdave/arion)

Arion is a tool to help develop and test gRPC service.
Arion can generate test clients (called **postgal** by default) based on `*.pb.go` files.

Main features (of postgal) include:
- Get information of gRPC services / endpoints / data types
- Debug gRPC endpoints as `curl` or Postman(TM) does for HTTP
- Do stress test on gRPC endpoints
- Support both Unary and Streaming

> **NOTE**
> * As of Nov.28, 2018, Arion supports client-side/server-side streaming (call & stress); for bi-directional streaming, only call is supported. Stress test for bi-directional streaming will be supported soon.
> * After this large scale code updating, it would be a refactory for the code which becomes more or less a mess. :)
> * Any questions please feel free to open a new issue ~

## Get Arion
```
$ go get -u github.com/straightdave/arion
```

> `-u` option would upgrade Arion if already installed.

## Basic usage of Arion

```
$ arion -h
Usage of arion:
  -cross string
      Cross-platform building flags. e.g 'GOOS=linux GOARCH=amd64'
  -l  list Postgals in current folder or all ./temp* folders
  -o string
      output executable binary file
  -src string
      source pb.go file
  -u  update dependencies when building Postgal
  -verbose
      print verbose information when building postgals
```

## Generate _postgal_
```
$ arion -src <your.any.pb.go>
arion -src myapp.pb.go
2018/10/31 21:50:50 generating new pb file: temp-myapp-829300873/pb.go
2018/10/31 21:50:50 generating meta source file: temp-myapp-829300873/meta.go
2018/10/31 21:50:50 creating new source file: temp-myapp-829300873/main.go
2018/10/31 21:50:50 creating new source file: temp-myapp-829300873/static.go
2018/10/31 21:50:50 change dir to /Users/straightdave/arion_files/temp-myapp-829300873
2018/10/31 21:50:50 Analyzing dependencies ...
2018/10/31 21:50:51 Install dependencies ...
complete
2018/10/31 21:50:51 Build ...
complete
2018/10/31 21:50:53 change dir back to /Users/straightdave/arion_files
2018/10/31 21:50:53 SUCCESS
```

> **Tips**
> * Using `-u` option to force-update local dependencies. It's required (once) if some underlying packages are not up-to-date.
> * Using `-verbose` option to have more output during execution of command.
> * Using `-o <custom-name>` to call the binary a _custom-name_ rather than _postgal_.

Then by default Arion will generate a temporary folder containing source files and compile those files into an executable binary called **postgal** in it.

### Cross-build
If you want Arion to build the _postgal_ working on another platform (e.g linux/amd64), you can use `-cross` option:

```
$ arion -src <your.any.pb.go> -cross 'GOOS=linux GOARCH=amd64'
```
> **Tips**
> * There are only several valid pairs like linux/amd64, etc.
> * Experimental. Sometimes it fails :smile:. In that case, you can go to source code folder and do the cross-build yourself.

## List *postgals*

```
$ arion -l
[-] ./postgal-xxx
Service:  Myapp
Generated:  Wed Jun 20 23:40:59 CST 2018
Checksum: 5de493383a0ec6ad79a7a655ae8aecbf

[-] temp080184789/postgal
Service:  Myapp
Generated:  Wed Jun 20 23:40:39 CST 2018
Checksum: 5de493383a0ec6ad79a7a655ae8aecbf

[-] temp252099608/postgal
Service:  Myapp
Generated:  Wed Jun 20 23:40:51 CST 2018
Checksum: 5de493383a0ec6ad79a7a655ae8aecbf

```
> Arion looks for _postgal_ in current and all `./temp*` folders

## Usage of `postgal`

```
Usage of ./postgal:
  -B string
    	protobuf binary data file
  -G string
    	the bin file name to generate
  -N uint
    	how many concurrent streaming connections when doing stress test mode (-x) (default 1)
  -at string
    	address to host Postgal in browser mode (default ":9999")
  -d string
    	request data string
  -debug
    	print some debug info (for dev purpose)
  -df string
    	request data file. Data in the file will be sent line by line
  -dumpto string
    	dump massive call responses to file
  -duration duration
    	test duration in stress test mode: 10s, 20m (default 10s)
  -e string
    	endpoint name (<svc_name>#<end_name> or just <end_name>) to execute
  -h string
    	hosts of target service (commas to seperate multiple hosts) (default ":8087")
  -i	print basic service info
  -json
    	print response in JSON format
  -loop
    	repeatly sending all requests in data file (-df)
  -m string
    	unary | client | server | bidirect (default "unary")
  -meta string
    	gRPC metadata (format: 'key1=value1 key2=value2')
  -n uint
    	how many times to send streaming msg to server in one connection (default 10)
  -rate uint
    	expected QPS (query per second) in stress test mode (default 1)
  -serve
    	browser mode
  -t string
    	data type name
  -v	print version info
  -worker uint
    	workers (max volume of goroutine pool) (default 10)
  -x	stress test mode
```

**TL;DR**, let's see some examples below.

## Console Mode
To use _postgal_ in command line.

### Read endpoint list

_postgal_ can list simple endpoints of gRPC services defined in your pb.go file.
You can use `-i` simply to see whether your PostGal is ok or not:
```
$ ./postgal -i
Myapp
> Hello
```

### Read endpoint info

```
$ ./postgal -i -e Hello
Myapp#Hello
- Request entity:
--- Name string (json field name: Name)
- Response entity:
--- Message string (json field name: Message)
```

### Read data entity details
For instance, to get the type definition of entity _HelloRequest_:

```
$ ./postgal -i -t HelloRequest
- Name string (json field name: Name)
```

### Generate binary data file from JSON
```
$ ./postgal -e RouteGuide#RecordRoute -d '{"latitude":123,"longitude":123}' -G ddd.dat
[8 123 16 123]
```

This is one handy tool currently for testing purpose.

### Using binary data file
```
$ ./postgal -e RouteGuide#RecordRoute -B ddd.dat -h 0.0.0.0:10000 -m client -n 10
point_count:10
```

### Call endpoints

First, a simple unary call:

```
$ ./postgal -e Hello -d '{"Name": "Dave"}'
Message: Hello Dave
```

It calls a gRPC endpoint named as 'hello' on local machine with plain text data (a JSON string).

> **Quick Tips**
> * `-e <endpoint name>`, case sensitive
> * `-d <data in JSON format>` provides request data from plain text in JSON format
> * `-B <bin data file>` reads request data from binary file
> * If both `-d` and `-B` are given, `-d` will take effect
> * If `-h` option is not given, postgal assumes the service is running at '0.0.0.0:8087' by default.
> * You can compose the JSON request data based on the knowledge you get by using `-i -t` or `-i -e`
> * If the type of request object is `protobuf.Empty`, the data given by `-d` option would be ignored

#### prettier response

Using `-json` to format response data in JSON:

```
$ ./postgal -e Hello -d '{"Name": "Dave"}' -json
{
    "Message": "Hello Dave"
}
```

#### gRPC metadata or headers

You can use `-meta` to add metadata to the gRPC call:

```
$ ./postgal -e Hello -d '{"Name":"dave"}' -meta 'k1=v1 k2=v2 k1=v3'
```

#### [experimental] Call via streaming

Now calling via streaming is supported by Arion/Postgal.
You can use `-m` to provide a stream mode for postgal when calling endpoints:

For example, a client-side streaming call:

```
$ ./postgal -e RouteGuide#RecordRoute -d '{"latitude":123, "longitude":123}' -m client -n 10
point_count:10
```

`-m client` indicates it's a client-side streaming call; `-n` (only works in this mode) means how many times to send the request data into the stream (to server).
Then if the server has a response for this streamed requests, _postgal_ would print that.

The possible value for `-m` could be one of followings (NOT case-sensitive):
* `unary` : unary call (non-streaming; by default)
* `client` : client-side streaming
* `server` : server-side streaming
* `bidirect` : bidirectional streaming

For server-side streaming, you can use `-m server`:

```
$ ./postgal -e RouteGuide#ListFeatures -d '{"lo":{"latitude":1,"longitude":-900000000},"hi":{"latitude":923123123,"longitude":923123123}}' -h 0.0.0.0:10000 -m server
name:"Patriots Path, Mendham, NJ 07945, USA" location:<latitude:407838351 longitude:-746143763 >
name:"101 New Jersey 10, Whippany, NJ 07981, USA" location:<latitude:408122808 longitude:-743999179 >
name:"U.S. 6, Shohola, PA 18458, USA" location:<latitude:413628156 longitude:-749015468 >
name:"5 Conners Road, Kingston, NY 12401, USA" location:<latitude:419999544 longitude:-740371136 >
... < omit > ...
name:"3387 Richmond Terrace, Staten Island, NY 10303, USA" location:<latitude:406411633 longitude:-741722051 >
name:"261 Van Sickle Road, Goshen, NY 10924, USA" location:<latitude:413069058 longitude:-744597778 >
location:<latitude:418465462 longitude:-746859398 >
location:<latitude:411733222 longitude:-744228360 >
name:"3 Hasta Way, Newton, NJ 07860, USA" location:<latitude:410248224 longitude:-747127767 >
```

In this sample, you send a piece of data and server returns a lot to you (in a stream 'til the end).

And a bidirectional example:

```
$ ./postgal -e RouteGuide#RouteChat -d '{"location":{"latitude":1,"longitude":-900000000},"message":"hello"}' -h 0.0.0.0:10000 -m bidirect
finished sending
0 location:<latitude:1 longitude:-900000000 > message:"hello"
1 location:<latitude:1 longitude:-900000000 > message:"hello"
2 location:<latitude:1 longitude:-900000000 > message:"hello"
3 location:<latitude:1 longitude:-900000000 > message:"hello"
4 location:<latitude:1 longitude:-900000000 > message:"hello"
5 location:<latitude:1 longitude:-900000000 > message:"hello"
6 location:<latitude:1 longitude:-900000000 > message:"hello"
... < omit > ...

63 location:<latitude:1 longitude:-900000000 > message:"hello"
64 location:<latitude:1 longitude:-900000000 > message:"hello"
read done.
```

Currently Postal only sends one message in bidirectional streaming. Will improve more complexe logic in the future.

> Here it uses google.golang.org/grpc/examples/route_guide as the streaming server.

### Performance test

#### 1. to execute performance tests against one endpoint:
```
$ ./postgal -e Hello -d '{"Name": "Dave"}' -x -rate 100 -duration 60s
Massive Call...
... (report)
```

> **NOTE**
> * Again, if omit `-h` option, the default target host is `0.0.0.0:8087`
> * If only `-x` is specified, it implies `-rate 1 -duration 10s` by default
> * `-meta` is still available in massive gRPC call; However we only use one set of metadata for all requests. If needed, I'll add this functionality to call with multiple metadata sets.

#### 2. data file (only used in `-x` mode)
You can use `-df` to specify a data file which consists of multiple request data (each in one line) to conduct the massive call:
```
$ ./postgal -e Hello -df ./myreqs.txt -x -rate 10 -duration 30s -loop
```
>If you don't use the option `-loop` when using a data file, the massive call will stop after all request data in the file are sent once.

#### 3. worker number (only used in `-x` mode)
Using `-worker` to specify a number of worker pool size (maximum concurrent goroutines; default is 10) in performance testing:
```
$ ./postgal -e Hello -d '{"Name":"dave"}' -x -rate 10 -duration 30s -worker 16
```

#### 4. multiple hosts
Also, you can specify a list of IP addresses to execute preformance test against a cluster.
In this case, `postgal` would send requests to those hosts in a simple (client side) round-robin way.

```
$ ./postgal -h 192.168.0.1:8087,192.168.0.2:8087 -e Hello -d '{"Name":"dave"}' -x
```

#### 5. dump responses (only used in `-x` mode / unary)
To dump responses to a file:
```
$ ./postgal -e Hello -d '{"Name":"daveeeeee"}' -x -dumpto responses.dump
```
So _postgal_ would serialize all successful responses into JSON format and write them to the file, line by line.

#### [experimental] 6. auto-generated value (`-x` mode and unary only)
Currently _postgal_ supports to generate unique data per request:
```
$ ./postgal -e Hello -d '{"Name":"Ultraman-<<arion:unique:string>>"}' -x -rate 5 -duration 10s
```

It will generate requests like:
```
{"Name":"Ultraman-0"}
{"Name":"Ultraman-1"}
...
{"Name":"Ultraman-49"}
```

#### [experimental] 7. stress testing for non-unary endpoints (streaming)

Similar to the 'call' section, we need to provide `-m` option to indicate in which kind of mode to call the endpoints. Currently we only support some simple logic to do the stress test against streaming endpoints, mainly focus on concurrent connections (`-N`) as workload. No metrics (e.g. latencies) will be measured for now.

For example, a client-side streaming one, you can use `-N` to set concurrent connections and use `-duration` to indicate how long will the test lasts:

```
$ ./postgal -e RouteGuide#RecordRoute -d '{"latitude":123, "longitude":123}' -x -duration 10s -m client -N 5 -h 0.0.0.0:10000
Massive Call on RouteGuide#RecordRoute ...
[# 3] response: point_count:302204 elapsed_time:10
[# 4] response: point_count:301157 elapsed_time:10
[# 2] response: point_count:304024 elapsed_time:10
[# 0] response: point_count:298563 elapsed_time:10
[# 1] response: point_count:303236 elapsed_time:10
```

`-N` is the concurrent connection number. `-duration` (same used as in unary stress test) can be used here to indicate the stress test duration (by default 10s). During the time, client will keep sending the data to the server.

For server-side streaming, we also use `-N` and `-duration`, similar to client-side example above (except the tons of response messages to output).

Stress test for Bi-directional streaming has not yet been supported so far.

> **NOTE**
> * Currently no metrics are supported.

## Browser Mode
Browser mode is the graphic way to use _postgal_. Like using Postman(TM).

```
$ ./postgal -serve
```
Then open the browser and navigate to `http://localhost:9999`.

In the _postgal_ browser mode, you can easily:
1. Read details about the service, requests and responses
2. Call endpoints

## Development

### build
If snippets (templates) or web pages are updated, you have to run:
```
$ go run ./_build/build.go
```
or
```
$ ./build.sh
```
to update `main2.go`.

### start dev server
```
$ ./dev.sh
```
The developing web server will serve all static files in `web/` folder.
This is helpful for developing web pages before compiling them into binary.

## License

MIT © [Dave Wu](https://github.com/straightdave)
