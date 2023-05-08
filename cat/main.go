package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

const MAN_PAGE = `...`

type Program struct {
	number         bool
	numberNonBlank bool
	showEnds       bool
	showTabs       bool
	squeezeBlank   bool
	paths          []string
}

func NewProgram() *Program {
	cat := &Program{
		number:         false,
		numberNonBlank: false,
		showEnds:       false,
		showTabs:       false,
		squeezeBlank:   false,
		paths:          make([]string, 0),
	}

	flag.BoolVar(&cat.number, "n", false, "number all output lines")
	flag.BoolVar(&cat.numberNonBlank, "b", false, "number nonempty output lines")
	flag.BoolVar(&cat.showEnds, "E", false, "display $ at end of each line")
	flag.BoolVar(&cat.showTabs, "T", false, "display TAB characters as ^I")
	flag.BoolVar(&cat.squeezeBlank, "s", false, "squeeze multiple adjacent empty lines")

	help := flag.Bool("h", false, "display this help and exit")
	flag.Parse()

	if *help {
		fmt.Println(MAN_PAGE)
		os.Exit(0)
	}

	cat.paths = flag.Args()
	return cat
}

func main() {
	cat := NewProgram()
	if len(cat.paths) == 0 {
		cat.paths = []string{"-"}
	}
	cat.execute()
}

func (p *Program) execute() {
	for _, path := range p.paths {
		var file *os.File
		var err error

		if path == "-" {
			file = os.Stdin
		} else {
			file, err = os.Open(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: could not open file '%s'\n", path)
				os.Exit(1)
			}
			defer file.Close()
		}

		err = p.processFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: could not process file '%s': %v\n", path, err)
			os.Exit(1)
		}
	}
}

func (p *Program) processFile(file *os.File) error {
	reader := bufio.NewReader(file)
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()

	lineNumber := 1
	prevLineBlank := false

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		isBlank := len(line) == 1 && line[0] == '\n'

		if p.squeezeBlank && isBlank && prevLineBlank {
			continue
		}

		if p.numberNonBlank && !isBlank || p.number {
			fmt.Fprintf(writer, "%6d\t", lineNumber)
			lineNumber++
		}

		if p.showTabs {
			line = replaceTabs(line)
		}

		if p.showEnds {
			line = addEndMarker(line)
		}

		writer.WriteString(line)

		prevLineBlank = isBlank
	}
	return nil
}

func replaceTabs(s string) string {
	return strings.ReplaceAll(s, "\t", "^I")
}

func addEndMarker(s string) string {
	if len(s) > 0 && s[len(s)-1] == '\n' {
		return s[:len(s)-1] + "$\n"
	}
	return s
}
