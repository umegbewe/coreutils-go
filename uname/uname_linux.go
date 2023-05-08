package main

import "golang.org/x/sys/unix"

func toGoString(cString []byte) string {
	byteSlice := make([]byte, len(cString))
	for i, v := range cString {
		byteSlice[i] = byte(v)
	}
	nullTerminated := bytes.IndexByte(byteSlice, 0)
	if nullTerminated != -1 {
		byteSlice = byteSlice[:nullTerminated]
	}
	return string(byteSlice)
}
