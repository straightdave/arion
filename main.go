package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/straightdave/lesphina"
	"github.com/straightdave/lesphina/entry"
)

var (
	source  = flag.String("src", "", "source pb.go file")
	outFile = flag.String("out", "xclient", "output file")
	isClean = flag.Bool("clean", false, "to remove the temp files")

	regexPackageLine = regexp.MustCompile(`package (.+)`)
)

func main() {
	flag.Parse()

	if *source == "" {
		fmt.Println("source cannot be blank")
		return
	}

	// create temp dir inside current dir
	tmpDir, err := ioutil.TempDir(".", "temp")
	if err != nil {
		fmt.Println("cannot create temp dir:", err.Error())
		return
	}
	if *isClean {
		defer os.RemoveAll(tmpDir)
	}

	// gen modified pb file
	err = genTempPbFile(*source, tmpDir, "pb")
	if err != nil {
		fmt.Println("failed to gen new pb:", err.Error())
		return
	}

	// gen meta snippet file
	err = genMetaFile(*source, tmpDir, "meta")
	if err != nil {
		fmt.Println("failed to gen meta:", err.Error())
		return
	}

	// restore main.go
	err = restoreFile(compressedMain, tmpDir, "main")
	if err != nil {
		fmt.Println("failed to restore main:", err.Error())
		return
	}

	// restore static.go
	err = restoreFile(compressedStatic, tmpDir, "static")
	if err != nil {
		fmt.Println("failed to restore static:", err.Error())
		return
	}

	// compile all
	err = compileDir(tmpDir, *outFile)
	if err != nil {
		fmt.Println("failed to compile:", err.Error())
		return
	}
	fmt.Println("SUCCESS")
}

func genTempPbFile(sourceFile, dirName, fileName string) error {
	source, err := os.Open(sourceFile)
	if err != nil {
		return err
	}
	defer source.Close()

	fullName := filepath.Join(dirName, fileName)
	fullName = strings.TrimRight(fullName, ".go") + ".go"
	log.Println("generating new pb file:", fullName)

	f, err := os.Create(fullName)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(source)
	writer := bufio.NewWriter(f)
	hasChanged := false

	for scanner.Scan() {
		line := scanner.Text()
		if !hasChanged && regexPackageLine.MatchString(line) {
			line = "package main"
			hasChanged = true // change only once
		}
		fmt.Fprintln(writer, line)
	}
	return writer.Flush()
}

func genMetaFile(pbFile, dirName, fileName string) error {
	fullName := filepath.Join(dirName, fileName)
	fullName = strings.TrimRight(fullName, ".go") + ".go"
	log.Println("generating meta source file:", fullName)

	les, err := lesphina.Read(pbFile)
	if err != nil {
		return err
	}
	lesDump := les.DumpString()

	var lesTypeList []string
	allStructs := les.Query().ByKind(entry.KindStruct).All()
	for _, s := range allStructs {
		lesTypeList = append(lesTypeList, s.GetName())
	}

	tpl, err := template.New("meta").Parse(snippetMeta)
	if err != nil {
		return err
	}
	tf, err := os.Create(fullName)
	if err != nil {
		return err
	}
	defer tf.Close()

	return tpl.Execute(tf, struct {
		LesDump string
		List    []string
	}{
		LesDump: lesDump,
		List:    lesTypeList,
	})
}

func newSourceFile(content, dirName, fileName string) error {
	fullName := filepath.Join(dirName, fileName)
	fullName = strings.TrimRight(fullName, ".go") + ".go"
	log.Println("creating new source file:", fullName)

	f, err := os.Create(fullName)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	w.WriteString(content)
	return w.Flush()
}

func restoreFile(raw, dirName, fileName string) error {
	c, err := decompress(raw)
	if err != nil {
		return err
	}
	return newSourceFile(c, dirName, fileName)
}

func compileDir(dirName, binFileName string) error {
	if !path.IsAbs(dirName) {
		dirName = "./" + dirName
	}

	fullName := filepath.Join(dirName, binFileName)
	log.Println("compiling to file:", fullName)

	_, err := exec.LookPath("go")
	if err != nil {
		return err
	}

	var opts []string
	opts = append(opts, "build", "-v", "-o", fullName, dirName)

	return exec.Command("go", opts...).Run()
}

func decompress(raw string) (string, error) {
	raw = strings.TrimSpace(raw)

	zippedStr, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return "", err
	}

	unzipReader, err := gzip.NewReader(bytes.NewReader(zippedStr))
	if err != nil {
		return "", err
	}

	res, err := ioutil.ReadAll(unzipReader)
	if err != nil {
		return "", err
	}
	return string(res), nil
}

// ========== template ===============
const snippetMeta = `
//
// auto-generated snippet
//
package main

import (
	"reflect"
	"github.com/golang/protobuf/proto"
)

const _lesDump = "{{ .LesDump }}"

func getVarByTypeName(typeName string) proto.Message {
    switch typeName {
    {{ range $element := .List -}}
    case "{{- $element -}}":
        return reflect.New(reflect.TypeOf(& {{- $element -}} {}).Elem()).Interface().(proto.Message)
    {{ end }}
	}
    return nil
}`

const compressedMain = `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`
const compressedStatic = `H4sIAAAAAAAA/1SPQU87IRTE73yK9+//QHsp6VVZLurRaKIXb6Xsq7BlYcN7a90Yv7vZxcb0BMwMv8kM1p3sO0JvQxLC5UQMnvsIDeyF/nf/dPf69vywSEboy4G2NQJAc+CI5tESY4EXHg9aVWk2Y0gnKBgbSTxFJI/IEngasJGMn6wckQRf8NjIfjs/jNCqsvUht9NC8TvzH64a/M5oX0AtNrkSBgYqboZ0dFXQ2Q9bA9JoVW9zR4VrVefsL7sd0TLb7+BLAAC4HHO5gUMc8VZ8/wW7mjuH1ObzNqeYbdscx+Q45ATrze93G7HweuUxxrzazIAfAAAA//8BAAD//8CYnIRuAQAA`
