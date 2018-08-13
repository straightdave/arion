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
	fListPostgals   = flag.Bool("l", false, "list Postgals in current folder or all ./temp* folders")
	fCrossBuild     = flag.String("cross", "", "Cross-platform building flags. e.g 'GOOS=linux GOARCH=amd64'")
	fVerbose        = flag.Bool("verbose", false, "print verbose information when building postgals")

	vRegexPackageLine = regexp.MustCompile(`package (.+)`)

	green  = color.New(color.FgGreen).SprintfFunc()
	yellow = color.New(color.FgYellow).SprintfFunc()
)

func main() {
	flag.Parse()

	if *fListPostgals {
		listPostgals()
		return
	}

	if *fSourceFile == "" {
		log.Fatalln("sourceFile cannot be blank")
	}

	baseName := filepath.Base(*fSourceFile)
	baseName = strings.TrimSuffix(baseName, ".go")
	baseName = strings.TrimSuffix(baseName, ".pb")
	tmpDir, err := ioutil.TempDir(".", "temp-"+baseName+"-")
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
	err = compileDir(tmpDir, *fOutputFile, *fCrossBuild, *fGoGetUpdate, *fVerbose)
	if err != nil {
		log.Fatalln("failed to compile:", err.Error())
	}
	log.Println(green("SUCCESS"))
}

func listPostgals() {
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
		if strings.Title(s.GetName()) != s.GetName() {
			// hide non-exporting structures
			continue
		}
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

func compileDir(dirName, binOutputName, crossBuild string, usingUpdate, verbose bool) error {
	_, err := exec.LookPath("go")
	if err != nil {
		log.Println(yellow("no go installed"))
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

	var output []byte

	// get dependent packages
	log.Printf("GO-get all dependencies ... ")
	updateOptions := []string{"get", "-d"}
	if usingUpdate {
		log.Println("(force-update)")
		updateOptions = append(updateOptions, "-u")
	} else {
		log.Println()
	}
	if verbose {
		updateOptions = append(updateOptions, "-v")
	}
	updateOptions = append(updateOptions, "./...")
	output, err = exec.Command("go", updateOptions...).CombinedOutput()

	if verbose {
		log.Println(string(output))
	}
	if err != nil {
		log.Println(yellow("failed to get dependent packages: %v", err))
	}

	// build
	log.Println("Go-build ...")
	buildOptions := []string{}
	if crossBuild != "" {
		log.Println("Cross building:", crossBuild)

		// get golang.org/x/sys/unix
		// this is required to build linux binaries
		log.Println("Getting golang.org/x/sys/unix ...")
		if err := exec.Command("go", "get", "golang.org/x/sys/unix").Run(); err != nil {
			log.Println(yellow("failed to get golang.org/x/sys/unix"))
			return err
		}

		splts := strings.Split(crossBuild, " ")
		buildOptions = append(buildOptions, splts...)
		buildOptions = append(buildOptions, "go", "build", "-o", binOutputName, "-v")
		output, err = exec.Command("env", buildOptions...).CombinedOutput()
	} else {
		buildOptions = append(buildOptions, "build", "-o", binOutputName, "-v")
		output, err = exec.Command("go", buildOptions...).CombinedOutput()
	}

	if verbose {
		log.Println(string(output))
	}
	if err != nil {
		log.Println("failed to compile:", err.Error())
		return err
	}
	return nil
}
