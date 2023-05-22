package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"syscall"

	px "github.com/mitchellh/go-ps"
)

var signalName string
var signalMap = map[string]syscall.Signal{
	"SIGABRT": syscall.SIGABRT,
	"SIGALRM": syscall.SIGALRM,
	"SIGBUS":  syscall.SIGBUS,
	"SIGFPE":  syscall.SIGFPE,
	"SIGHUP":  syscall.SIGHUP,
	"SIGILL":  syscall.SIGILL,
	"SIGINT":  syscall.SIGINT,
	"SIGKILL": syscall.SIGKILL,
	"SIGPIPE": syscall.SIGPIPE,
	"SIGQUIT": syscall.SIGQUIT,
	"SIGSEGV": syscall.SIGSEGV,
	"SIGTERM": syscall.SIGTERM,
	"SIGUSR1": syscall.SIGUSR1,
	"SIGUSR2": syscall.SIGUSR2,
}

func init() {
	flag.StringVar(&signalName, "S", "SIGTERM", "signal name to send")
}

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}
	if len(flag.Args()) == 0 {
		log.Fatal("Usage: kill [-s sigspec] pid|pgid|pname")
	}

	sig, ok := signalMap[signalName]
	if !ok {
		log.Fatalf("unknown signal %s", signalName)
	}

	target := flag.Arg(0)
	pid, err := strconv.Atoi(target)

	if err == nil && pid < 0 {
		err = syscall.Kill(-pid, sig) // negative pid is a process group
		if err != nil {
			log.Fatalf("failed to send signal to process group: %v", err)
		}
		fmt.Printf("sent signal %s to the process group %d", signalName, pid)
	} else if err == nil {
		// pid is a process
		process, err := os.FindProcess(pid)
		if err != nil {
			log.Fatalf("Failed to find process: %v", err)
		}

		err = process.Signal(sig)
		if err != nil {
			log.Fatalf("failed to send signal to process: %d\n", err)
		}
	} else {
		// if pid is not a number, assume it's a process name
		processes, err := px.Processes()
		if err != nil {
			log.Fatalf("failed to get processes: %v", err)
		}
		for _, process := range processes {
			if strings.ToLower(process.Executable()) == strings.ToLower(target) {
				proc, err := os.FindProcess(process.Pid())
				if err != nil {
					log.Printf("failed to find process: %v", err)
					continue
				}
				err = proc.Signal(sig)
				if err != nil {
					log.Printf("failed to send signal to process: %v", err)
				}
				fmt.Printf("sent signal %s to process %d\n", signalName, proc.Pid)
			}
		}
	}
}
