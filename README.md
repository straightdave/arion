Arion
=======
[![Build Status](https://travis-ci.org/straightdave/arion.svg?branch=master)](https://travis-ci.org/straightdave/arion)

Arion is a gRPC tool to:
- get endpoints definition
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
  -l  list Postgals in current folder or all ./temp* folders
  -o string
      output executable binary file (default "postgal")
  -src string
      source pb.go file
  -u  update dependencies when building Postgal
```

## Use Arion to generate PostGal
```bash
$ ./arion -src <your.any.pb.go> -u
2018/05/24 22:52:46 generating new pb file: temp342770882/pb.go
2018/05/24 22:52:46 generating meta source file: temp342770882/meta.go
2018/05/24 22:52:46 creating new source file: temp342770882/main.go
2018/05/24 22:52:46 creating new source file: temp342770882/static.go
2018/05/24 22:52:46 change dir to ./temp342770882
2018/05/24 22:52:46 install all dependencies...
2018/05/24 22:52:49 change dir back to ...
2018/05/24 22:52:49 SUCCESS
```

> Using `-u` to force-update local dependencies. It's required to use once if some underlying packages are not up-to-date.

Then by default Arion will generate a temporary folder containing source files and compile those files into an executable binary called *Postgal*. You can use `-o` to specify other path/name for this executable file. Also you can use `-c` to clear temp folder after *Postgal* is generated.

> At the moment Arion looks for Postgals in current and all `./temp*` folders

*NOTE*
When using Arion, your machine should have internet access since Arion will `go get` some official gRPC related packages including:
* github.com/golang/protobuf/jsonpb
* golang.org/x/net/context
* google.golang.org/grpc

Besides, some packages supporting code analyzing and performance testing:
* github.com/straightdave/lesphina
* github.com/straightdave/trunks
>If you find strange panic when using Arion or Postgals, you can manually update those packages by `go get -u -f <package>` and re-generate Postgals to see if this can solve the problems or not.

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

Using `-json` to have JSON format output:
```bash
$ ./postgal -e Hello -d '{"Name": "Dave"}' -json
{
    "Message": "Hello Dave"
}
```

>You can compose the JSON request data based on the knowledge you get by using `-i -t` or `-i -e`

To execute an performance test against one endpoint:
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
