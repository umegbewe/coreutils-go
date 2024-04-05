package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/encoding/charmap"
)

type Options struct {
	BlockSize       int
	Count           int64
	IfFile          string
	OfFile          string
	Seek            int64
	Skip            int64
	Conv            string
	Progress        bool
	InputBlockSize  int
	OutputBlockSize int
	ConvBlockSize   int
	InputFlags      string
	OutputFlags     string
	Dereference     bool
}

var log = logrus.New()

type InputError struct {
	Err error
}

func (e *InputError) Error() string {
	return fmt.Sprintf("input error: %v", e.Err)
}

type OutputError struct {
	Err error
}

func (e *OutputError) Error() string {
	return fmt.Sprintf("output error: %v", e.Err)
}

type ConversionError struct {
	Err error
}

func (e *ConversionError) Error() string {
	return fmt.Sprintf("conversion error: %v", e.Err)
}

func handleError(err error) {
	if err == nil {
		return
	}

	var exitCode int

	switch errors.Cause(err).(type) {
	case *InputError:
		log.WithError(err).Error("Input error occurred")
		exitCode = 3
	case *OutputError:
		log.WithError(err).Error("Output error occurred")
		exitCode = 4
	case *ConversionError:
		log.WithError(err).Error("Conversion error occurred")
		exitCode = 5
	default:
		log.WithError(err).Error("An error occurred")
		exitCode = 1
	}

	os.Exit(exitCode)
}

func main() {
	opts, err := ParseOptions()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	if err := copy(opts); err != nil {
		handleError(err)
	}
}

type BlockSize struct {
	Value *int
}

func (b *BlockSize) String() string {
	return fmt.Sprintf("%d", *b.Value)
}
func (b *BlockSize) Set(s string) error {
	var multiplier int
	var value int

	switch suffix := s[len(s)-1]; suffix {
	case 'k', 'K':
		multiplier = 1024
		value, _ = strconv.Atoi(s[:len(s)-1])
	case 'm', 'M':
		multiplier = 1024 * 1024
		value, _ = strconv.Atoi(s[:len(s)-1])
	case 'g', 'G':
		multiplier = 1024 * 1024 * 1024
		value, _ = strconv.Atoi(s[:len(s)-1])
	default:
		multiplier = 1
		value, _ = strconv.Atoi(s)
	}

	*b.Value = value * multiplier
	return nil
}

func ParseOptions() (*Options, error) {
	opts := &Options{}

	flag.Var(&BlockSize{&opts.BlockSize}, "bs", "block size")
	flag.Int64Var(&opts.Count, "count", 0, "number of blocks to copy")
	flag.StringVar(&opts.IfFile, "if", "", "input file")
	flag.StringVar(&opts.OfFile, "of", "", "output file")
	flag.Int64Var(&opts.Seek, "seek", 0, "skip output blocks")
	flag.Int64Var(&opts.Skip, "skip", 0, "skip input blocks")
	flag.StringVar(&opts.Conv, "conv", "", "convert the file as per the comma-separated symbol list")
	flag.BoolVar(&opts.Progress, "progress", true, "display progress")
	flag.IntVar(&opts.ConvBlockSize, "cbs", 0, "conversion block size")
	flag.Parse()

	if opts.IfFile == "" {
		return nil, fmt.Errorf("missing input file")
	}
	if opts.OfFile == "" {
		return nil, fmt.Errorf("missing output file")
	}

	return opts, nil
}


func copy(opts *Options) error {
	var in *os.File
	var err error

	if opts.IfFile == "-" {
		in = os.Stdin
	} else {
		in, err = os.Open(opts.IfFile)
		if err != nil {
			return fmt.Errorf("failed to open input file: %v", err)
		}
		defer in.Close()

		inInfo, err := os.Stat(opts.IfFile)
		if err != nil {
			return fmt.Errorf("failed to get input file info: %v", err)
		}

		// Handle device files
		if inInfo.Mode()&os.ModeDevice != 0 {
			in, err = os.OpenFile(opts.IfFile, os.O_RDWR, 0)
			if err != nil {
				return fmt.Errorf("failed to open device file: %v", err)
			}
			defer in.Close()
		}

		// Handle named pipes (FIFOs)
		if inInfo.Mode()&os.ModeNamedPipe != 0 {
			in, err = os.OpenFile(opts.IfFile, os.O_RDONLY, 0)
			if err != nil {
				return fmt.Errorf("failed to open named pipe: %v", err)
			}
			defer in.Close()
		}

		// Handle symbolic links
		if inInfo.Mode()&os.ModeSymlink != 0 {
			if opts.Dereference {
				targetPath, err := os.Readlink(opts.IfFile)
				if err != nil {
					return fmt.Errorf("failed to read symbolic link: %v", err)
				}
				in, err = os.Open(targetPath)
				if err != nil {
					return fmt.Errorf("failed to open target file: %v", err)
				}
				defer in.Close()
			} else {
				err := os.Symlink(opts.IfFile, opts.OfFile)
				if err != nil {
					return fmt.Errorf("failed to create symbolic link: %v", err)
				}
				return nil
			}
		}
	}

	var out *os.File

	if opts.OfFile == "-" {
		out = os.Stdout
	} else {
		out, err = os.OpenFile(opts.OfFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("failed to open output file: %v", err)
		}
		defer out.Close()
	}

	if opts.Skip > 0 {
		if _, err := in.Seek(opts.Skip*int64(opts.BlockSize), io.SeekStart); err != nil {
			return fmt.Errorf("failed to skip input blocks: %v", err)
		}
	}

	if opts.Seek > 0 {
		if _, err := out.Seek(opts.Seek*int64(opts.BlockSize), io.SeekStart); err != nil {
			return fmt.Errorf("failed to seek output blocks: %v", err)
		}
	}

	buf := make([]byte, opts.BlockSize)
	var tBytes int64
	var fullBlocks, partialBlocks int64

	start := time.Now()

	for {
		if opts.Count > 0 && tBytes >= opts.Count*int64(opts.BlockSize) {
			break
		}

		n, err := io.ReadFull(in, buf)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				if n > 0 {
					partialBlocks++
				}
				break
			}
			return &InputError{Err: fmt.Errorf("failed to read input file: %v", err)}
		}

		if opts.Conv != "" {
			var err error
			buf, err = conversions(buf[:n], opts.Conv, opts)
			if err != nil {
				return &ConversionError{Err: fmt.Errorf("failed to apply conversions: %v", err)}
			}
			n = len(buf)
		}

		written := 0
		for written < n {
			m, err := out.Write(buf[written:n])
			if err != nil {
				return &OutputError{Err: fmt.Errorf("failed to write output file: %v", err)}
			}
			written += m
		}

		tBytes += int64(n)
		if n == len(buf) {
			fullBlocks++
		} else {
			partialBlocks++
		}

	}

	dur := time.Since(start)

	fmt.Fprintf(os.Stderr, "%d+%d records in\n", fullBlocks, partialBlocks)
	fmt.Fprintf(os.Stderr, "%d+%d records out\n", fullBlocks, partialBlocks)
	fmt.Fprintf(os.Stderr, "%d bytes (%s) copied, %.4f s, %.0f MB/s\n",
		tBytes, humanize(tBytes), dur.Seconds(), float64(tBytes)/dur.Seconds()/1024/1024)

	return nil
}

