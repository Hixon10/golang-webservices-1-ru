package main

import (
	"os"
	"io"
	"log"
	"io/ioutil"
	"strconv"
	"fmt"
	"sort"
	"strings"
)

type fileName struct {
	files []fileName
	name  string
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

func dirTree(output io.Writer, dirPath string, showFiles bool) error {

	fileNames := dirTraversal(dirPath, showFiles)

	needDelimiter := make(map[int]bool)

	printFiles(output, *fileNames, 0, false, needDelimiter)

	return nil
}

func dirTraversal(dirPath string, showFiles bool) *fileName {
	dirFiles, err := ioutil.ReadDir(dirPath)
	if err != nil {
		log.Fatal(err)
	}

	files := make([]fileName, 0)

	for _, f := range dirFiles {
		if f.IsDir() {
			newDirPath := dirPath + string(os.PathSeparator) + f.Name()
			newFiles := dirTraversal(newDirPath, showFiles)
			files = append(files, fileName{name: f.Name(), files:(*newFiles).files})
		} else {
			if !showFiles {
				continue;
			}
			sizeStr := "empty"
			if f.Size() > 0 {
				sizeStr = strconv.Itoa(int(f.Size())) + "b"
			}

			fName := f.Name() + " (" + sizeStr + ")"
			files = append(files, fileName{name:fName})
		}
	}

	sort.SliceStable(files, func(i, j int) bool {
		return strings.Compare(files[i].name, files[j].name) < 0
	})

	result := fileName{files:files}
	return &result
}

func printFiles(output io.Writer, files fileName, level int, last bool, needDelimiter map[int]bool) {
	newNeedDelimiter := make(map[int]bool)
	for key, value := range needDelimiter {
		newNeedDelimiter[key] = value
	}

	if len(files.name) > 0 {
		space := ""
		for i := 0; i < level; i++ {
			if i + 1 == level {
				if last {
					space = space + "└───"
					newNeedDelimiter[level] = false
				} else {
					space = space + "├───"
				}
			} else {
				needD := true
				f, ok :=newNeedDelimiter[i + 1]
				if ok {
					needD = f
				}

				if needD {
					space = space + "│" + "\t"
				} else {
					space = space + "\t"
				}
			}
		}

		fmt.Fprintln(output, space + files.name)
	}

	for fileIndex := 0; fileIndex < len(files.files); fileIndex++ {
		last := fileIndex+1 == len(files.files)
		printFiles(output, files.files[fileIndex], level + 1, last, newNeedDelimiter)
	}
}
