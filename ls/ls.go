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

func formatPermissions(mode os.FileMode) string {
	var b strings.Builder

	b.WriteString(fileType(mode))
	b.WriteString(filePermissions(mode))

	if mode&os.ModeSticky != 0 {
		b.WriteRune('t')
	} else if mode&os.ModeSetuid != 0 {
		b.WriteRune('s')
	} else if mode&os.ModeSetgid != 0 {
		b.WriteRune('s')
	} else {
		b.WriteRune('-')
	}

	return b.String()
}

func fileType(mode os.FileMode) string {
	switch {
	case mode.IsRegular():
		return "-"
	case mode.IsDir():
		return "d"
	case mode&os.ModeSymlink != 0:
		return "l"
	case mode&os.ModeNamedPipe != 0:
		return "p"
	case mode&os.ModeSocket != 0:
		return "s"
	case mode&os.ModeDevice != 0:
		if mode&os.ModeCharDevice != 0 {
			return "c"
		}
		return "b"
	default:
		return "?"
	}
}

func filePermissions(mode os.FileMode) string {
	var perms [9]rune

	fillPerms(perms[:], 0, mode.Perm()>>6)
	fillPerms(perms[:], 3, (mode.Perm()>>3)&7)
	fillPerms(perms[:], 6, mode.Perm()&7)
	return string(perms[:])
}

func fillPerms(perms []rune, start int, p os.FileMode) {
	if p&4 != 0 {
		perms[start] = 'r'
	} else {
		perms[start] = '-'
	}
	if p&2 != 0 {
		perms[start+1] = 'w'
	} else {
		perms[start+1] = '-'
	}
	if p&1 != 0 {
		perms[start+2] = 'x'
	} else {
		perms[start+2] = '-'
	}
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
			formatPermissions(file.Mode()),
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

	totalSize := int64(0)

	if *aFlag {
		dot, err := os.Stat(".")
		if err != nil {
			fmt.Println(err)
			return
		}
		dotDot, err := os.Stat("..")
		if err != nil {
			fmt.Println(err)
			return
		}

		files = append([]os.FileInfo{dot, dotDot}, files...)
	}

	if *lFlag {
		for _, file := range files {
			totalSize += file.Sys().(*syscall.Stat_t).Blocks
		}
		fmt.Printf("total %d\n", totalSize)
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
			subDir := filepath.Join(dirname, file.Name())
			fmt.Printf("\n%s:\n", subDir)
			ls(subDir, recursive)
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
