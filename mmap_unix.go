// Copyright 2011 Evan Shaw. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE-MMAP-GO file.

//go:build darwin || dragonfly || freebsd || linux || openbsd || solaris || netbsd
// +build darwin dragonfly freebsd linux openbsd solaris netbsd

// Modifications (c) 2017 The Memory Authors.

package memory // import "modernc.org/memory"

import (
	"os"
	"syscall"
)

const pageSizeLog = 20

var (
	osPageMask = osPageSize - 1
	osPageSize = os.Getpagesize()
)

func unmap(addr uintptr, size int) error {
	_, _, errno := syscall.Syscall(syscall.SYS_MUNMAP, addr, uintptr(size), 0)
	if errno != 0 {
		return errno
	}

	return nil
}

// pageSize aligned.
func mmap(size int) (uintptr, int, error) {
	size = roundup(size, osPageSize)
	// Copied from https://cs.opensource.google/go/go/+/refs/tags/go1.17.8:src/syscall/zsyscall_linux_amd64.go;l=1575-1584
	p, _, err := syscall.Syscall6(syscall.SYS_MMAP, 0, uintptr(size+pageSize), uintptr(syscall.PROT_READ|syscall.PROT_WRITE), uintptr(syscall.MAP_PRIVATE|syscall.MAP_ANON), ^uintptr(0) /* -1 */, 0)
	if err != 0 {
		return 0, 0, err
	}

	n := size + pageSize
	if p&uintptr(osPageMask) != 0 {
		panic("internal error")
	}

	mod := int(p) & pageMask
	if mod != 0 {
		m := pageSize - mod
		if err := unmap(p, m); err != nil {
			return 0, 0, err
		}

		n -= m
		p += uintptr(m)
	}

	if p&uintptr(pageMask) != 0 {
		panic("internal error")
	}

	if n-size != 0 {
		if err := unmap(p+uintptr(size), n-size); err != nil {
			return 0, 0, err
		}
	}

	return p, size, nil
}
