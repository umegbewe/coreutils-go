package main

import (
	"flag"
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

func main() {
	all := flag.Bool("a", false, "Behave as though all of the options -s, -n, -r, -v, and -m were specified.")
	sysname := flag.Bool("s", false, "Print the kernel name.")
	nodename := flag.Bool("n", false, "Print the network node hostname.")
	release := flag.Bool("r", false, "Print the kernel release.")
	version := flag.Bool("v", false, "Print the kernel version.")
	machine := flag.Bool("m", false, "Print the machine hardware name.")

	flag.Parse()

	if !*all && !*sysname && !*nodename && !*release && !*version && !*machine {
		*sysname = true
	}

	utsname := &unix.Utsname{}
	if err := unix.Uname(utsname); err != nil {
		fmt.Fprintf(os.Stderr, "uname: %v\n", err)
		os.Exit(1)
	}

	printInfo := func(flag *bool, cString interface{}) {
		if *flag {
			var byteSlice []byte
			switch v := cString.(type) {
			case [256]byte:
				byteSlice = make([]byte, len(v))
				for i, c := range v {
					byteSlice[i] = byte(c)
				}
			case [65]byte:
				byteSlice = make([]byte, len(v))
				for i, c := range v {
					byteSlice[i] = byte(c)
				}
			default:
				fmt.Fprintf(os.Stderr, "Unsupported array type: %T\n", v)
				os.Exit(1)
			}
			s := toGoString(byteSlice)
			fmt.Printf("%s\n", s)
		}
	}

	if *all {
		fmt.Printf("%s %s %s %s %s\n", utsname.Sysname, utsname.Nodename, utsname.Release, utsname.Version, utsname.Machine)
	}

	printInfo(sysname, utsname.Sysname)
	printInfo(nodename, utsname.Nodename)
	printInfo(release, utsname.Release)
	printInfo(version, utsname.Version)
	printInfo(machine, utsname.Machine)

}
