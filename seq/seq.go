package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
	formatString string
	separator    string
	equalWidth   bool
)

func usage() {
	fmt.Printf("Usage: %s [OPTION]... LAST\n       or:  %s [OPTION]... FIRST LAST\n       or:  %s [OPTION]... FIRST INCREMENT LAST\n", os.Args[0], os.Args[0], os.Args[0])
	fmt.Println("Print numbers from FIRST to LAST, in steps of INCREMENT.")
	fmt.Println("FIRST, INCREMENT default to 1.")
	fmt.Println("Options:")
	fmt.Println("  -f, --format FORMAT     use printf style floating-point FORMAT")
	fmt.Println("  -s, --separator STRING  use STRING to separate numbers (default: \\n)")
	fmt.Println("  -w, --equal-width       equalize width by padding with leading zeroes")
	os.Exit(1)
}

func parseArgs(args []string) (first float64, step float64, last float64) {
	var err error
	switch len(args) {
	case 1:
		first = 1.0
		step = 1.0
		last, err = strconv.ParseFloat(args[0], 64)
	case 2:
		first, err = strconv.ParseFloat(args[0], 64)
		if err != nil {
			fmt.Println("Invalid number provided for FIRST")
			usage()
		}
		step = 1.0
		last, err = strconv.ParseFloat(args[1], 64)
	case 3:
		first, err = strconv.ParseFloat(args[0], 64)
		if err != nil {
			fmt.Println("Invalid number provided for FIRST")
			usage()
		}
		step, err = strconv.ParseFloat(args[1], 64)
		if err != nil {
			fmt.Println("Invalid number provided for STEP")
			usage()
		}
		last, err = strconv.ParseFloat(args[2], 64)
	default:
		usage()
	}
	if err != nil {
		fmt.Println("Invalid number provided for LAST")
		usage()
	}
	if step == 0.0 {
		fmt.Println("STEP must not be zero")
		usage()
	}
	return
}

func main() {
	flag.StringVar(&formatString, "f", "%g", "use printf style floating-point format")
	flag.StringVar(&formatString, "format", "%g", "use printf style floating-point format")
	flag.StringVar(&separator, "s", "\n", "use STRING to separate numbers (default: \\n)")
	flag.StringVar(&separator, "separator", "\n", "use STRING to separate numbers (default: \\n)")
	flag.BoolVar(&equalWidth, "w", false, "equalize width by padding with leading zeroes")
	flag.BoolVar(&equalWidth, "equal-width", false, "equalize width by padding with leading zeroes")
	flag.Parse()

	first, step, last := parseArgs(flag.Args())
	var seq []string
	if step > 0 {
		for i := first; i <= last; i += step {
			seq = append(seq, fmt.Sprintf(formatString, i))
		}
	} else {
		for i := first; i >= last; i += step {
			seq = append(seq, fmt.Sprintf(formatString, i))
		}
	}

	if equalWidth {
		maxLen := 0
		for _, num := range seq {
			if len(num) > maxLen {
				maxLen = len(num)
			}
		}
		for i, num := range seq {
			seq[i] = fmt.Sprintf("%0*s", maxLen, num)
		}
	}

	fmt.Println(strings.Join(seq, separator))
}
