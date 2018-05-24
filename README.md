Arion
=======
[![Build Status](https://travis-ci.org/straightdave/arion.svg?branch=master)](https://travis-ci.org/straightdave/arion)

Arion is a gRPC tool to:
- get endpoints definition
- debug endpoints

## Get Arion
```bash
> go get github.com/straightdave/arion
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
When using Arion, your machine should have internet access since Arion will `go get` some official gRPC related packages.

## Use PostGal

### Console Mode
In console mode, PostGal can list simple endpoints of gRPC services defined in your pb.go file.
You can use `-list` simply to see whether your PostGal is ok or not:
```bash
$ ./postgal -list
MyappClient
> Hello
```

Also you can call endpoints in console mode:
```bash
$ ./postgal -call Hello -req '{"Name": "Dave"}'
Message: Hello Dave
```
### Broswer Mode
Browser mode is recommended way to use PostGal.

```bash
$ ./postgal -serve
```
Then you can open browser (by default PostGal hosts on port 9999)

In the PostGal web page, you can easily:
1. see detailed information about endpoints and definitions about each request/response
2. call endpoints
3. change endpoint locations (gRPC server, IP:port)

You can use this page to play with gRPC services as you do the similar thing against HTTP.

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
