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
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	gozip "github.com/straightdave/gozip/lib"
)

var (
	dev       = flag.Bool("dev", false, "to serve static pages")
	port      = flag.String("port", "9999", "local port to serve")
	targetDir = flag.String("target", ".", "target dir")
)

func main() {
	flag.Parse()

	if *targetDir == "" {
		fmt.Println("lack of target dir (-target)")
		return
	}

	if *dev {
		rootDir := filepath.Join(*targetDir, "web")
		serveStaticPage(rootDir)
		return
	}

	build(*targetDir)
}

func serveStaticPage(rootDir string) {
	fs := http.FileServer(http.Dir(rootDir))
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
		CompressedMeta:   compressFileContent(filepath.Join(targetDir, "/snippets/t_meta.go.txt")),
		CompressedMain:   compressFileContent(filepath.Join(targetDir, "/snippets/t_main.go.txt")),
		CompressedStatic: getCompressedStatic(filepath.Join(targetDir, "/snippets/t_web.go.txt")),
	})

	// build
	fmt.Println("begin building ...")
	output, err := exec.Command("go", "build").CombinedOutput()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(output))
	fmt.Println("SUCCESS: Arion is built.")
}

func generateMain2ForArion(dir string, data interface{}) {
	fmt.Println("generating main2 ...")

	// remove existing main2~ under dir
	fmt.Println("= removing all main2* under", dir)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if strings.HasPrefix(file.Name(), "main2") {
			fmt.Println("removing", file.Name())
			os.Remove(file.Name())
		}
	}

	// gen content
	tplName := filepath.Join(dir, "/snippets/arion_main2.go.txt")
	fmt.Println("= generating main2.go from template:", tplName)
	t, err := template.ParseFiles(tplName)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		panic(err)
	}

	// write to new main2
	fullname := filepath.Join(dir, "main2.go")
	if err := ioutil.WriteFile(fullname, buf.Bytes(), 0666); err != nil {
		panic(err)
	}

	fmt.Println("= main2 generated")
}

func getCompressedStatic(filename string) string {
	fmt.Println("generate and compress static file:", filename)
	t, err := template.ParseFiles(filename)
	if err != nil {
		panic(err)
	}

	htmlContent := compressFileContent(filepath.Join(*targetDir, "/web/index.html"))
	htmlTemplate := compressFileContent(filepath.Join(*targetDir, "/snippets/t_index.html.txt"))
	cssContent := compressFileContent(filepath.Join(*targetDir, "/web/m.css"))
	jsContent := compressFileContent(filepath.Join(*targetDir, "/web/m.js"))

	var buf bytes.Buffer
	if err := t.Execute(&buf, struct {
		HtmlSource, CssSource, JsSource, HtmlTemplate string
	}{
		HtmlSource:   htmlContent,
		CssSource:    cssContent,
		JsSource:     jsContent,
		HtmlTemplate: htmlTemplate,
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
