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
	"github.com/straightdave/arion/lib/asyncexec"
	gozip "github.com/straightdave/gozip/lib"
	"github.com/straightdave/lesphina"
)

var (
	fSourceFile   = flag.String("src", "", "source pb.go file")
	fOutputFile   = flag.String("o", "", "output executable binary file")
	fGoGetUpdate  = flag.Bool("u", false, "update dependencies when building Postgal")
	fListPostgals = flag.Bool("l", false, "list Postgals in current folder or all ./temp* folders")
	fCrossBuild   = flag.String("cross", "", "Cross-platform building flags. e.g 'GOOS=linux GOARCH=amd64'")
	fVerbose      = flag.Bool("verbose", false, "print verbose information when building postgals")
	fDebug        = flag.Bool("debug", false, "debug mode")

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
		log.Println(yellow("The 'go' compiler is not installed"))
		return err
	}

	if binOutputName == "" {
		binOutputName = "postgal"
	}

	cDir, err := os.Getwd()
	if err != nil {
		log.Println("failed to get working dir:", err.Error())
		return err
	}

	if !path.IsAbs(dirName) {
		dirName = path.Join(cDir, dirName)
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

	log.Printf("Analyzing dependencies ...")
	deps, err := listDepsOfCurrentPackage2()
	if err != nil {
		return err
	}

	debug("deps => %+v\n", deps)

	log.Printf("Install dependencies ...")
	cmdToGetDep := &asyncexec.AsyncExec{
		Name: "go",
		Args: []string{"get", "-d"},
	}

	if verbose {
		cmdToGetDep.Args = append(cmdToGetDep.Args, "-v")
	}

	if usingUpdate {
		log.Println(yellow("force update"))
		cmdToGetDep.Args = append(cmdToGetDep.Args, "-u", "-f")
	}

	cmdToGetDep.Args = append(cmdToGetDep.Args, deps...)
	err = cmdToGetDep.StartWithTimeout(300 * time.Second)
	if err != nil {
		return err
	}

	cmdToBuild := &asyncexec.AsyncExec{
		Name: "go",
		Args: []string{"build"},
	}

	if verbose {
		cmdToBuild.Args = append(cmdToBuild.Args, "-v")
	}

	if crossBuild != "" {
		log.Println(yellow("cross build"))
		for _, s := range strings.Split(crossBuild, " ") {
			s = strings.Trim(s, " \n")
			indexOfEqual := strings.Index(s, "=")
			if indexOfEqual >= 0 {
				cmdToBuild.SetEnv(s[:indexOfEqual], s[indexOfEqual+1:])
			}
		}

		log.Println("Get golang.org/x/sys/unix ...")
		c := &asyncexec.AsyncExec{
			Name: "go",
			Args: []string{"get", "-d", "golang.org/x/sys/unix"},
		}
		if err := c.StartWithTimeout(60 * time.Second); err != nil {
			return err
		}
	}

	log.Println("Build ...")
	cmdToBuild.Args = append(cmdToBuild.Args, "-o", binOutputName)
	if err := cmdToBuild.StartWithTimeout(60 * time.Second); err != nil {
		return err
	}
	return nil
}

func listDepsOfCurrentPackage() ([]string, error) {
	raw := `go list -f '{{join .Imports "\n"}}' | xargs go list -f '{{if not .Standard}}{{.ImportPath}}{{end}}'`

	cmd := &asyncexec.AsyncExec{
		Name: "bash",
		Args: []string{"-c", raw},
	}

	if err := cmd.StartWithTimeout(10 * time.Second); err != nil {
		return nil, err
	}

	if *fDebug {
		fmt.Println("[debug] stdout")
		for _, l := range cmd.Stdout {
			fmt.Println(l)
		}

		fmt.Println("[debug] stderr")
		for _, l := range cmd.Stderr {
			fmt.Println(l)
		}
	}
	return cmd.Stdout, nil
}

func listDepsOfCurrentPackage2() ([]string, error) {
	cdir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	files, err := ioutil.ReadDir(cdir)
	if err != nil {
		return nil, err
	}

	var res []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".go") {
			l, err := lesphina.Read(f.Name())
			if err != nil {
				return nil, err
			}

			for _, i := range l.Meta.Imports {
				path := strings.Trim(i.Name, ` "`)
				spl := strings.Split(path, `/`)
				if len(spl) > 1 && strings.Contains(spl[0], `.`) {
					// possibly a non-built-in packages
					res = append(res, path)
				}
			}
		}
	}
	return res, nil
}

func debug(format string, args ...interface{}) {
	if *fDebug {
		fmt.Printf(yellow("[debug] ")+format, args...)
	}
}
