// hand-made build tool for this project:
// - do texting work while compiling
// - serve pages while developing
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	gozip "github.com/straightdave/gozip/lib"
)

var (
	dev  = flag.Bool("dev", false, "to serve static pages")
	port = flag.String("port", "8168", "local port to serve")
	root = flag.String("root", "../web/", "root dir")
)

func main() {
	flag.Parse()

	if *dev {
		if *root == "" {
			fmt.Println("lack of '-r' value (root dir)")
			return
		}

		serveStaticPage()
		return
	}

	build("../")
}

func serveStaticPage() {
	fs := http.FileServer(http.Dir(*root))
	http.Handle("/", fs)

	p := ":" + strings.TrimLeft(*port, ":")
	fmt.Println("serving dev at localhost on", p)
	http.ListenAndServe(p, nil)
}

func build(targetDir string) {
	if !canBuild() {
		fmt.Println("cannot build: cannot find command 'go'...")
		return
	}

	generateMain2ForArion(targetDir, struct {
		CompressedMeta, CompressedMain, CompressedStatic string
	}{
		CompressedMeta:   compressFileContent("../templates/t_meta.go.txt"),
		CompressedMain:   compressFileContent("../templates/t_main.go.txt"),
		CompressedStatic: getCompressedStatic("../templates/t_web.go.txt"),
	})

	if err := exec.Command("go", "build", targetDir); err != nil {
		panic(err)
	}
}

func generateMain2ForArion(dir string, data interface{}) {
	// remove existing main2 under dir
	if err := exec.Command("rm", filepath.Join(dir, "main2*")); err != nil {
		panic(err)
	}

	// gen content
	t, err := template.New("main2").ParseFiles("../templates/arion_main2.go.txt")
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		panic(err)
	}

	// write to file
	filename := fmt.Sprint("main2.%s.go", time.Now().UTC()) // be careful: not very random
	fullname := filepath.Join(dir, filename)
	if err := ioutil.WriteFile(fullname, buf.Bytes(), 0666); err != nil {
		panic(err)
	}
}

func getCompressedStatic(filename string) string {
	htmlContent := compressFileContent(filepath.Join(*root, "index.html"))
	cssContent := compressFileContent(filepath.Join(*root, "m.css"))
	jsContent := compressFileContent(filepath.Join(*root, "m.js"))

	t, err := template.New("static").ParseFiles(filename)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, struct {
		HtmlSource, CssSource, JsSource string
	}{
		HtmlSource: htmlContent,
		CssSource:  cssContent,
		JsSource:   jsContent,
	}); err != nil {
		panic(err)
	}

	return gozip.CompressString(buf.String())
}

func compressFileContent(filename string) string {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return ""
	}
	return gozip.CompressString(string(content))
}

func canBuild() (res bool) {
	defer func() {
		if r := recover(); r != nil {
			res = false
		}
	}()

	if _, err := exec.LookPath("go"); err != nil {
		return
	}
	return true
}
