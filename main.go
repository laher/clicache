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
	maxDur    = flag.String("t", "5m", "max duration to cache output (cache keys are rounded by this amount)")
	dir       = flag.String("dir", os.Getenv("HOME")+"/.cache/clicache", "directory to store/retrieve cache info")
)

// TODO: os-independent HOME-dir
// TODO: (optionally) include CWD in hash
func main() {
	flag.Parse()
	if *isHelp {
		flag.PrintDefaults()
		os.Exit(1)
	}
	args := flag.Args()
	hashed := hash(args)
	maxDuration, err := time.ParseDuration(*maxDur)
	if err != nil {
		if *isVerbose {
			log.Printf("Cache file error: %s", err)
		}
		os.Exit(1)
	}
	filename := file(hashed, time.Now(), maxDuration)
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
		err := os.MkdirAll(*dir, 0700)
		if err != nil {
			if *isVerbose {
				log.Printf("Error (create dir): %s", err)
			}
			os.Exit(1)
		}
		// redirect IO
		file, err := os.Create(filename)
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
				log.Printf("Error (exit code %d): %s", ret, err)
			}
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

func file(hash string, time time.Time, d time.Duration) string {
	t := time.Truncate(d)
	return fmt.Sprintf("%s/%s-%d.stdout", *dir, hash, t.Unix())
}

func run(args []string, out io.Writer) (int, error) {
	p, err := exec.LookPath(args[0])
	if err != nil {
		log.Printf("Couldn't find exe %s - %s", p, err)
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
