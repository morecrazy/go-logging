package logging

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type FileLogWriter struct {
	filename string
	file     *os.File
	writeMtx *sync.Mutex

	rotate bool

	// rotate at size
	maxsize int64
	cursize int64

	// rotate hourly
	hourly   bool
	lasthour int
}

func NewFileLogWriter(filename string, rotate bool) (*FileLogWriter, error) {
	w := &FileLogWriter{
		filename: filename,
		rotate:   rotate,
		writeMtx: &sync.Mutex{},
	}
	// open the file for the first time
	if err := w.Rotate(); err != nil {
		fmt.Fprintf(os.Stderr, "FileLogWriter(%q): %s\n", w.filename, err)
		return nil, err
	}

	return w, nil
}

// If this is called in a threaded context, it MUST be synchronized
func (w *FileLogWriter) Rotate() error {
	// Close any log file that may be open
	if w.file != nil {
		// fmt.Fprint(w.file, fmt.Sprintf("file logger closed at %s", time.Now().String()))
		w.file.Close()
	}

	// If we are keeping log files, move it to the next available number
	if w.rotate {
		_, err := os.Lstat(w.filename)
		if err == nil { // file exists
			// Find the next available number
			num := 1
			fname := ""
			now := time.Now()
			for ; err == nil && num <= 999; num++ {
				fname = fmt.Sprintf("%s.%04d-%02d-%02d.%02d.%03d",
					w.filename, now.Year(), int(now.Month()), now.Day(), now.Hour(), num)
				_, err = os.Lstat(fname)
			}
			// return error if the last file checked still existed
			if err == nil {
				return fmt.Errorf("Rotate: Cannot find free log number to rename %s\n", w.filename)
			}

			// Rename the file to its newfound home
			err = os.Rename(w.filename, fname)
			if err != nil {
				return fmt.Errorf("Rotate: %s\n", err)
			}
		}
	}

	// Open the log file
	fd, err := os.OpenFile(w.filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		return err
	}
	w.file = fd

	// initialize rotation values
	w.cursize = 0

	// set log open hour
	w.lasthour = time.Now().Hour()

	return nil
}

func (w *FileLogWriter) needRotate() bool {
	if (w.maxsize > 0 && w.cursize >= w.maxsize) ||
		(w.hourly && w.lasthour != time.Now().Hour()) {
		return true
	}

	return false
}

func (w *FileLogWriter) SetRotateSize(maxsize int64) *FileLogWriter {
	w.maxsize = maxsize
	return w
}

func (w *FileLogWriter) SetRotateHourly(hourly bool) *FileLogWriter {
	w.hourly = true
	return w
}

func (w *FileLogWriter) Write(p []byte) (int, error) {
	w.writeMtx.Lock()
	defer w.writeMtx.Unlock()

	if w.needRotate() {
		if err := w.Rotate(); err != nil {
			fmt.Fprintf(os.Stderr, "FileLogWriter(%q): %s\n", w.filename, err)
			return 0, err
		}
	}

	// Perform the write
	n, err := w.file.Write(p)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FileLogWriter(%q): %s\n", w.filename, err)
		return n, err
	}

	w.cursize += int64(n)

	return n, err
}

func (w *FileLogWriter) Close() {
	w.file.Close()
}
