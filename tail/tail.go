package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"io"
	"os"
	"sync"
)

var (
	numLines int
	follow   bool
	verbose  bool
	retry 	 bool

	printedHeader map[string]bool = make(map[string]bool)
)

func printLastNLines(filename string, n int) error {

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	bufSize := stat.Size()
	if bufSize > 1e6 { // limit buffer size to 1MB
		bufSize = 1e6
	}

	buf := make([]byte, bufSize)
	startPos := stat.Size()
	lineCount := 0
	var positions []int64

	for {
		length, err := file.ReadAt(buf, startPos-int64(len(buf)))
		if err != nil && err != io.EOF {
			return err
		}

		for i := length - 1; i >= 0; i-- {
			if buf[i] == '\n' {
				lineCount++
				if lineCount > n {
					positions = append(positions, startPos-int64(length)+int64(i+1))
					break
				}
			}
		}

		if lineCount > n || startPos-int64(length) <= 0 {
			break
		}

		startPos -= int64(length)
	}

	if len(positions) == 0 {
		return nil
	}

	_, err = file.Seek(positions[len(positions)-1], 0)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	return scanner.Err()
}


func followFile(filename string, wg *sync.WaitGroup) error {
	defer wg.Done()

	if verbose {
		fmt.Printf("==> %s <==\n", filename)
	}

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Start from the end of the file
	pos, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	err = watcher.Add(filename)
	if err != nil {
		return err
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				newPos, err := file.Seek(0, io.SeekEnd)
				if err != nil {
					return err
				}
				if newPos < pos {
					// The file got smaller, so it must have been truncated
					pos = newPos
				}
				if newPos > pos {
					// The file grew, so fucking read it
					readBuf := make([]byte, newPos-pos)
					_, err = file.ReadAt(readBuf, pos)
					if err != nil {
						return err
					}
					fmt.Print(string(readBuf))
					pos = newPos
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			return err
		}
	}
}


func main() {
	flag.IntVar(&numLines, "n", 10, "the number of lines to display")
	flag.BoolVar(&follow, "f", false, "output appended data as the file grows")
	flag.BoolVar(&verbose, "v", false, "always output headers giving file names")
	flag.BoolVar(&retry, "R", false, "keep trying to open a file eveb when it's not accessible")
	flag.Parse()

	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: tail [-n numlines] [-f] <file> [<file> ...]")
		os.Exit(1)
	}

	var wg sync.WaitGroup

	for _, filename := range args {

		if verbose && !printedHeader[filename] {
			fmt.Printf("==> %s <==\n", filename)
			printedHeader[filename] = true
		}
		err := printLastNLines(filename, numLines)
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}

		if follow {
			wg.Add(1)
			go func(filename string) {
				err := followFile(filename, &wg)
				if err != nil {
					fmt.Println("Error: ", err)
					os.Exit(1)
				}
			}(filename)
		}
	}

	wg.Wait()
}
