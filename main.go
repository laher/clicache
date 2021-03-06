package main

import (
	"flag"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

var (
	isHelp    = flag.Bool("h", false, "Show this help")
	isDel     = flag.Bool("del", false, "delete this entry")
	isVerbose = flag.Bool("v", false, "verbose")
	isHashCwd = flag.Bool("c", false, "include cwd in hash")
	maxDur    = flag.String("t", "5m", "max duration to cache output (cache keys are rounded by this amount)")
	dir       = flag.String("dir", os.Getenv("HOME")+"/.cache/clicache", "directory to store/retrieve cache info")
)

// TODO: os-independent HOME-dir (currently Windows not supported AFAIK)
// TODO: (optionally) include CWD in hash
// TODO: delete-all cache entries
// MAYBE: clear old entries automatically
func main() {
	flag.Parse()
	var (
		args   = flag.Args()
		hashed string
	)
	if *isHelp || len(args) < 1 {
		fmt.Println("clicache caches the STDOUT of a given command")
		fmt.Println("")
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *isHashCwd {
		wd, err := os.Getwd()
		if err != nil {
			if *isVerbose {
				log.Printf("Working dir error: %s", err)
			}
			os.Exit(1)
		}
		hashed = hash(append([]string{wd}, args...))
	} else {
		hashed = hash(args)
	}
	maxDuration, err := time.ParseDuration(*maxDur)
	if err != nil {
		if *isVerbose {
			log.Printf("Cache file error: %s", err)
		}
		os.Exit(1)
	}
	filename, tmpFilename := file(hashed, time.Now(), maxDuration)
	if _, err := os.Stat(filename); err != nil {
		if !os.IsNotExist(err) {
			//exit
			if *isVerbose {
				log.Printf("Cache file error: %s", err)
			}
			os.Exit(1)
		}
		if *isDel {
			// OK
			os.Exit(0)
		}
		// new file
		err := os.MkdirAll(*dir, 0700)
		if err != nil {
			if *isVerbose {
				log.Printf("Error (create dir): %s", err)
			}
			os.Exit(1)
		}

		// redirect IO
		file, err := os.Create(tmpFilename)
		if err != nil {
			if *isVerbose {
				log.Printf("Error (create file): %s", err)
			}
			os.Exit(1)
		}
		defer file.Close()
		// tee output to file
		ret, err := run(args, file)
		if err != nil || ret != 0 {
			if *isVerbose {
				log.Printf("Error (exit code %d): %v", ret, err)
			}
			err := file.Sync()
			log.Printf("Error (while synching file): %v", err)
			err = file.Close()
			log.Printf("Error (while closing file): %v", err)

			err = os.Remove(tmpFilename)
			log.Printf("Error (deleting file): %v", err)
			os.Exit(ret)
		}
		err = os.Rename(tmpFilename, filename)
		if err != nil {
			log.Printf("Error (moving file into place): %v", err)
			os.Exit(1)
		}
		os.Exit(ret)
	} else {
		if *isDel {
			err := os.Remove(filename)
			if err != nil {
				os.Exit(1)
			}
			os.Exit(0)
		}
		// spit out file
		file, err := os.Open(filename)
		if err != nil {
			if *isVerbose {
				log.Printf("Error (open file): %s", err)
			}
			os.Exit(1)
		}
		defer file.Close()
		_, err = io.Copy(os.Stdout, file)
		if err != nil {
			if *isVerbose {
				log.Printf("Error (open file): %s", err)
			}
			os.Exit(1)
		}
	}
}

//non-cryptographic hash
func hash(args []string) string {
	s := strings.Join(args, "_")
	h := fnv.New64a()
	h.Write([]byte(s))
	return fmt.Sprintf("%d", h.Sum64())
}

func file(hash string, time time.Time, d time.Duration) (string, string) {
	t := time.Truncate(d)

	return fmt.Sprintf("%s/%s-%d.stdout", *dir, hash, t.Unix()), fmt.Sprintf("%s/%s-%d.stdout.tmp", *dir, hash, time.UnixNano())
}

func run(args []string, out io.Writer) (int, error) {
	if len(args) < 1 {
		return 1, fmt.Errorf("No command supplied")
	}
	p, err := exec.LookPath(args[0])
	if err != nil {
		log.Printf("Couldn't find exe %s - %s", p, err)
		return 1, err
	}
	cmd := exec.Command(args[0])
	cmd.Args = args
	if *isVerbose {
		log.Printf("Running cmd: %s", args)
	}

	multiwriter := io.MultiWriter(os.Stdout, out)
	cmd.Stdout = multiwriter
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err = cmd.Start()
	if err != nil {
		log.Printf("Launch error: %s", err)
		return 1, err
	}
	if *isVerbose {
		log.Printf("Waiting for command to finish...")
	}
	err = cmd.Wait()
	if err != nil {
		if *isVerbose {
			log.Printf("Command exited with error: %v", err)
		}
	} else {
		if *isVerbose {
			log.Printf("Command completed without error")
		}
	}
	if err != nil {
		if e2, ok := err.(*exec.ExitError); ok { // there is error code
			processState, ok2 := e2.Sys().(syscall.WaitStatus)
			if ok2 {
				errcode := processState.ExitStatus()
				log.Printf("%s returned exit status: %d", cmd.Args[0], errcode)
				return errcode, err
			}
		}
		return 1, err
	}
	return 0, nil
}
