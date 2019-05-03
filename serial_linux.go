/**
 * Copyright 2019 Weston Schmidt
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package serial provides a simple but usable way to interact with devices
// that have serial ports.
package serial

import (
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
)

var baudMap = map[int]uint32{
	50:      unix.B50,
	75:      unix.B75,
	110:     unix.B110,
	134:     unix.B134,
	150:     unix.B150,
	200:     unix.B200,
	300:     unix.B300,
	600:     unix.B600,
	1200:    unix.B1200,
	1800:    unix.B1800,
	2400:    unix.B2400,
	4800:    unix.B4800,
	9600:    unix.B9600,
	19200:   unix.B19200,
	38400:   unix.B38400,
	57600:   unix.B57600,
	115200:  unix.B115200,
	230400:  unix.B230400,
	460800:  unix.B460800,
	500000:  unix.B500000,
	576000:  unix.B576000,
	921600:  unix.B921600,
	1000000: unix.B1000000,
	1152000: unix.B1152000,
	1500000: unix.B1500000,
	2000000: unix.B2000000,
	2500000: unix.B2500000,
	3000000: unix.B3000000,
	3500000: unix.B3500000,
	4000000: unix.B4000000,
}

var dataBitsMap = map[byte]uint32{
	'5': unix.CS5,
	'6': unix.CS6,
	'7': unix.CS7,
	'8': unix.CS8,
}

var stopBitsMap = map[byte]uint32{
	'1': 0,
	'2': unix.CSTOPB,
}

var parityMap = map[byte]uint32{
	'N': 0,
	'O': unix.PARENB | unix.PARODD,
	'E': unix.PARENB,
}

// Serial structure
type Serial struct {
	Name string	// The filename of the serial port
	file *os.File
}

func (s *Serial) ioctl(req, arg uintptr) unix.Errno {
	if nil == s.file {
		return unix.EBADFD
	}

	_, _, errno := unix.Syscall(unix.SYS_IOCTL, s.file.Fd(), req, arg)

	return errno
}

func validateConfig(baud int, cfg string) (rate, flags uint32, err error) {
	if tmp, ok := baudMap[baud]; ok {
		rate = tmp
	} else {
		return 0, 0, fmt.Errorf("Invalid baud rate parameter.")
	}

	if tmp, ok := dataBitsMap[cfg[0]]; ok {
		flags |= tmp
	} else {
		return 0, 0, fmt.Errorf("Invalid data bits parameter.")
	}

	if tmp, ok := parityMap[cfg[1]]; ok {
		flags |= tmp
	} else {
		return 0, 0, fmt.Errorf("Invalid parity parameter.")
	}

	if tmp, ok := stopBitsMap[cfg[2]]; ok {
		flags |= tmp
	} else {
		return 0, 0, fmt.Errorf("Invalid parity parameter.")
	}

	return rate, flags, nil
}


// Close closes the serial port or returns an error if one happens
func (s *Serial) Close() error {
	if nil != s.file {
		s.file.Close()
		s.file = nil
	}

	return nil
}

// SetBaud sets the baud rate for the serial port as well as the rest of
// the configuration.  The configuration is a string in the form: '8N1' or
// similar.
func (s *Serial) SetBaud(baud int, cfg string) error {
	if nil == s.file {
		return fmt.Errorf("Serial port '%s' not open.", s.Name)
	}

	rate, flags, err := validateConfig(baud, cfg)
	if nil != err {
		return err
	}

	t := unix.Termios{
		Iflag:  unix.IGNPAR,
		Cflag:  unix.CREAD | unix.CLOCAL | rate | flags,
		Ispeed: rate,
		Ospeed: rate,
	}

	t.Cc[unix.VMIN] = 1
	t.Cc[unix.VTIME] = 4

	errno := s.ioctl(uintptr(unix.TCSETS), uintptr(unsafe.Pointer(&t)))

	if 0 != errno {
		return fmt.Errorf("ioctl( '%s', TCSETS, &t ) error: %d\n", s.Name, errno)
	}

	return unix.SetNonblock(int(s.file.Fd()), false)
}

// Open opens the specified file name for serial port access
func (s *Serial) Open() error {
	if nil != s.file {
		return fmt.Errorf("Serial port '%s' already open.", s.Name)
	}

	f, err := os.OpenFile(s.Name, unix.O_RDWR|unix.O_NOCTTY|unix.O_NONBLOCK, 0666)
	if nil != err {
		return err
	}
	s.file = f

	return nil
}

// Write an array of bytes and return the number of bytes written
func (s *Serial) Write(b []byte) (n int, err error) {
	if nil == s.file {
		return 0, fmt.Errorf("Serial port '%s' not open.", s.Name)
	}

	return s.file.Write(b)
}

// Read into the specified array of bytes and return the number of bytes written
func (s *Serial) Read(b []byte) (n int, err error) {
	if nil == s.file {
		return 0, fmt.Errorf("Serial port '%s' not open.", s.Name)
	}

	return s.file.Read(b)
}

// Flush any characters that may be in a incoming or outgoing buffer
func (s *Serial) Flush() error {
	if nil == s.file {
		return fmt.Errorf("Serial port '%s' not open.", s.Name)
	}

	errno := s.ioctl(uintptr(unix.TCFLSH), uintptr(unix.TCIOFLUSH))
	if 0 != errno {
		return errno
	}

	return nil
}

// SendBreak sends the serial break signal
func (s *Serial) SendBreak() error {
	if nil == s.file {
		return fmt.Errorf("Serial port '%s' not open.", s.Name)
	}

	errno := s.ioctl(uintptr(unix.TCSBRKP), uintptr(0))
	if 0 != errno {
		return errno
	}

	return nil
}
