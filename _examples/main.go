package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/m90/targz"
)

func main() {
	// Create a temporary file structiure we can use
	tmpDir, dirToCompress := createExampleData()

	// Comress a folder to my_archive.tar.gz
	err := targz.Compress(dirToCompress, filepath.Join(tmpDir, "my_archive.tar.gz"))
	if err != nil {
		fmt.Println("Comress error")
		panic(err)
		os.Exit(1)
	}

	fmt.Println(tmpDir)
	os.Exit(0)
}

func createExampleData() (string, string) {
	tmpDir, err := ioutil.TempDir("", "targz-example")
	if err != nil {
		fmt.Println("tmpdir error")
		panic(err)
		os.Exit(1)
	}

	directory := filepath.Join(tmpDir, "my_folder")
	subDirectory := filepath.Join(directory, "my_sub_folder")
	err = os.MkdirAll(subDirectory, 0755)
	if err != nil {
		fmt.Println("mkdir error")
		panic(err)
		os.Exit(1)
	}

	emptyDir := filepath.Join(directory, "empty")
	err = os.MkdirAll(emptyDir, 0755)

	if err := os.WriteFile(filepath.Join(subDirectory, "my_file.txt"), []byte("example data\n"), 0666); err != nil {
		fmt.Println("create file error")
		panic(err)
		os.Exit(1)
	}

	if err := os.Symlink(filepath.Join(subDirectory, "my_file.txt"), filepath.Join(subDirectory, "my_link")); err != nil {
		fmt.Println("create symlink error")
		panic(err)
		os.Exit(1)
	}

	return tmpDir, directory
}
