package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

func main() {
	fmt.Println("Start generating...")

	wd, err := os.Executable()
	if err != nil {
		log.Fatalln(errors.Wrap(err, "failed to os.Executable(): "))
	}
	appRootDir := filepath.Dir(filepath.Dir(filepath.Dir(wd)))
	contentsDirPath := filepath.Join(appRootDir, "/contents")
	fmt.Println("contentsDirPath:", contentsDirPath)

	fmt.Println("Loading template.html")
	// Read a template from the path.
	templateFileName := filepath.Join(appRootDir, "/gen", "template.html")
	t := template.Must(template.ParseFiles(templateFileName))

	buff := new(bytes.Buffer)
	fw := io.Writer(buff)
	// data := struct {
	// 	Title string
	// }{
	// 	Title: "Test",
	// }
	data, err := ListFilesInDir(contentsDirPath)
	if err != nil {
		log.Fatalln(errors.Wrap(err, "failed to ListFilesInDir(): "))
	}

	if err := t.Execute(fw, data); err != nil {
		log.Fatalln(errors.Wrap(err, "failed to t.Execute(): "))
	}

	// fmt.Println(string(buff.Bytes()))

	if err := Mkfile(filepath.Join(contentsDirPath, "/index.html"), buff.Bytes()); err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Generated successfully!!")
}

type FileList []File

type File struct {
	Name string
	Mod  string
	// Path string
	// tag  []string
}

// -------------------------------------------------------------------

// Errors
var (
	ErrNotFileButDir = errors.New("the path is not a file, but a directory")
	ErrNotDirButFile = errors.New("the path is not a directory, but a file")
)

func Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func IsDir(dirPath string) bool {
	fi, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		return false // The path not found.
	}
	return fi.IsDir()
}

func ExistsAndIsDir(dirPath string) error {
	if !Exists(dirPath) {
		return os.ErrNotExist
	}
	if !IsDir(dirPath) {
		return ErrNotDirButFile
	}
	return nil // dirPath is indeed a directory.
}

func ListFilesInDir(dirPath string) (FileList, error) {
	if err := ExistsAndIsDir(dirPath); err != nil {
		return nil, err
	}

	dirEntriesInfo, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	fl := make(FileList, 0, len(dirEntriesInfo))
	for _, info := range dirEntriesInfo {
		if info.IsDir() {
			continue
		}
		fName := info.Name() // fileName
		if fName == "index.html" {
			continue
		}
		fl = append(
			fl,
			File{
				Name: fName,
				Mod:  info.ModTime().Format("2006-01-02"),
				// Mod:  info.ModTime().String(),
			},
		)
	}
	return fl, nil
}

func Mkfile(filePath string, contents []byte) error {
	return ioutil.WriteFile(filePath, contents, 0666)
}
