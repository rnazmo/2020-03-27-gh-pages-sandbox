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

type FileList []File

type File struct {
	Name string
	Mod  string
	// Path string
	// tag  []string
}

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

// -------------------------------------------------------------------

// Errors
var (
	ErrNotFileButDir = errors.New("the path is not a file, but a directory")
	ErrNotDirButFile = errors.New("the path is not a directory, but a file")
)

// Exists check if the filepath/dirpath exists. This
// return true if the directory or file already exists,
// false if does not exists.
// Ref:
//   https://golang.org/pkg/os/#File.Stat
//   https://golang.org/pkg/os/#Stat
//   https://golang.org/pkg/os/#IsNotExist
//
func Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// IsDir check if the dirPath is directory or not.
// NOTE: Behavior is undefined if the path does not exist
//
// NOTE: そのパスが存在しない場合の動作は未定義.
func IsDir(dirPath string) bool {
	fi, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		return false // The path not found.
	}
	return fi.IsDir()
}

// ExistsAndIsDir is ...
//
// dirPath が存在し, かつ確かにディレクトリ（≠ファイル）であることを確かめる.
// dirPath が存在し, かつ確かにディレクトリであるなら nil,
// それ以外なら non-nil error を返す.
//
func ExistsAndIsDir(dirPath string) error {
	if !Exists(dirPath) {
		return os.ErrNotExist
	}
	if !IsDir(dirPath) {
		return ErrNotDirButFile
	}
	return nil // dirPath is indeed a directory.
}

// ListFilesInDir returns a list of filenames directly under the dirPath. The list
// is sorted by filename. ListFilesInDir returns non-nil err if "dirPath" does
// not exist, if "dirPath" is a file (not a directory), or if failed to
// read list of sample files under the "dirPath".
//
// Note that if there is nothing under dirPath or there are only
// directories, ListFilesInDir returns a zero-length list and a nil error.
//
// TODO: ListFilesInDir 関数内のリファクタリング.
// TODO: filepath.Walk() のラッパーとして方が良い？
//   Ref: https://golang.org/pkg/path/filepath/#Walk
//
func ListFilesInDir(dirPath string) (FileList, error) {
	if err := ExistsAndIsDir(dirPath); err != nil {
		return nil, err
	}

	// dirEntriesInfos is a list of filenames and dirnames directory under the dirPath.
	// Note that the list returned by ioutil.ReadDir() includes not only files but
	// also directories. Ref:
	// https://golang.org/src/io/ioutil/ioutil.go?s=2978:3029#L86
	// https://golang.org/pkg/os/#File.Readdir
	// https://golang.org/pkg/os/#FileInfo
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

// Mkfile make file under the current directory. You should Exists()
// before doing this.
// Note: If the file already exists, the old file is truncated and create new one.
//
// The content passed as an argument is written to the file as it is
// (No newline characters are added. If necessary, add a newline character
//  before passing it as a argument).
//
// ioutil.WriteFile() のラッパー.
// NOTE: filePath が既に存在する場合の動作は未定義.
// NOTE: contents のサイズが小さいことを想定している. 大きい場合は bufio
// を使ってバッファリングする等すること.
// NOTE: 自前で実装しようとしたが, f.Close() の返り値の扱いが難しかったので,
// 安心して使える ioutil.WriteFile() のラッパー関数に落ち着いた.
//   Ref: https://github.com/golang/go/blob/1cd724acb6304d30d8998d14a5469fbab24dd3b1/src/io/ioutil/ioutil.go#L84-L88
//        https://github.com/golang/go/blob/7d2473dc81c659fba3f3b83bc6e93ca5fe37a898/src/os/example_test.go#L30-L36
//
// 挙動のメモ:
//   - filePath が既に存在する場合, そのファイル内容を上書きをする.
//   - contents が nil または []byte("") の場合, 何も書き込まれない.
//
// TODO: 挙動を再度確認し, ここのコメントを書き直す
//
// Ref: https://stackoverflow.com/q/1821811
func Mkfile(filePath string, contents []byte) error {
	return ioutil.WriteFile(filePath, contents, 0666)
}
