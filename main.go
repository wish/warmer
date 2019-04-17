package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"syscall"
	"time"

	"github.com/frostschutz/go-fibmap"
	flags "github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
)

type File struct {
	path  string
	info  os.FileInfo
	order uint64
}

type Files []File

func (f Files) Len() int           { return len(f) }
func (f Files) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f Files) Less(i, j int) bool { return f[i].order < f[j].order }

const SSIZE_MAX = 9223372036854775807

type CLIOpts struct {
	Root            string `long:"root" short:"r" description:"root path to walk from" required:"true"`
	EnableHibernate bool   `long:"enable-hibernate" description:"sleep forever after finishing, useful if this is a sidecare container"`
}

func main() {
	var opts CLIOpts
	parser := flags.NewParser(&opts, flags.Default)
	if _, err := parser.Parse(); err != nil {
		// If the error was from the parser, then we can simply return
		// as Parse() prints the error already
		if _, ok := err.(*flags.Error); ok {
			os.Exit(1)
		}
		logrus.Fatalf("Error parsing flags: %v", err)
	}

	files := Files{}
	mu := &sync.Mutex{}
	err := filepath.Walk(filepath.Clean(flags.Root), func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		exts, errno := fibmap.NewFibmapFile(f).Fiemap(32)
		if errno != 0 {
			return fmt.Errorf("fiemap error")
		}
		if len(exts) < 1 {
			fmt.Printf("no exts found: %v", path)
			return nil
		}
		order := exts[0].Physical
		if order == 0 {
			stat, ok := info.Sys().(*syscall.Stat_t)
			if !ok {
				return fmt.Errorf("stat error")
			}
			order = stat.Ino
		}
		mu.Lock()
		files = append(files, File{
			path:  path,
			order: order,
			info:  info,
		})
		mu.Unlock()
		return nil
	})
	if err != nil {
		panic(err)
	}
	sort.Sort(files)
	ch := make(chan File, len(files))
	for _, file := range files {
		ch <- file
	}
	wg := &sync.WaitGroup{}
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go worker(wg, ch)
	}
	wg.Wait()
	if opts.EnableHibernate {
		for {
			time.Sleep(time.Hour)
		}
	}
}

func worker(wg *sync.WaitGroup, files chan File) {
	for true {
		select {
		case file := <-files:
			chunks, err := sendfile(file.path, file.info)
			if err != nil {
				panic(err) // TODO(tvi): Handle better.
			}
			fmt.Printf("Done: %v @[block=%v chunks=%v]\n", file.path, file.order, chunks)
		default:
			wg.Done()
			return
		}
	}
}

func sendfile(path string, info os.FileInfo) (int, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return 0, err
	}
	null, err := os.OpenFile("/dev/null", os.O_WRONLY, os.ModePerm)
	if err != nil {
		return 0, err
	}
	offset := int64(0)
	ln := info.Size()
	chunks := 0
	for offset < ln {
		count := 0
		remaining := ln - offset
		if remaining > SSIZE_MAX {
			count = SSIZE_MAX
		} else {
			count = int(remaining)
		}
		_, err := syscall.Sendfile(int(null.Fd()), int(f.Fd()), &offset, count)
		if err != nil {
			return 0, fmt.Errorf("could not sendfile: %v", err)
		}
		chunks += 1
	}
	return chunks, nil
}
