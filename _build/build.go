// hand-made build tool for this project:
// - do rendering work while compiling
// - serve pages while developing
// Basically this is a dev tool for Arion.

package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	gozip "github.com/straightdave/gozip/lib"
)

const (
	main2FileName = "z_main2.go"
)

var (
	dev       = flag.Bool("dev", false, "to serve static pages")
	port      = flag.String("port", "9999", "local port to serve")
	targetDir = flag.String("target", ".", "target dir")
)

func main() {
	flag.Parse()

	if *targetDir == "" {
		log.Fatalf("Build Arion: Lack of target dir (-target)")
	}

	if *dev {
		rootDir := filepath.Join(*targetDir, "web")
		serveStaticPage(rootDir)
		return
	}

	build(*targetDir)
}

func serveStaticPage(rootDir string) {
	http.Handle("/", http.FileServer(http.Dir(rootDir)))
	p := ":" + strings.TrimLeft(*port, ":")
	log.Printf("Build Arion: Serving dev at localhost at %s", p)
	http.ListenAndServe(p, nil)
}

func build(targetDir string) {
	if !canBuild() {
		log.Fatalf("Build Arion: Cannot find go toolchain. Please check Golang is installed and set to $PATH.")
	}

	generateMain2ForArion(targetDir, struct {
		ModFileContent,

		// for postgal
		CompressedMeta,
		CompressedMain,
		CompressedStatic,

		// for mock server
		CompressedMockMain,
		CompressedMockHTTPServer,
		CompressedMockGRPCServer string
	}{
		ModFileContent:           compressFileContent(filepath.Join(targetDir, "/templates/go.mod.tpl")),
		CompressedMeta:           compressFileContent(filepath.Join(targetDir, "/templates/postgal/meta.go.tpl")),
		CompressedMain:           compressFileContent(filepath.Join(targetDir, "/templates/postgal/main.go.tpl")),
		CompressedStatic:         getCompressedStatic(filepath.Join(targetDir, "/templates/postgal/web.go.tpl")),
		CompressedMockMain:       compressFileContent(filepath.Join(targetDir, "/templates/mock/main.go.tpl")),
		CompressedMockHTTPServer: compressFileContent(filepath.Join(targetDir, "/templates/mock/http_server.go.tpl")),
		CompressedMockGRPCServer: compressFileContent(filepath.Join(targetDir, "/templates/mock/grpc_server.go.tpl")),
	})

	log.Printf("Start building ...")
	output, err := exec.Command("go", "build").CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	if len(output) > 0 {
		log.Printf(string(output))
	}
	log.Printf("SUCCESS ✅")
}

func generateMain2ForArion(dir string, data interface{}) {
	log.Printf("Removing old auto-generated files ...")

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), "z_") {
			log.Printf("- removing %s", file.Name())
			os.Remove(file.Name())
		}
	}

	tplName := filepath.Join(dir, "/templates/z_main2.go.tpl")
	log.Printf("Generating %s from template %s", main2FileName, tplName)
	t, err := template.ParseFiles(tplName)
	if err != nil {
		log.Fatal(err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		log.Fatal(err)
	}

	fullname := filepath.Join(dir, main2FileName)
	if err := ioutil.WriteFile(fullname, buf.Bytes(), 0666); err != nil {
		log.Fatal(err)
	}

	log.Printf("Generated %s", main2FileName)
}

func getCompressedStatic(filename string) string {
	var (
		htmlTemplate = compressFileContent(filepath.Join(*targetDir, "/templates/postgal/index.html.tpl"))
		htmlContent  = compressFileContent(filepath.Join(*targetDir, "/web/index.html"))
		cssContent   = compressFileContent(filepath.Join(*targetDir, "/web/m.css"))
		jsContent    = compressFileContent(filepath.Join(*targetDir, "/web/m.js"))
		buf          bytes.Buffer
	)

	log.Printf("Generate and compress static file %s", filename)

	t, err := template.ParseFiles(filename)
	if err != nil {
		log.Fatal(err)
	}

	if err := t.Execute(&buf, struct {
		HTMLSource, CSSSource, JSSource, HTMLTemplate string
	}{
		HTMLSource:   htmlContent,
		CSSSource:    cssContent,
		JSSource:     jsContent,
		HTMLTemplate: htmlTemplate,
	}); err != nil {
		log.Fatal(err)
	}

	return gozip.CompressString(buf.String())
}

func compressFileContent(filename string) string {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("❌ Failed to read file %s: %v", filename, err)
	}
	return gozip.CompressString(string(content))
}

func canBuild() bool {
	_, err := exec.LookPath("go")
	return err == nil
}
