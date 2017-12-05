package main

import (
	"fmt"
	"os"
	"io/ioutil"
	"sync"
	"sort"
	"text/tabwriter"
)

var wg sync.WaitGroup

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		fmt.Println("Incorrect args.")
		fmt.Println("Run example: app.exe C:\\directory")
		os.Exit(2)
	}
	directory := args[0]
	if dirInfo, err := os.Stat(directory); os.IsNotExist(err) {
		fmt.Println("Directory", directory, "not exists")
	} else {
		if !dirInfo.IsDir() {
			fmt.Println(directory, "not a directory")
		} else {
			files, err := ioutil.ReadDir(directory)
			if err != nil {
				fmt.Println("Error: ", err)
			} else {
				wg.Add(len(files))
				data := make(chan string)
				for _, file := range files {
					go getFileInfo(directory, file, data)
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
				writer.Flush()
			}
		}
	}
}

type fileInfo struct {
	name string
	size string
	isDirectory bool
}

func getFileInfo(baseDirectory string, info os.FileInfo, data chan string) {
	defer wg.Done()
	var size int64
	if info.IsDir() {
		size = getDirectorySize(baseDirectory, info)
	} else {
		size = info.Size()
	}
	outData := fileInfo{name: info.Name(), size: formatSizeString(size), isDirectory: info.IsDir()}
	data <- formatOutString(outData)
}

func getDirectorySize(baseDirectory string, info os.FileInfo) int64 {
	if info.IsDir() {
		var size int64
		files, err := ioutil.ReadDir(baseDirectory + "\\" + info.Name())
		if err != nil {
			fmt.Println(err)
			return 0
		} else {
			for _, file := range files {
				size += getDirectorySize(baseDirectory + "\\" + info.Name(), file)
			}

			return size
		}
	} else {
		return info.Size()
	}
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
	} else {
		return fmt.Sprintf("%d.%d GB", gBytes, mBytes % 1024)
	}
}

func formatOutString(info fileInfo) string {
	return fmt.Sprintf("Name: %s\tSize: %s\tDirectory: %t", info.name, info.size, info.isDirectory)
}