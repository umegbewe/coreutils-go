/* In no way is this code ready, please don't use
*/
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
	"syscall"
	"time"
)

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

func getLoadAvg() (*syscall.Sysinfo_t, error) {
	// The Sysinfo_t struct contains the load averages, among other things
	var info syscall.Sysinfo_t

	if err := syscall.Sysinfo(&info); err != nil {
		return nil, err
	}

	return &info, nil
}
