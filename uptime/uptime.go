package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type LoadAvg struct {
	One     float64
	Five    float64
	Fifteen float64
}

const (
	UT_LINESIZE = 32
	UT_NAMESIZE = 32
	UT_HOSTSIZE = 256
)

// utmp record types
const (
	EMPTY         = 0
	RUN_LVL       = 1
	BOOT_TIME     = 2
	NEW_TIME      = 3
	OLD_TIME      = 4
	INIT_PROCESS  = 5
	LOGIN_PROCESS = 6
	USER_PROCESS  = 7
	DEAD_PROCESS  = 8
	ACCOUNTING    = 9
)

type utmp struct {
	UtType int16
	UtPid  int32
	UtLine [UT_LINESIZE]uint8
	UtID   [4]uint8
	UtUser [UT_NAMESIZE]uint8
	UtHost [UT_HOSTSIZE]uint8
	UtExit struct {
		ETermination int16
		EExit        int16
	}
	UtSession int32
	UtTv      struct {
		TvSec  int32
		TvUsec int32
	}
	UtAddrV6 [4]int32
	Unused   [20]uint8
}

func main() {

	now := time.Now()

	uptime, err := getUptime()
	if err != nil {
		log.Fatal(err)
	}

	users, err := getUsers()
	if err != nil {
		log.Fatal(err)
	}

	loadavg, err := getLoadAvg()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf(" %v up %v,  %v,  load average: %v, %v, %v\n",
		now.Format("15:04:05"),
		uptime,
		users,
		loadavg.One,
		loadavg.Five,
		loadavg.Fifteen,
	)
}

func getUptime() (time.Duration, error) {
	/* The /proc/uptime file contains the system uptime in seconds
	 * the uptime is the first field in the file, followed by the idle time in seconds}
	 */
	data, err := ioutil.ReadFile("/proc/uptime")
	if err != nil {
		return 0, err
	}

	fields := strings.Fields(string(data))
	if len(fields) < 1 {
		return 0, fmt.Errorf("unexpected format in /proc/uptime")
	}

	uptimeSecs, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, fmt.Errorf("parsing uptime: %w", err)
	}

	return time.Duration(uptimeSecs) * time.Second, nil
}

func getUsers() (int, error) {

	// Okay this is longgg and might be stupid to put here, but might also help someone

	/*
	   	utmp is a "login records file" i.e keeps track of all logins and logouts of the systems
	   	Each record in the utmp file corresponds to a terminal connected to the system, whether it's a physical terminal or a virtual one (like SSH).
	   	When a user logs in or out, the system writes a new record to utmp.

	   	The structure of each utmp record is defined in the utmp.h header file in the C library.
	   	It includes several fields, like the type of the record, the PID of the login process, the username, the hostname for remote logins etc
	   	This is the structure that my Go program is reading, represented at the top of this file (although a simplified version, utmp struct can be different across Linux distros)
	   /*

	   /*The structure of each utmp record is defined in the utmp.h header file in the C library

	   But here is one from the man page (https://man7.org/linux/man-pages/man5/wtmp.5.html):

	// struct utmp {
	// 	short   ut_type;              /* Type of record */
	// 	pid_t   ut_pid;               /* PID of login process */
	// 	char    ut_line[UT_LINESIZE]; /* Device name of tty - "/dev/" */
	// 	char    ut_id[4];             /* Terminal name suffix, or inittab(5) ID */
	// 	char    ut_user[UT_NAMESIZE]; /* Username */
	// 	char    ut_host[UT_HOSTSIZE]; /* Hostname for remote login, or kernel version for run-level messages */
	// 	struct  exit_status ut_exit;  /* Exit status of a process marked as DEAD_PROCESS; not used by Linux init(8) */
	// 	/* The ut_session and ut_tv fields must be the same size when compiled 32- and 64-bit.
	// 	   This allows data files and shared memory to be shared between 32- and 64-bit applications. */
	// 	#if __WORDSIZE == 64 && defined __WORDSIZE_COMPAT32
	// 	int32_t ut_session;           /* Session ID (getsid(2)), used for windowing */
	// 	struct {
	// 		int32_t tv_sec;           /* Seconds */
	// 		int32_t tv_usec;          /* Microseconds */
	// 	} ut_tv;                      /* Time entry was made */
	// 	#else
	// 	 long   ut_session;           /* Session ID */
	// 	 struct timeval ut_tv;        /* Time entry was made */
	// 	#endif
	// 	int32_t ut_addr_v6[4];        /* Internet address of remote host; IPv4 address uses just ut_addr_v6[0] */
	// 	char    __unused[20];         /* Reserved for future use */
	//  };

	/* Using the binary.Read function from the encoding/binary package to read an utmp record from the file
	   all that's going on here is reading data from the reader and decoding it into the given variable, according to its type.

	   Checking the UtType field of each record to see if it's USER_PROCESS, which represents a logged-in user.
	   If it is, we increment our user count. This gives us a count of the currently logged-on users, as represented in the /var/run/utmp file.

	   One complexity here is that we need to account for different types of logins (for example, multiple logins by the same user or logins from system processes)
	   which is why we specifically check for USER_PROCESS. This ensures we're only counting "real" user sessions.

	   PS: Utmp isn't meant to be secured as it can be manipulated by users with sufficient privileges, crazy right?
	   	For secure logging, systems typically use the wtmp and btmp files, which keep track of all logins and logouts (wtmp) and all failed login attempts (btmp),

	   Thanks for reading.
	*/

	file, err := os.Open("/var/run/utmp")
	if err != nil {
		return 0, err
	}
	defer file.Close()

	users := 0
	reader := bufio.NewReader(file)
	var record utmp

	for {
		// Read an utmp record. If we're at EOF, break the loop
		err = binary.Read(reader, binary.LittleEndian, &record)
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}

		// If the ut_type field of the record is USER_PROCESS, increment the user count
		if record.UtType == USER_PROCESS {
			users++
		}
	}

	return users, nil
}

func getLoadAvg() (*LoadAvg, error) {
	file, err := os.Open("/proc/loadavg")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var load LoadAvg
	_, err = fmt.Fscanf(file, "%f %f %f", &load.One, &load.Five, &load.Fifteen)
	if err != nil {
		return nil, err
	}

	return &load, nil
}
