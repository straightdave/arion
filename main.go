package main

import (
	"bufio"
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

	"github.com/fatih/color"
	gozip "github.com/straightdave/gozip/lib"
	"github.com/straightdave/lesphina"
)

var (
	sourceFile       = flag.String("src", "", "source pb.go file")
	outFile          = flag.String("out", "postgal", "output executable binary file")
	regexPackageLine = regexp.MustCompile(`package (.+)`)
)

func main() {
	flag.Parse()
	green := color.New(color.FgGreen).SprintfFunc()

	if *sourceFile == "" {
		log.Fatalln("sourceFile cannot be blank")
	}

	// generate temporary folder in current one
	tmpDir, err := ioutil.TempDir(".", "temp")
	if err != nil {
		log.Fatalln("cannot create temp dir:", err.Error())
	}

	// modify package name of the pb.go file
	err = genTempPbFile(*sourceFile, tmpDir, "pb")
	if err != nil {
		log.Fatalln("failed to gen new pb:", err.Error())
	}

	// generate source code of meta info used by Lesphina
	err = genMetaFile(*sourceFile, tmpDir, "meta")
	if err != nil {
		log.Fatalln("failed to gen meta:", err.Error())
	}

	// restore main.go
	err = restoreFile(_compressedMain, tmpDir, "main")
	if err != nil {
		log.Fatalln("failed to restore main:", err.Error())
	}

	// restore static.go
	err = restoreFile(_compressedStatic, tmpDir, "static")
	if err != nil {
		log.Fatalln("failed to restore static:", err.Error())
	}

	// compile all
	err = compileDir(tmpDir, *outFile)
	if err != nil {
		log.Fatalln("failed to compile:", err.Error())
	}
	log.Println(green("SUCCESS"))
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
	allStructs := les.Query().ByKind(lesphina.KindStruct).All()
	for _, s := range allStructs {
		lesTypeList = append(lesTypeList, s.GetName())
	}

	tpl, err := template.New("meta").Parse(gozip.DecompressString(_compressedMeta))
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
	return newSourceFile(gozip.DecompressString(raw), dirName, fileName)
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
