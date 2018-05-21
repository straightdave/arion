arion
=======

## Get Arion
```bash
> go get github.com/straightdave/arion
```

## Use Arion to generate PostGal
```bash
> ./arion -src <your.any.pb.go>
```
>In this process, after all temporary source code are generated,
Arion will try to call local `go` to compile.

go to temporary dir and use the binary `postgal`:
```bash
# list endpoints
tmpDir> ./postgal -list

# invoke
tmpDir> ./postgal -call SomeEndpoint -req '{"field": 123}'
```
or start in broswer mode:
```bash
tmpDir> ./postgal -serve
```
Then you can open browsers to use PostGal.

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
