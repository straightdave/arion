package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/fatih/color"
)

var (
	fSourceFiles = flag.String("src", "", "source <service>.pb.go and/or <service>_grpc.pb.go files, comma seperating")
	fMockServer  = flag.Bool("mock", false, "whether to generate mock server")
	fOutputDir   = flag.String("o", "", "output directory")
	fGoGetUpdate = flag.Bool("u", false, "force updating dependencies when building Postgal")
	fCrossBuild  = flag.String("cross", "", "Cross-platform building flags. e.g 'GOOS=linux GOARCH=amd64'")
	fVerbose     = flag.Bool("verbose", false, "print verbose information when building postgals")
	fCmdTimeout  = flag.Duration("cmdtimeout", 30*time.Second, "Cmd execution timeout (parsing deps, compiling, etc.)")
	fDebug       = flag.Bool("debug", false, "debug mode")

	green  = color.New(color.FgGreen).SprintfFunc()
	yellow = color.New(color.FgYellow).SprintfFunc()
)

func main() {
	flag.Parse()

	sourceFiles := mustReadSourceFiles(*fSourceFiles)
	outDir := mustCreateOutDir(*fOutputDir, *fMockServer)

	for _, f := range sourceFiles {
		mustCopySource(f.Src, outDir)
	}

	if *fMockServer {
		log.Printf("Generating Mock Server ...")
		mustGenMockServer(sourceFiles, outDir)
	} else {
		log.Println("Generating Postgal ...")
		mustGenProtoTypeFile(sourceFiles, outDir)
		mustRestoreFile(_compressedMain, outDir, "main.go")
		mustRestoreFile(_compressedStatic, outDir, "static.go")
	}

	mustGenModFile(outDir, outDir)
	mustCompileDir(outDir, *fMockServer, *fGoGetUpdate, *fVerbose)
	log.Printf(green("SUCCESS"))
}

func debug(format string, args ...interface{}) {
	if *fDebug {
		fmt.Printf(yellow("[debug] ")+format, args...)
	}
}
