package main

import (
	"archive/zip"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	outPath = flag.String("to", "", "Output file path")
)

func usage() {
	fmt.Print(`Usage of zip:

  zip [-to=out/path] file1 [file2 ...]

  Provide one or more file/folder paths to be archived into one .zip file.
  All paths must be rooted in the same folder.
  The default output file name is that of the first given input file with the
  extension changed to .zip. Provide the -to option as the first argument to
  overwrite this path.
`)
}

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) == 0 {
		usage()
		return
	}

	root := filepath.Dir(args[0])
	for _, arg := range args[1:] {
		dir := filepath.Dir(arg)
		if dir != root {
			panic("all input files must be in the same folder")
		}
	}

	if *outPath == "" {
		// if no output path is given, use the first input path's name and put
		// it in the current working dir
		*outPath = stripExt(filepath.Base(args[0])) + ".zip"
		// in case this file already exists, create a new file with " (2)" or
		// " (3)" etc. appended
		uniqueOutPath := *outPath
		n := 1
		for exists(uniqueOutPath) {
			n++
			uniqueOutPath = extendFileName(*outPath, n)
		}
		*outPath = uniqueOutPath
	}

	var buf bytes.Buffer
	func() {
		w := zip.NewWriter(&buf)
		defer func() {
			check(w.Close())
		}()
		for _, path := range args {
			check(filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
				check(err)

				rel, err := filepath.Rel(root, path)
				check(err)

				if info.IsDir() {
					_, err := w.Create(rel + string(filepath.Separator))
					check(err)
				} else {
					f, err := w.Create(rel)
					check(err)
					in, err := os.Open(path)
					check(err)
					defer in.Close()
					_, err = io.Copy(f, in)
					check(err)
				}
				return nil
			}))
		}
	}()
	check(ioutil.WriteFile(*outPath, buf.Bytes(), 0666))
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func stripExt(path string) string {
	return strings.TrimSuffix(path, filepath.Ext(path))
}

func exists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil || os.IsExist(err)
}

func extendFileName(path string, n int) string {
	return fmt.Sprintf("%s (%d)%s", stripExt(path), n, filepath.Ext(path))
}
