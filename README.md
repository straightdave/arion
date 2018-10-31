<h1 align="center">
	<br>
	<br>
	<img width="320" src="media/logo.svg" alt="Arion">
	<br>
	<br>
	<br>
</h1>

[![Build Status](https://travis-ci.org/straightdave/arion.svg?branch=master)](https://travis-ci.org/straightdave/arion)

Arion is a tool to test gRPC service. Arion is able to:
- Get information of gRPC services / endpoints / data entities
- Debug gRPC endpoints like `curl` does for HTTP
- Do performance test for gRPC services

## Get Arion
```bash
$ go get -u github.com/straightdave/arion
```

> `-u` option would upgrade Arion if already installed.

## Basic usage of Arion

```
$ arion -h
Usage of arion:
  -c  to clear temp folder after Postgal is generated
      *NOTE*: use -o to generate Postgal out of temp folder
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
Usage of postgal:
  -at string
      address to host PostGal in browser mode (default ":9999")
  -d string
      request data
  -debug
      print some debug info (for dev purpose)
  -df string
      request data file
      Data in the file will be read line by line
  -dumpto string
      dump massive call responses to file
  -duration duration
      execution duration like 10s, 20m (default 10s)
  -e string
      endpoint name (<svc_name>#<end_name> or just <end_name>) to execute
  -h string
      hosts of target service (commas to seperate multiple hosts) (default ":8087")
  -i  show info
  -json
      response in JSON format
  -loop
      repeat all requests in loops
  -rate uint
      execution frequency per second (default 1)
  -serve
      use PostGal in browser mode
  -t string
      data type name
  -v  print version info
  -worker uint
      workers (concurrent goroutines) (default 10)
  -x  massive call endpoint (needs -rate and -duration)
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

### Call endpoints

```
$ ./postgal -e Hello -d '{"Name": "Dave"}' -h 192.168.0.1:8087
Message: Hello Dave
```

> **Tips**
> * `-e <endpoint name>`, case sensitive.
> * `-d <data in JSON format>`
> * If `-h` option is not given, postgal uses '0.0.0.0:8087' by default.
> * You can compose the JSON request data based on the knowledge you get by using `-i -t` or `-i -e`
> * If the type of request object is `protobuf.Empty`, the data given by `-d` option would be ignored

Using `-json` to format response data in JSON:
```
$ ./postgal -e Hello -d '{"Name": "Dave"}' -json
{
    "Message": "Hello Dave"
}
```

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

#### 2. data file
You can use `-df` to specify a data file which consists of multiple request data (each in one line) to conduct the massive call:
```
$ ./postgal -e Hello -df ./myreqs.txt -x -rate 10 -duration 30s -loop
```
>If you don't use the option `-loop` when using a data file, the massive call will stop after all request data in the file are sent once.

#### 3. worker number
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

#### 5. dump responses
To dump responses to a file:
```
$ ./postgal -e Hello -d '{"Name":"daveeeeee"}' -x -dumpto responses.dump
```
So _postgal_ would serialize all successful responses into JSON format and write them to the file, line by line.

#### 6. auto-generated value
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
> **Experimental**

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
