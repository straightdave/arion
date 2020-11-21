package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	packageName = "main"
)

var (
	regexPackageLine = regexp.MustCompile(`package (.+)`)
)

func mustCreateOutDir(customOutputDir string, isMockServer bool) string {
	dir, err := createOutDir(customOutputDir, isMockServer)
	if err != nil {
		log.Fatal(err)
	}
	return dir
}

func createOutDir(customOutputDir string, isMockServer bool) (string, error) {
	outputName := customOutputDir
	if outputName == "" {
		outputName = "arion_"
	}

	if isMockServer {
		return ioutil.TempDir(".", outputName+"mock")
	}
	return ioutil.TempDir(".", outputName)
}

func mustCopySource(src, destDir string) {
	if err := copySource(src, destDir); err != nil {
		log.Fatal(err)
	}
}

func copySource(src, destDir string) error {
	log.Printf("Copy pb file %s to %s:", src, destDir)

	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	fileName := filepath.Base(src)
	newFile := filepath.Join(destDir, fileName)
	newFile = strings.TrimRight(newFile, ".go") + ".go"

	f, err := os.Create(newFile)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(s)
	writer := bufio.NewWriter(f)
	hasChanged := false

	for scanner.Scan() {
		line := scanner.Text()
		if !hasChanged && regexPackageLine.MatchString(line) {
			line = "package " + packageName
			hasChanged = true
		}
		fmt.Fprintln(writer, line)
	}
	return writer.Flush()
}
