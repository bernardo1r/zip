package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"time"
)

const usage = ``

func checkError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func defaultZipHeader(filename string) *zip.FileHeader {
	t, err := time.Parse(time.DateOnly, "2000-01-01")
	if err != nil {
		log.Fatalln(err)
	}
	return &zip.FileHeader{
		Name:     filename,
		Method:   zip.Deflate,
		Modified: t,
	}
}

func walkDir(root string, zw *zip.Writer) fs.WalkDirFunc {
	return func(path string, d fs.DirEntry, err error) error {
		if path == "." {
			return nil
		}
		if err != nil {
			fmt.Printf("Could not add %s:", err.Error())
			return nil
		}
		if d.IsDir() {
			return nil
		}

		filename := root + "/" + path
		fmt.Println(filename)
		return compressFile(filename, zw)
	}
}

func compressDir(dirname string, zw *zip.Writer) error {
	rootfs := os.DirFS(dirname)
	return fs.WalkDir(rootfs, ".", walkDir(dirname, zw))
}

func compressFile(filename string, zw *zip.Writer) error {
	fileIn, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func() {
		checkError(fileIn.Close())
	}()

	fileOut, err := zw.CreateHeader(defaultZipHeader(filename))
	if err != nil {
		return err
	}

	_, err = io.Copy(fileOut, fileIn)
	if err != nil {
		return fmt.Errorf("compressing file %s: %w", filename, err)
	}
	return nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println(usage)
		os.Exit(1)
	}

	root := filepath.Base(os.Args[1])
	file, err := os.Create(root + ".zip")
	checkError(err)
	defer func() {
		checkError(file.Close())
	}()
	zw := zip.NewWriter(file)
	defer func() {
		checkError(zw.Close())
	}()

	info, err := os.Stat(root)
	checkError(err)
	if info.IsDir() {
		err = compressDir(root, zw)
	} else {
		fmt.Println(root)
		err = compressFile(root, zw)
	}
	checkError(err)
}
