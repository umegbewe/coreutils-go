package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"syscall"
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

	pid, err := strconv.Atoi(flag.Arg(0))
	if err != nil {
		log.Fatalf("Invalid pid %s", err)
	}

	sig, ok := signalMap[signalName]
	if !ok {
		log.Fatalf("unknown signal %s", signalName)
	}

	if pid < 0 {
		err = syscall.Kill(-pid, sig) // negative pid is a process group
		if err != nil {
			log.Fatalf("failed to send signal to process group: %v", err)
		}
	} else {
		process, err := os.FindProcess(pid)
		if err != nil {
			log.Fatalf("Failed to find process: %v", err)
		}

		err = process.Signal(sig)
		if err != nil {
			log.Fatalf("failed to send signal to process: %d\n", err)
		}
	}

	fmt.Printf("sent signal %s to the process %d", signalName, pid)
}
