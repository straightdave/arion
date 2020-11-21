package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/straightdave/lesphina/v2"
)

// SourceFile ...
type SourceFile struct {
	Src      string
	BaseName string
	Ext      string
	Les      *lesphina.Lesphina
}

func mustReadSourceFiles(src string) []*SourceFile {
	res, err := readSourceFiles(src)
	if err != nil {
		log.Fatal(err)
	}
	return res
}

func readSourceFiles(src string) ([]*SourceFile, error) {
	if strings.TrimSpace(src) == "" {
		return nil, fmt.Errorf("src shouldn't be blank")
	}

	var files []*SourceFile
	paths := strings.Split(src, ",")

	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}

		_, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("file %s not exist", path)
		}

		les, err := lesphina.Read(path)
		if err != nil {
			return nil, fmt.Errorf("cannot parse using lesphina: %v", err)
		}

		f := &SourceFile{
			Src:      path,
			BaseName: filepath.Base(path),
			Ext:      filepath.Ext(path),
			Les:      les,
		}
		files = append(files, f)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no valid file parsed")
	}
	return files, nil
}
