package main

import (
	"bufio"
	"crypto/md5"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/fatih/color"
	gozip "github.com/straightdave/gozip/lib"
	"github.com/straightdave/lesphina"
)

var (
	fSourceFile     = flag.String("src", "", "source pb.go file")
	fOutputFile     = flag.String("o", "", "output executable binary file")
	fClearTempFiles = flag.Bool("c", false, "to clear temp folder after Postgal is generated\n*NOTE*: use -o to generate Postgal out of temp folder")
	fGoGetUpdate    = flag.Bool("u", false, "update dependencies when building Postgal")
	fListPostgal    = flag.Bool("l", false, "list Postgals in current folder or all ./temp* folders")

	vRegexPackageLine = regexp.MustCompile(`package (.+)`)
)

func main() {
	flag.Parse()
	green := color.New(color.FgGreen).SprintfFunc()

	// list postgals
	if *fListPostgal {
		listPostgal()
		return
	}

	if *fSourceFile == "" {
		log.Fatalln("sourceFile cannot be blank")
	}

	// generate temporary folder in current one
	tmpDir, err := ioutil.TempDir(".", "temp")
	if err != nil {
		log.Fatalln("cannot create temp dir:", err.Error())
	}
	if *fClearTempFiles {
		defer func(dirName string) {
			log.Printf("clear temp folder %s\n", dirName)
			if err := os.RemoveAll(dirName); err != nil {
				log.Fatalln(err)
			}
		}(tmpDir)
	}

	// modify package name of the pb.go file
	err = genTempPbFile(*fSourceFile, tmpDir, "pb")
	if err != nil {
		log.Fatalln("failed to gen new pb:", err.Error())
	}

	// generate source code of meta info used by Lesphina
	err = genMetaFile(*fSourceFile, tmpDir, "meta")
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
	err = compileDir(tmpDir, *fOutputFile, *fGoGetUpdate)
	if err != nil {
		log.Fatalln("failed to compile:", err.Error())
	}
	log.Println(green("SUCCESS"))
}

func listPostgal() {
	cDir, err := os.Getwd()
	if err != nil {
		log.Fatalln("failed to get working dir:", err.Error())
	}

	files, err := ioutil.ReadDir(cDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), "postgal") {
			printPostgalInfo(file, ".")
			continue
		}

		if file.IsDir() && strings.HasPrefix(file.Name(), "temp") {
			fs, e := ioutil.ReadDir(file.Name())
			if e != nil {
				continue
			}

			for _, f := range fs {
				if !f.IsDir() && strings.HasPrefix(f.Name(), "postgal") {
					printPostgalInfo(f, file.Name())
				}
			}
		}
	}
}

func printPostgalInfo(file os.FileInfo, folder string) {
	fullName := fmt.Sprintf("%s/%s", folder, file.Name())
	fmt.Println("[-]", fullName)
	out, err := exec.Command(fullName, "-v").Output()
	if err == nil {
		fmt.Println(string(out))
	}
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
		if !hasChanged && vRegexPackageLine.MatchString(line) {
			line = "package main"
			hasChanged = true // change only once
		}
		fmt.Fprintln(writer, line)
	}
	return writer.Flush()
}

func getFileMD5(fileName string) (checksum string, err error) {
	f, err := os.Open(fileName)
	if err != nil {
		return
	}
	defer f.Close()

	h := md5.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return
	}
	checksum = fmt.Sprintf("%x", h.Sum(nil))
	return
}

func genMetaFile(pbFile, dirName, fileName string) error {
	checksum, err := getFileMD5(pbFile)
	if err != nil {
		return err
	}

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
		GeneratedTime string
		Checksum      string
		LesDump       string
		List          []string
	}{
		GeneratedTime: time.Now().Format(time.UnixDate),
		Checksum:      checksum,
		LesDump:       lesDump,
		List:          lesTypeList,
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

func compileDir(dirName, binOutputName string, usingUpdate bool) error {
	_, err := exec.LookPath("go")
	if err != nil {
		log.Println("no go installed")
		return err
	}

	cDir, err := os.Getwd()
	if err != nil {
		log.Println("failed to get working dir:", err.Error())
		return err
	}

	if !path.IsAbs(dirName) {
		dirName = path.Join(cDir, dirName)
	}

	if binOutputName != "" {
		if !path.IsAbs(binOutputName) {
			binOutputName = path.Join(cDir, binOutputName)
		}
	} else {
		binOutputName = "postgal"
	}

	// change dir
	log.Println("change dir to", dirName)
	err = os.Chdir(dirName)
	if err != nil {
		log.Println("failed to change dir:", err.Error())
		return err
	}

	// change dir back
	defer func() {
		log.Println("change dir back to", cDir)
		err = os.Chdir(cDir)
		if err != nil {
			log.Println("failed to change dir:", err.Error())
		}
	}()

	// run go get ./... first
	if usingUpdate {
		log.Println("force-update all dependencies...")
		_ = exec.Command("go", "get", "-f", "-u", "./...").Run() // ignore exit error
	} else {
		log.Println("get/check all dependencies...")
		_ = exec.Command("go", "get", "./...").Run() // ignore exit error
	}

	// build
	var opts []string
	opts = append(opts, "build", "-v", "-o", binOutputName)
	err = exec.Command("go", opts...).Run()
	if err != nil {
		log.Println("failed to compile:", err.Error())
		return err
	}
	return nil
}
