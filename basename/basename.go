package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	suffixFlag := flag.String("s", "", "Suffix to remove from the file names")
	helpFlag := flag.Bool("help", false, "Show help message and exit")
	multipleFlag := flag.Bool("a", false, "Support multiple arguments and treat each as a NAME")
	zeroFlag := flag.Bool("z", false, "Separate output with NUL rather than a newline")

	flag.Parse()

	if *helpFlag || len(flag.Args()) == 0 {
		fmt.Println("Usage: basename [OPTION] NAME [SUFFIX]")
		fmt.Println("Print NAME with any leading directory components removed.")
		fmt.Println("If specified, also remove a trailing SUFFIX.")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		os.Exit(0)
	}

	var suffix string
	if len(flag.Args()) > 1 && !*multipleFlag {
		suffix = flag.Args()[1]
	} else {
		suffix = *suffixFlag
	}

	if *multipleFlag {
		separator := "\n"
		if *zeroFlag {
			separator = "\x00"
		}

		for _, filePath := range flag.Args() {
			baseName := filepath.Base(filePath)
			if suffix != "" && strings.HasSuffix(baseName, suffix) {
				baseName = baseName[:len(baseName)-len(suffix)]
			}
			fmt.Print(baseName)
			fmt.Print(separator)
		}
	} else {
		filePath := flag.Args()[0]
		baseName := filepath.Base(filePath)
		if suffix != "" && strings.HasSuffix(baseName, suffix) {
			baseName = baseName[:len(baseName)-len(suffix)]
		}
		fmt.Println(baseName)
	}
}
