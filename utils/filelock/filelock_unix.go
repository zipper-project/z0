// Copyright 2018 The zipper team Authors
// This file is part of the z0 library.
//
// The z0 library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The z0 library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the z0 library. If not, see <http://www.gnu.org/licenses/>.

// Package filelock provides portable file locking.
package filelock

import (
	"os"
	"syscall"
)

type unixLock struct {
	f *os.File
}

func (l *unixLock) Release() error {
	if err := l.set(false); err != nil {
		return err
	}
	return l.f.Close()
}

func (l *unixLock) set(lock bool) error {
	how := syscall.LOCK_UN
	if lock {
		how = syscall.LOCK_EX
	}
	return syscall.Flock(int(l.f.Fd()), how|syscall.LOCK_NB)
}

func newLock(fileName string) (Releaser, error) {
	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	l := &unixLock{f}
	err = l.set(true)
	if err != nil {
		f.Close()
		return nil, err
	}
	return l, nil
}
