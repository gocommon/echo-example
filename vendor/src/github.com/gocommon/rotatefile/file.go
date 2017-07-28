package rotatefile

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Writer implements LoggerInterface.
// It writes messages by lines limit, file size limit, or time frequency.
type Writer struct {
	mw *MuxWriter
	// The opened file
	Filename string `json:"filename"`

	Maxlines         int `json:"maxlines"`
	maxlinesCurLines int

	// Rotate at size
	Maxsize        int `json:"maxsize"`
	maxsizeCurSize int

	// Rotate daily
	Daily         bool  `json:"daily"`
	Maxdays       int64 `json:"maxdays"`
	dailyOpendate int

	Rotate bool `json:"rotate"`

	startLock sync.Mutex // Only one log can write to the file

}

// Options Options
type Options struct {
	Filename string `json:"filename"`
	Maxlines int    `json:"maxlines"`
	Maxsize  int    `json:"maxsize"`
	Daily    bool   `json:"daily"`
	Maxdays  int64  `json:"maxdays"`
	Rotate   bool   `json:"rotate"`
}

// MuxWriter an *os.File writer with locker.
type MuxWriter struct {
	sync.Mutex
	fd *os.File
}

// write to os.File.
func (l *MuxWriter) Write(b []byte) (int, error) {
	l.Lock()
	defer l.Unlock()
	return l.fd.Write(b)
}

// SetFd set os.File in writer.
func (l *MuxWriter) SetFd(fd *os.File) {
	if l.fd != nil {
		l.fd.Close()
	}
	l.fd = fd
}

// NewWriter create a Writer returning as LoggerInterface.
func NewWriter(opt ...Options) (*Writer, error) {
	w := &Writer{
		Filename: "log/rotatefile.log",
		Maxlines: 1000000,
		Maxsize:  1 << 28, //256 MB
		Daily:    true,
		Maxdays:  7,
		Rotate:   true,
	}
	// use MuxWriter instead direct use os.File for lock write when rotate
	w.mw = new(MuxWriter)

	if len(opt) > 0 {
		w.setOptions(opt[0])
	}

	err := w.start()
	if err != nil {
		return nil, err
	}

	return w, nil
}

func (w *Writer) setOptions(opt Options) {
	if len(opt.Filename) > 0 {
		w.Filename = opt.Filename
	}

	if opt.Maxlines > 0 {
		w.Maxlines = opt.Maxlines
	}

	if opt.Maxsize > 0 {
		w.Maxsize = opt.Maxsize
	}

	w.Daily = opt.Daily
	w.Rotate = opt.Rotate

	if opt.Maxdays > 0 {
		w.Maxdays = opt.Maxdays
	}
}

// Init file logger with json config.
// config like:
//	{
//	"filename":"log/gogs.log",
//	"maxlines":10000,
//	"maxsize":1<<30,
//	"daily":true,
//	"maxdays":15,
//	"rotate":true
//	}
func (w *Writer) init(config string) error {
	if err := json.Unmarshal([]byte(config), w); err != nil {
		return err
	}
	if len(w.Filename) == 0 {
		return errors.New("config must have filename")
	}
	return nil
}

// start file logger. create log file and set to locker-inside file writer.
func (w *Writer) start() error {
	fd, err := w.createFile()
	if err != nil {
		return err
	}
	w.mw.SetFd(fd)
	if err = w.initFd(); err != nil {
		return err
	}
	return nil
}

func (w *Writer) docheck(size int) {
	w.startLock.Lock()
	defer w.startLock.Unlock()
	if w.Rotate && ((w.Maxlines > 0 && w.maxlinesCurLines >= w.Maxlines) ||
		(w.Maxsize > 0 && w.maxsizeCurSize >= w.Maxsize) ||
		(w.Daily && time.Now().Day() != w.dailyOpendate)) {
		if err := w.DoRotate(); err != nil {
			fmt.Fprintf(os.Stderr, "Writer(%q): %s\n", w.Filename, err)
			return
		}
	}
	w.maxlinesCurLines++
	w.maxsizeCurSize += size
}

// Write io.Writer
func (w *Writer) Write(b []byte) (int, error) {

	w.docheck(len(b))

	return w.mw.Write(b)
}

func (w *Writer) createFile() (*os.File, error) {
	os.MkdirAll(path.Dir(w.Filename), os.ModePerm)
	// Open the log file
	return os.OpenFile(w.Filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
}

func (w *Writer) initFd() error {
	fd := w.mw.fd
	finfo, err := fd.Stat()
	if err != nil {
		return fmt.Errorf("get stat: %s", err)
	}
	w.maxsizeCurSize = int(finfo.Size())
	w.dailyOpendate = time.Now().Day()
	if finfo.Size() > 0 {
		content, err := ioutil.ReadFile(w.Filename)
		if err != nil {
			return err
		}
		w.maxlinesCurLines = len(strings.Split(string(content), "\n"))
	} else {
		w.maxlinesCurLines = 0
	}
	return nil
}

// DoRotate means it need to write file in new file.
// new file name like xx.log.2013-01-01.2
func (w *Writer) DoRotate() error {
	_, err := os.Lstat(w.Filename)
	if err == nil { // file exists
		// Find the next available number
		num := 1
		fname := ""
		for ; err == nil && num <= 999; num++ {
			fname = w.Filename + fmt.Sprintf(".%s.%03d", time.Now().Format("2006-01-02"), num)
			_, err = os.Lstat(fname)
		}
		// return error if the last file checked still existed
		if err == nil {
			return fmt.Errorf("rotate: cannot find free log number to rename %s", w.Filename)
		}

		// block Logger's io.Writer
		w.mw.Lock()
		defer w.mw.Unlock()

		fd := w.mw.fd
		fd.Close()

		// close fd before rename
		// Rename the file to its newfound home
		if err = os.Rename(w.Filename, fname); err != nil {
			return fmt.Errorf("Rotate: %s", err)
		}

		// re-start logger
		if err = w.start(); err != nil {
			return fmt.Errorf("Rotate StartLogger: %s", err)
		}

		go w.deleteOldLog()
	}

	return nil
}

func (w *Writer) deleteOldLog() {
	dir := filepath.Dir(w.Filename)
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) (returnErr error) {
		defer func() {
			if r := recover(); r != nil {
				returnErr = fmt.Errorf("Unable to delete old log '%s', error: %+v", path, r)
			}
		}()

		if !info.IsDir() && info.ModTime().Unix() < (time.Now().Unix()-60*60*24*w.Maxdays) {
			if strings.HasPrefix(filepath.Base(path), filepath.Base(w.Filename)) {
				os.Remove(path)
			}
		}
		return returnErr
	})
}

// Close file logger, close file writer.
func (w *Writer) Close() {
	w.mw.fd.Close()
}

// Sync flush file means sync file from disk.
func (w *Writer) Sync() {
	w.mw.fd.Sync()
}
