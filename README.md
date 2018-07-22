<h1 align="center">
	<br>
	<br>
	<img width="320" src="media/logo.svg" alt="Arion">
	<br>
	<br>
	<br>
</h1>

[![Build Status](https://travis-ci.org/straightdave/arion.svg?branch=master)](https://travis-ci.org/straightdave/arion)

Arion is a gRPC tool to:
- get service information
- debug endpoints
- do the performance test agains endpoints

## Get Arion
```bash
$ go get -u github.com/straightdave/arion
```

## Usage of Arion
```bash
$ ./arion -h
Usage of ./arion:
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

## Use Arion to generate PostGal
```bash
$ ./arion -src <your.any.pb.go> -u
2018/05/24 22:52:46 generating new pb file: temp342770882/pb.go
2018/05/24 22:52:46 generating meta source file: temp342770882/meta.go
2018/05/24 22:52:46 creating new source file: temp342770882/main.go
2018/05/24 22:52:46 creating new source file: temp342770882/static.go
2018/05/24 22:52:46 change dir to ./temp342770882
2018/05/24 22:52:46 force-update all dependencies...
2018/05/24 22:52:49 change dir back to ...
2018/05/24 22:52:49 SUCCESS
```

> Using `-u` to force-update local dependencies. It's required (once) if some underlying packages are not up-to-date.

Then by default Arion will generate a temporary folder containing source files and compile those files into an executable binary called *Postgal*. You can use `-o` to specify other path/name for this executable file. Also you can use `-c` to clear temp folder after *Postgal* is generated.

*NOTE*
When using Arion, your machine should have internet access since Arion will `go get` some official gRPC related packages including:
* github.com/golang/protobuf/jsonpb
* golang.org/x/net/context
* google.golang.org/grpc

Besides, some packages provide code analyzing and performance testing:
* github.com/straightdave/lesphina
* github.com/straightdave/trunks
>If you find strange panic when using Arion or Postgals, you can manually update those packages by `go get -u -f <package>` and re-generate Postgals to see if this can solve the problems or not.

### Cross-build
If you want Arion to build a postgal working on another platform (e.g linux/amd64), you can use `-cross` flag:
```
$ ./arion -src <your.any.pb.go> -cross 'GOOS=linux GOARCH=amd64'
```
> There are only several valid pairs like linux/amd64, etc.

## List *Postgals*
```bash
$ ./arion -l
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
> At the moment Arion looks for Postgals in current and all `./temp*` folders

## Use PostGal

Usage:
```
  -at string
      address to host PostGal in browser mode (default ":9999")
  -d string
      request data
  -debug
      print some debug info (for dev purpose)
  -df string
      request data file
      Data in the file will be read line by line
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

### Console Mode
In console mode, PostGal can list simple endpoints of gRPC services defined in your pb.go file.
You can use `-i` simply to see whether your PostGal is ok or not:
```bash
$ ./postgal -i
Myapp
> Hello
```

To see some details about one endpoint:
```bash
$ ./postgal -i -e Hello
Myapp#Hello
- Request entity:
--- Name string (json field name: Name)
- Response entity:
--- Message string (json field name: Message)
```

To see some details about one entity structure:
```bash
$ ./postgal -i -t HelloRequest
- Name string (json field name: Name)
```

Also you can call one endpoint:
```bash
$ ./postgal -e Hello -d '{"Name": "Dave"}'
Message: Hello Dave
```
> **NOTE**
> - If `-h` option is not specified, postgal uses '0.0.0.0:8087' by default.
> - You can compose the JSON request data based on the knowledge you get by using `-i -t` or `-i -e`
>If the type of request object is `protobuf.Empty`, the data given by `-d` option would be ignored


Using `-json` to have JSON format output:
```bash
$ ./postgal -e Hello -d '{"Name": "Dave"}' -json
{
    "Message": "Hello Dave"
}
```

### Performance test with Postgal

To execute performance tests against one endpoint:
```bash
$ ./postgal -e Hello -d '{"Name": "Dave"}' -x -rate 10 -duration 10s
Massive Call...
... (report)
```

You can use `-df` to specify a data file which consists of multiple request data to conduct the massive call:
```bash
$ ./postgal -e Hello -df ./myreqs.txt -x -rate 10 -duration 30s -loop
```

>If you don't use the option `-loop` when using a data file, the massive call will stop after all requests are sent once.

To specify number of workers (maximun concurrent goroutines; default is 10) in performance testing:
```bash
$ ./postgal -e Hello -d '{"Name":"dave"}' -x -rate 10 -duration 30s -worker 16
```

You can specify a list of IP addresses to execute preformance test against a cluster:
```bash
$ ./postgal -h <address>,<address> -e <endpoint> -d 'test data' -x
```

#### auto-generated value
Currently Postgal supports generating unique values when doing the performance test:
```bash
$ ./postgal -e Hello -d '{"Name":"Ultraman-<<arion:unique:string>>"}' -x -rate 5 -duration 10s
```
It will generate requests like:
```
{"Name":"Ultraman-0"}
{"Name":"Ultraman-1"}
...
{"Name":"Ultraman-49"}
```
> __NOTE__ for now it's an experimental feature and being improved

### Browser Mode
Browser mode is the graphic way to use Postgals. You can use is just like Postman.

```bash
$ ./postgal -serve
```
Then you can open browser (by default PostGal hosts on port 9999)

In the PostGal web page, you can easily:
1. see detailed information about endpoints and definitions about each request/response
2. call endpoints
3. change endpoint locations (gRPC server, IP:port)

You can use this page to play with gRPC services as you do the similar thing against HTTP with Postman.


## Development

### build
```bash
> go run ./_build/build.go
```
or
```bash
> ./build.sh
```

### start dev server

```bash
> ./dev.sh
```

>The developing web server will serve all static files in `web/` folder.
This is helpful for developing web pages before compiling them into binary.

## License

MIT © [Dave Wu](https://github.com/straightdave)
