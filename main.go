package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"text/tabwriter"
)

var wg sync.WaitGroup

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		fmt.Println("Incorrect args.\nRun example: app.exe C:\\directory")
		os.Exit(2)
	}
	directory := args[0]
	if err := scanDir(directory); err != nil {
		fmt.Printf("got error: %+v", err)
		os.Exit(1)
	}
}

func scanDir(dir string) error {
	dirInfo, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory %s not exist", dir)
		}
		return fmt.Errorf("failed to get dir %s stat: %+v", dir, err)
	}

	if !dirInfo.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read %s dir: %+v", dir, err)
	}
	wg.Add(len(files))
	data := make(chan string)
	for _, file := range files {
		go getFileInfo(dir, file, data)
	}
	go func() {
		wg.Wait()
		close(data)
	}()
	results := []string{}
	for message := range data {
		results = append(results, message)
	}
	sort.Strings(results)
	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 10, ' ', 0)
	for _, message := range results {
		fmt.Fprintln(writer, message)
	}

	return writer.Flush()
}

type fileInfo struct {
	name        string
	size        string
	isDirectory bool
}

func getFileInfo(baseDirectory string, info os.FileInfo, data chan string) {
	defer wg.Done()
	var size int64
	var err error
	if info.IsDir() {
		size, err = getDirectorySize(baseDirectory, info)
		if err != nil {
			fmt.Printf("failed to get Directory size: %+v", err)
		}
	} else {
		size = info.Size()
	}
	outData := fileInfo{name: info.Name(), size: formatSizeString(size), isDirectory: info.IsDir()}
	data <- formatOutString(outData)
}
func getDirectorySize(baseDirectory string, info os.FileInfo) (int64, error) {
	if info.IsDir() {
		var size int64
		files, err := ioutil.ReadDir(filepath.Join(baseDirectory, info.Name()))
		if err != nil {
			return 0, fmt.Errorf("failed to read dir: %+v", err)
		}
		for _, file := range files {
			dirSize, err := getDirectorySize(filepath.Join(baseDirectory, info.Name()), file)
			if err != nil {
				return 0, err
			}
			size += dirSize
		}
		return size, nil
	}
	return info.Size(), nil
}

func formatSizeString(size int64) string {
	kBytes := size / 1024
	if kBytes < 1 {
		return fmt.Sprintf("%d B", size)
	}
	mBytes := kBytes / 1024
	if mBytes < 1 {
		return fmt.Sprintf("%d KB", kBytes)
	}
	gBytes := mBytes / 1024
	if gBytes < 1 {
		return fmt.Sprintf("%d MB", mBytes)
	}
	return fmt.Sprintf("%d.%d GB", gBytes, mBytes%1024)
}

func formatOutString(info fileInfo) string {
	return fmt.Sprintf("Name: %s\tSize: %s\tDirectory: %t", info.name, info.size, info.isDirectory)
}