func conversions(buf []byte, conv string, opts *Options) ([]byte, error) {
	switch conv {
	case "ascii":
		return charmap.ISO8859_1.NewDecoder().Bytes(buf)
	case "ebcdic":
		return charmap.CodePage037.NewEncoder().Bytes(buf)
	case "ibm":
		return charmap.CodePage037.NewEncoder().Bytes(buf)
	case "block":
		// pad buf to the speccified conversion block size
		padLength := opts.ConvBlockSize - len(buf)
		if padLength > 0 {
			buf = append(buf, bytes.Repeat([]byte{0}, padLength)...)
		}
		return buf, nil
	case "unblock":
		// trim trailing NUL characters from buf
		return bytes.TrimRight(buf, "\x00"), nil
	case "lcase":
		return bytes.ToLower(buf), nil
	case "ucase":
		return bytes.ToUpper(buf), nil
	case "sparse":
		if info, err := os.Stat(opts.IfFile); err == nil && info.Size() > int64(len(buf)) {
			out, err := os.Create(opts.OfFile)
			if err != nil {
				return nil, fmt.Errorf("failed to create sparse output file: %v", err)
			}
			defer out.Close()

			if _, err := out.Write(buf); err != nil {
				return nil, fmt.Errorf("failed to write to sparse output file: %v", err)
			}

			if err := out.Truncate(info.Size()); err != nil {
				return nil, fmt.Errorf("failed to create sparse output file: %v", err)
			}
		}
		return buf, nil
	case "swab":
		for i := 0; i < len(buf)-1; i += 2 {
			buf[i], buf[i+1] = buf[i+1], buf[i]
		}
		return buf, nil
	case "sync":
		// pad buf to specified input block size
		padLength := opts.InputBlockSize - len(buf)
		if padLength > 0 {
			buf = append(buf, bytes.Repeat([]byte{0}, padLength)...)
		}
		return buf, nil
	case "excl":
		if _, err := os.Stat(opts.OfFile); err == nil {
			return nil, fmt.Errorf("output file '%s' already exists", opts.OfFile)
		}
		return buf, nil
	case "nocreat":
		// open output file with the O_EXCL flag to ensure it doesn't create a new file
		out, err := os.OpenFile(opts.OfFile, os.O_WRONLY|os.O_EXCL, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open output file: %v", err)
		}
		defer out.Close()
		return buf, nil
	case "fdatasync":
		// flush written data to underlying storage device
		out, err := os.OpenFile(opts.OfFile, os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open output file for fdatasync: %v", err)
		}
		defer out.Close()

		if err := syscall.Fdatasync(int(out.Fd())); err != nil {
			return nil, fmt.Errorf("failed to perform fdatasync: %v", err)
		}
		return buf, nil
	case "fsync":
		// flush written data and metadta to the underlying storage device
		out, err := os.OpenFile(opts.OfFile, os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open output file for fsync: %v", err)
		}
		defer out.Close()

		if err := syscall.Fsync(int(out.Fd())); err != nil {
			return nil, fmt.Errorf("failed to perform fsync: %v", err)
		}
		return buf, nil
	case "notrunc":
		outFile, err := os.OpenFile(opts.OfFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open output file: %v", err)
		}
		defer outFile.Close()
		return buf, nil
	case "noerror":
		// skip and continue process
		return buf, nil
	default:
		return buf, nil
	}
}

func humanize(bytes int64) string {
    sizes := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}
    if bytes == 0 {
        return "0 B"
    }
    base := int64(1024)
    i := 0
    for bytes >= base {
        bytes /= base
        i++
    }
    return fmt.Sprintf("%d %s", bytes, sizes[i])
}