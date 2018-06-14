Arion
=======
[![Build Status](https://travis-ci.org/straightdave/arion.svg?branch=master)](https://travis-ci.org/straightdave/arion)

Arion is a gRPC tool to:
- get endpoints definition
- debug endpoints
- do the performance test agains endpoints

## Get Arion
```bash
> go get -u github.com/straightdave/arion
```

## Use Arion to generate PostGal
```bash
$ ./arion -src <your.any.pb.go>
2018/05/24 22:52:46 generating new pb file: temp342770882/pb.go
2018/05/24 22:52:46 generating meta source file: temp342770882/meta.go
2018/05/24 22:52:46 creating new source file: temp342770882/main.go
2018/05/24 22:52:46 creating new source file: temp342770882/static.go
2018/05/24 22:52:46 change dir to ./temp342770882
2018/05/24 22:52:46 install all dependencies...
2018/05/24 22:52:49 change dir back to ...
2018/05/24 22:52:49 SUCCESS
```

Then Arion will generate a temporary folder containing source files and compile those files into a
executable binary called *PostGal*.

*NOTE*
When using Arion, your machine should have internet access since Arion will `go get` some official gRPC related packages including:
* github.com/golang/protobuf/jsonpb
* golang.org/x/net/context
* google.golang.org/grpc

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
      endpoint name (<svc_name>#<end_name> or just <end_name>) to execute or query
  -h string
      hosts of target service (commas to seperate multiple hosts) (default ":8087")
  -i  show info
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

>You can create a JSON as request data based on knowledge you get using `-i -t` or `-i -e`

To execute an performance test against one endpoint:
```bash
$ ./postgal -e Hello -d '{"Name": "Dave"}' -x -rate 10 -duration 10s
Massive Call...
... (report)
```

You can use `-df` to specify a data file which consists of multiple request data to conduct the massive call:
```bash
$ ./postgal -e Hello -df ./myreqs.txt -x -rate 10 -duration 10s
```

>If you don't use the option `-loop` when using a data file, the massive call will stop after all requests are sent once.

### Broswer Mode
Browser mode is recommended way to use PostGal. You can use is just like Postman.

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
