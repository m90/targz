// Package targz contains methods to create and extract tar gz archives.
//
// Usage (discarding potential errors):
//   	targz.Compress("path/to/the/directory/to/compress", "my_archive.tar.gz")
// This creates an archive in ./my_archive.tar.gz with the folder "compress" (last in the path).
package targz

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"syscall"
)

// Compress creates a archive from the folder inputFilePath points to in the file outputFilePath points to.
// Only adds the last directory in inputFilePath to the archive, not the whole path.
// It tries to create the directory structure outputFilePath contains if it doesn't exist.
// It returns potential errors to be checked or nil if everything works.
func Compress(inputFilePath, outputFilePath string) (err error) {
	inputFilePath = stripTrailingSlashes(inputFilePath)
	inputFilePath, outputFilePath, err = makeAbsolute(inputFilePath, outputFilePath)
	if err != nil {
		return err
	}
	undoDir, err := mkdirAll(filepath.Dir(outputFilePath), 0755)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			undoDir()
		}
	}()

	err = compress(inputFilePath, outputFilePath, filepath.Dir(inputFilePath))
	if err != nil {
		return err
	}

	return nil
}

// Creates all directories with os.MakedirAll and returns a function to remove the first created directory so cleanup is possible.
func mkdirAll(dirPath string, perm os.FileMode) (func(), error) {
	var undoDir string

	for p := dirPath; ; p = path.Dir(p) {
		finfo, err := os.Stat(p)

		if err == nil {
			if finfo.IsDir() {
				break
			}

			finfo, err = os.Lstat(p)
			if err != nil {
				return nil, err
			}

			if finfo.IsDir() {
				break
			}

			return nil, &os.PathError{"mkdirAll", p, syscall.ENOTDIR}
		}

		if os.IsNotExist(err) {
			undoDir = p
		} else {
			return nil, err
		}
	}

	if undoDir == "" {
		return func() {}, nil
	}

	if err := os.MkdirAll(dirPath, perm); err != nil {
		return nil, err
	}

	return func() { os.RemoveAll(undoDir) }, nil
}

// Remove trailing slash if any.
func stripTrailingSlashes(path string) string {
	if len(path) > 0 && path[len(path)-1] == '/' {
		path = path[0 : len(path)-1]
	}

	return path
}

// Make input and output paths absolute.
func makeAbsolute(inputFilePath, outputFilePath string) (string, string, error) {
	inputFilePath, err := filepath.Abs(inputFilePath)
	if err == nil {
		outputFilePath, err = filepath.Abs(outputFilePath)
	}

	return inputFilePath, outputFilePath, err
}

// The main interaction with tar and gzip. Creates a archive and recursively adds all files in the directory.
// The finished archive contains just the directory added, not any parents.
// This is possible by giving the whole path except the final directory in subPath.
func compress(inPath, outFilePath, subPath string) (err error) {
	file, err := os.Create(outFilePath)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			os.Remove(outFilePath)
		}
	}()

	gzipWriter := gzip.NewWriter(file)
	tarWriter := tar.NewWriter(gzipWriter)

	var paths []string
	if err := filepath.WalkDir(inPath, func(path string, di fs.DirEntry, err error) error {
		paths = append(paths, path)
		return err
	}); err != nil {
		return err
	}
	for _, p := range paths {
		if err := writeTarGz(p, tarWriter); err != nil {
			return err
		}
	}

	err = tarWriter.Close()
	if err != nil {
		return err
	}

	err = gzipWriter.Close()
	if err != nil {
		return err
	}

	err = file.Close()
	if err != nil {
		return err
	}

	return nil
}

// Write path without the prefix in subPath to tar writer.
func writeTarGz(path string, tarWriter *tar.Writer) error {
	fileInfo, err := os.Lstat(path)
	if err != nil {
		return err
	}

	if fileInfo.Mode()&os.ModeSocket == os.ModeSocket {
		return nil
	}

	var link string
	if fileInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
		var err error
		if link, err = os.Readlink(path); err != nil {
			return err
		}
	}

	header, err := tar.FileInfoHeader(fileInfo, link)
	if err != nil {
		return err
	}
	header.Name = path

	err = tarWriter.WriteHeader(header)
	if err != nil {
		return err
	}

	if !fileInfo.Mode().IsRegular() {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(tarWriter, file)
	if err != nil {
		return err
	}

	return err
}
