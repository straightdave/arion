package main

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"text/template"
	"time"

	gozip "github.com/straightdave/gozip/lib"
	"github.com/straightdave/lesphina/v2"
)

func mustGenProtoTypeFile(files []*SourceFile, destDir string) {
	if err := genProtoTypeFile(files, destDir); err != nil {
		log.Fatal(err)
	}
}

func genProtoTypeFile(files []*SourceFile, destDir string) error {
	var lesTypeList []string

	// get all gRPC request / response types
	for _, file := range files {
		allStructs := file.Les.Query().ByKind(lesphina.KindStruct).All()
		for _, s := range allStructs {
			str, ok := s.(*lesphina.Struct)
			if !ok {
				continue
			}

			// should have ProtoMessage() method,
			// which makes it an ProtoMessage (interface)
			for _, m := range str.GetMethods() {
				if m.Name == "ProtoMessage" {
					lesTypeList = append(lesTypeList, s.GetName())
					break
				}
			}
		}
	}

	newFile := filepath.Join(destDir, "z_types.go")
	log.Println("generating types file:", newFile)

	tpl, err := template.New("meta").Parse(gozip.DecompressString(_compressedMeta))
	if err != nil {
		return err
	}

	tf, err := os.Create(newFile)
	if err != nil {
		return err
	}
	defer tf.Close()

	return tpl.Execute(tf, struct {
		GeneratedTime string
		LesDump       string
		List          []string
	}{
		GeneratedTime: time.Now().Format(time.UnixDate),
		LesDump:       mergeLesphina(files).DumpString(),
		List:          lesTypeList,
	})
}

func mergeLesphina(files []*SourceFile) *lesphina.Lesphina {
	totalMeta := new(lesphina.Meta)

	for _, file := range files {
		meta := file.Les.Meta

		totalMeta.Consts = append(totalMeta.Consts, meta.Consts...)
		totalMeta.Functions = append(totalMeta.Functions, meta.Functions...)
		totalMeta.Imports = append(totalMeta.Imports, meta.Imports...)
		totalMeta.Interfaces = append(totalMeta.Interfaces, meta.Interfaces...)
		totalMeta.Structs = append(totalMeta.Structs, meta.Structs...)
		totalMeta.Vars = append(totalMeta.Vars, meta.Vars...)

		totalMeta.NumConst += totalMeta.NumConst
		totalMeta.NumFunction += totalMeta.NumFunction
		totalMeta.NumImport += totalMeta.NumImport
		totalMeta.NumInterface += totalMeta.NumInterface
		totalMeta.NumStruct += totalMeta.NumStruct
		totalMeta.NumVar += totalMeta.NumVar
	}

	return &lesphina.Lesphina{
		Meta: totalMeta,
	}
}

func mustRestoreFile(raw, dir, file string) {
	if err := restoreFile(raw, dir, file); err != nil {
		log.Fatal(err)
	}
}

func restoreFile(raw, dirName, fileName string) error {
	fullName := filepath.Join(dirName, fileName)
	log.Println("creating new source file:", fullName)

	f, err := os.Create(fullName)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	w.WriteString(gozip.DecompressString(raw))
	return w.Flush()
}

func mustGenModFile(modName, dir string) {
	if err := genModFile(modName, dir); err != nil {
		log.Fatal(err)
	}
}

func genModFile(modName, dir string) error {
	newFile := filepath.Join(dir, "go.mod")

	tpl, err := template.New("mod").Parse(gozip.DecompressString(_modFile))
	if err != nil {
		return err
	}

	tf, err := os.Create(newFile)
	if err != nil {
		return err
	}
	defer tf.Close()

	return tpl.Execute(tf, struct {
		ModName string
	}{
		ModName: modName,
	})
}
