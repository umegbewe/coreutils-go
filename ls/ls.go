package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	lFlag   = flag.Bool("l", false, "use a long listing format")
	aFlag   = flag.Bool("a", false, "do not ignore entries starting with .")
	hFlag   = flag.Bool("h", false, "with -l and -s, print sizes like 1K 234M 2G etc.")
	RFlag   = flag.Bool("R", false, "list subdirectories recursively")
	tFlag   = flag.Bool("t", false, "sort by modification time, newest first;")
	dFlag   = flag.Bool("d", false, "list directories themselves, not their contents")
	oneFlag = flag.Bool("1", false, "list one file per line")
	rFlag   = flag.Bool("r", false, "reverse order while sorting")
)

func quote(s string) string {
	if strings.ContainsAny(s, " \t\n\"'\\") {
		return strconv.Quote(s)
	}
	return s
}

func formatSize(size int64) string {
	if *hFlag {
		const unit = 1024
		if size < unit {
			return fmt.Sprintf("%d", size)
		}
		div, exp := int64(unit), 0
		for n := size / unit; n >= unit; n /= unit {
			div *= unit
			exp++
		}
		return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
	} else {
		return fmt.Sprintf("%d", size)
	}
}

func formatFile(file os.FileInfo, path string) string {
	linkInfo := ""
	if file.Mode().IsRegular() && file.Mode()&os.ModeSymlink != 0 {
		linkDest, err := os.Readlink(path)
		if err != nil {
			linkDest = "???"
		}
		linkInfo = fmt.Sprintf(" -> %s", linkDest)
	}

	if *lFlag || *oneFlag {
		uid := file.Sys().(*syscall.Stat_t).Uid
		gid := file.Sys().(*syscall.Stat_t).Gid
		u, _ := user.LookupId(strconv.Itoa(int(uid)))
		g, _ := user.LookupGroupId(strconv.Itoa(int(gid)))
		userName := "-"
		groupName := "-"
		if u != nil {
			userName = u.Username
		}

		if g != nil {
			groupName = g.Name
		}
		return fmt.Sprintf(
			"%v %d %s %s %s %s %s%s",
			file.Mode(),
			file.Sys().(*syscall.Stat_t).Nlink,
			userName,
			groupName,
			formatSize(file.Size()),
			file.ModTime().Format(time.Stamp),
			quote(file.Name()),
			linkInfo,
		)

	} else {
		return quote(file.Name()) + linkInfo
	}
}

func ls(dirname string, recursive bool) {
	file, err := os.Open(dirname)
	if err != nil {
		fmt.Println(err)
		return
	}

	files, err := file.Readdir(-1) // -1 means read all files
	if err != nil {
		fmt.Println(err)
		return
	}

	if *dFlag {
		info, err := os.Stat(dirname)
		if err != nil {
			fmt.Println(err)
			return
		}
		files = []os.FileInfo{info}
	}

	if *tFlag {
		sort.Slice(files, func(i, j int) bool {
			return files[i].ModTime().After(files[j].ModTime())
		})
	}

	if *rFlag {
		sort.Slice(files, func(i, j int) bool {
			return files[i].ModTime().Before(files[j].ModTime())
		})
	}

	for _, file := range files {
		if *aFlag || !isHidden(file) {
			fmt.Println(formatFile(file, filepath.Join(dirname, file.Name())))
		}

		if recursive && file.IsDir() {
			fmt.Printf("\n%s:\n", filepath.Join(dirname, file.Name()))
			ls(filepath.Join(dirname, file.Name()), true)
		}
	}
}

func isHidden(file os.FileInfo) bool {
	return file.Name()[0] == '.'
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		ls(".", *RFlag)
	} else {
		for _, arg := range args {
			ls(arg, *RFlag)
		}
	}
}
