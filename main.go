package main

/*
	steps:
		- have a -dir flag
		- check if -dir is a directory
		-	list files in -dir
		-	for each file start then sync with flag start
		-	if error call stop and exit
		-	setup infinite loop
		- if sign int is received, exit each command sync
		- exit if all subcommands have been stopped
*/

import (
	"syscall"
	"fmt"
	"flag"
	"io/ioutil"
	"os"
	"path"
	"os/exec"
	"github.com/tehmoon/errors"
)

var (
	FlagDir string
)

func parseFlags() {
	flag.StringVar(&FlagDir, "dir", "", "Execute all the files from the directory")

	flag.Parse()

	if FlagDir == "" {
		fmt.Fprintln(os.Stderr, "Flag -dir is empty")
		flag.Usage()
		os.Exit(2)
	}
}

func execCommand(name string, action string, pc *ProcessCollector) (int) {
	cmd := exec.Command(name, action)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			if ws, ok := e.Sys().(syscall.WaitStatus); ok {
				return ws.ExitStatus()
			}
		}

		if e, ok := err.(*os.SyscallError); ok {
			if e.Err == syscall.ECHILD {
				ws := pc.GetStatus(cmd.Process.Pid)
				if ws == nil {
					panic(errors.Wrapf(err, "Process has not been reaped by init!!", err))
				}

				return ws.ExitStatus()
			}
		}

		panic(errors.Wrapf(err, "Unhandled error of type %T", err))
	}

	return 0
}

func start(names []string) (error) {
	c1 := make(chan os.Signal, 1)
	c2 := make(chan os.Signal, 1)

	sd := NewSigDispatcher()
	sd.Register(c1, syscall.SIGCHLD)
	sd.Register(c2, syscall.SIGTERM, os.Kill, os.Interrupt)
	sd.Start()

	pc := &ProcessCollector{}
	pc.Start(c1)

	for _, name := range names {
		fmt.Printf("Executing %s\n", name)
		rc := execCommand(name, "start", pc)
		if rc != 0 {
			fmt.Fprintf(os.Stderr, "Program %s returned exit code %d\n", name, rc)

			return errors.New("")
		}

		defer func(name string) {
			fmt.Printf("Stopping %s\n", name)

			execCommand(name, "stop", pc)
		}(name)
	}

	loop(c2)

	return nil
}

func loop(c chan os.Signal) {
	<- c
}

func main() {
	parseFlags()

	names, err := listExecFiles(FlagDir)
	if err != nil {
		panic(err)
	}

	err = start(names)
	if err != nil {
		os.Exit(2)
	}
}

func buildAbsoluteCw(dir string) (string, error) {
	if dir[0] == '.' {
		pwd, err := os.Getwd()
		if err != nil {
			return "", errors.Wrap(err, "Error getting the current directory")
		}

		return path.Join(pwd, dir), nil
	}

	return dir, nil
}

func listExecFiles(dir string) ([]string, error) {
	absDir, err := buildAbsoluteCw(dir)
	if err != nil {
		return nil, err
	}

	ff, err := ioutil.ReadDir(absDir)
	if err != nil {
		return nil, err
	}

	files := make([]string, 0)

	for _, f := range ff {
		mode := f.Mode()

		if ! mode.IsRegular() {
			continue
		}

		if (mode | 0100 != mode) &&
			 (mode | 0010 != mode) &&
			 (mode | 0001 != mode) {
			continue
		}

		files = append(files, path.Join(absDir, f.Name()))
	}

	return files, nil
}
