arion
=======

## Get Arion
```bash
> go get github.com/straightdave/arion
```

## Use Arion to generate XClient
```bash
> arion -src <your.any.pb.go>
```
>In this process, after all temporary source code are generated,
Arion will try to call local `go` to compile.


go to temporary dir and use the binary `xclient`:
```bash
# list endpoints
tmpDir> ./xclient -list

# invoke
tmpDir> ./xclient -call SomeEndpoint -req '{"field": 123}'
```
or start as web page:
```bash
tmpDir> ./xclient -serve ":8888"
```

## Develop

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
