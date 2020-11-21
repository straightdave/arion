package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"text/template"

	gozip "github.com/straightdave/gozip/lib"
	"github.com/straightdave/lesphina/v2"
)

type tplDataForMockServer struct {
	PBRegisterServerFunc string
	PBServerInterface    string
	Methods              []tplMethod
}

type tplMethod struct {
	Name         string
	RequestType  string
	ResponseType string
}

func mustGenMockServer(files []*SourceFile, dir string) {
	if err := genMockServer(files, dir); err != nil {
		log.Fatal(err)
	}
}

func genMockServer(files []*SourceFile, dir string) error {
	les := mergeLesphina(files)

	serverInterfaces := les.Query().ByKind(lesphina.KindInterface).ByName("~Server").All()
	if len(serverInterfaces) > 1 {
		log.Printf("Warning: more than 1 '~Server' interfaces found. Currently Arion only supports 1. So use the first.")
	}
	serverInterface, ok := serverInterfaces[0].(*lesphina.Interface)
	if !ok {
		return fmt.Errorf("failed to convert interface data")
	}
	log.Printf("Read server interface: %s\n", serverInterface.GetName())

	var methods []tplMethod
	for _, m := range serverInterface.Methods {
		// only collect unary method (not supporting streaming for now).
		// unary method will use pattern: <MethodName>(context.Context, *pb.XXXXRequest) (*pb.XXXXResponse, error)
		// (req/resp type name may vary, not with Request/Response suffix),
		// So, skip if input params are more than 2, and the first in-param is not 'context'.
		if len(m.InParams()) != 2 || len(m.OutParams()) != 2 {
			continue
		}
		if !strings.Contains(m.InParams()[0].BaseType, "context") {
			continue
		}

		methods = append(methods, tplMethod{
			Name:         m.Name,
			RequestType:  m.InParams()[1].BaseType, // excluding '*'
			ResponseType: m.OutParams()[0].BaseType,
		})
	}

	var regSvrFunc string
	regSvrFuncs := les.Query().ByKind(lesphina.KindFunction).ByName("Register~").All()
	for _, f := range regSvrFuncs {
		if strings.HasSuffix(f.GetName(), "Server") {
			regSvrFunc = f.GetName()
			break
		}
	}
	if regSvrFunc == "" {
		return fmt.Errorf("cannot find 'Register~Server' function in source file")
	}

	tplData := &tplDataForMockServer{
		PBRegisterServerFunc: regSvrFunc,
		PBServerInterface:    serverInterface.GetName(),
		Methods:              methods,
	}

	log.Printf("[debug] tpl data: %+v", tplData)

	err := genMockServerSource(_compressedMockMain, "mockmain", dir, "main.go", tplData)
	if err != nil {
		log.Printf("failed to generate mock main.go: %v", err)
		return err
	}

	err = genMockServerSource(_compressedMockGRPCServer, "mockgrpc", dir, "grpc_server.go", tplData)
	if err != nil {
		log.Printf("failed to generate mock grpc_server.go: %v", err)
		return err
	}

	err = restoreFile(_compressedMockHTTPServer, dir, "http_server.go")
	if err != nil {
		log.Printf("failed to restore mock http_server.go: %v", err)
		return err
	}

	return nil
}

func genMockServerSource(raw, tplName, dirName, fileName string, tplData interface{}) error {
	tpl, err := template.New(tplName).Parse(gozip.DecompressString(raw))
	if err != nil {
		return err
	}
	tf, err := os.Create(path.Join(dirName, fileName))
	if err != nil {
		return err
	}
	defer tf.Close()
	return tpl.Execute(tf, tplData)
}
