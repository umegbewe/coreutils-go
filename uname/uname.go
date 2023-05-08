package main

import (
	"bytes"
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

	printInfo := func(flag *bool, byteArray [256]byte) {
		if *flag {
			byteArray := byteArray[:]
			nullTerminated := bytes.IndexByte(byteArray, 0)
			if nullTerminated != -1  {
				byteArray = byteArray[:nullTerminated]
			}
			fmt.Printf("%s\n", byteArray)
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
