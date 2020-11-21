package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/straightdave/lesphina/v2"
)

const (
	defaultTimeout = 60 * time.Second
)

func mustCompileDir(dir string, mock, verbose bool) {
	if err := compileDir(dir, mock, verbose); err != nil {
		log.Fatal(err)
	}
}

func compileDir(dir string, mock, verbose bool) error {
	_, err := exec.LookPath("go")
	if err != nil {
		log.Println(yellow("Go compiler seems not installed. Please check your Go toolchain."))
		return err
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

	log.Println("Build ...")
	raw, err := exec.Command("go", "build").CombinedOutput()
	fmt.Println(string(raw))
	return err
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
