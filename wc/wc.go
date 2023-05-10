package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"unicode"
)

type wcResult struct {
	file      string
	lineCount int
	wordCount int
	charCount int
}

func main() {
	lFlag := flag.Bool("l", false, "count lines")
	wFlag := flag.Bool("w", false, "count words")
	cFlag := flag.Bool("c", false, "count characters")
	flag.Parse()

	files := flag.Args()

	if !*lFlag && !*wFlag && !*cFlag {
		*lFlag, *wFlag, *cFlag = true, true, true
	}

	if len(files) == 0 {
		processStdin(*lFlag, *wFlag, *cFlag)
	} else  {
		results := make([]wcResult, len(files))

		for i, file := range files {
			fileResult, err := wc(file)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error processing file '%s': %v\n", file, err)
				continue
			}
			results[i] = fileResult
		}
		displayResults(results, *lFlag, *wFlag, *cFlag)
	}

}

func processStdin(lFlag, wFlag, cFlag bool) {
	result, err := wcReader(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error processing standard input: %v\n", err)
		return
	}
	displayResults([]wcResult{result}, lFlag, wFlag, cFlag)
}

func wc(file string) (wcResult, error) {
	f, err := os.Open(file)
	if err != nil {
		return wcResult{}, err
	}
	defer f.Close()

	result, err := wcReader(f)
	if err != nil {
		return wcResult{}, err
	}
	result.file = file

	return result, nil
}

func wcReader(reader io.Reader) (wcResult, error) {
	bufReader := bufio.NewReader(reader)
	lineCount, wordCount, charCount := 0, 0, 0
	inWord := false

	for {
		r, size, err := bufReader.ReadRune()
		if err != nil && err != io.EOF {
			return wcResult{}, err
		}

		if err == io.EOF {
			break
		}

		charCount += size

		if r == '\n' {
			lineCount++
		}

		if unicode.IsSpace(r) {
			inWord = false
		} else if !inWord {
			wordCount++
			inWord = true
		}
	}

	return wcResult{
		lineCount: lineCount,
		wordCount: wordCount,
		charCount: charCount,
	}, nil
}

func displayResults(results []wcResult, lFlag, wFlag, cFlag bool) {
	for _, result := range results {
		if lFlag {
			fmt.Printf("%d	", result.lineCount)
		}
		if wFlag {
			fmt.Printf("%d 	", result.wordCount)
		}
		if cFlag {
			fmt.Printf("%d 	", result.charCount)
		}
		if result.file != "" {
			fmt.Printf("%s", result.file)
		}
		fmt.Println()
	}
}
