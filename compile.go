package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/straightdave/arion/lib/asyncexec"
	"github.com/straightdave/lesphina/v2"
)

const (
	defaultTimeout = 60 * time.Second
)

func mustCompileDir(dir string, mock, update, verbose bool) {
	if err := compileDir(dir, mock, update, verbose); err != nil {
		log.Fatal(err)
	}
}

func compileDir(dir string, mock, update, verbose bool) error {
	_, err := exec.LookPath("go")
	if err != nil {
		log.Println(yellow("Go compiler seems not installed. Please check your Go toolchain."))
		return err
	}

	var outBin string
	if mock {
		outBin = "mock"
	} else {
		outBin = "postgal"
	}

	cDir, err := os.Getwd()
	if err != nil {
		log.Println("failed to get working dir:", err.Error())
		return err
	}

	if !path.IsAbs(dir) {
		dir = path.Join(cDir, dir)
	}

	// change dir
	log.Println("change dir to", dir)
	err = os.Chdir(dir)
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

	if update {
		log.Println(yellow("force updating"))
		cmdToGetDep.Args = append(cmdToGetDep.Args, "-u", "-f")
	}

	cmdToGetDep.Args = append(cmdToGetDep.Args, deps...)
	err = cmdToGetDep.StartWithTimeout(defaultTimeout)
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

	log.Println("Build ...")
	cmdToBuild.Args = append(cmdToBuild.Args, "-o", outBin)
	if err := cmdToBuild.StartWithTimeout(defaultTimeout); err != nil {
		return err
	}
	return nil
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

	resMap := make(map[string]bool)
	for _, f := range files {
		debug("analyzing file: %s\n", f.Name())
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".go") {
			l, err := lesphina.Read(f.Name())
			if err != nil {
				debug("failed to analyze file: %s\n", f.Name())
				return nil, err
			}

			for _, i := range l.Meta.Imports {
				path := strings.Trim(i.Name, ` "`)
				spl := strings.Split(path, `/`)
				if len(spl) > 1 && strings.Contains(spl[0], `.`) {
					// possibly a non-built-in packages
					debug("dep: %v\n", path)
					resMap[path] = true
				}
			}
		}
	}

	var res []string
	for key := range resMap {
		res = append(res, key)
	}
	return res, nil
}
