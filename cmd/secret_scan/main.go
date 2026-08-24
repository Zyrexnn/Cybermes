package main

import (
	"encoding/json"
	"fmt"
	"os"

	"cybermes/pkg/secrets"
)

func printFinding(f secrets.Finding) {
	data, err := json.Marshal(f)
	if err == nil {
		fmt.Println(string(data))
	}
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		fi, err := os.Stdin.Stat()
		if err != nil || (fi.Mode()&os.ModeCharDevice) != 0 {
			fmt.Fprintln(os.Stderr, "Usage: echo \"<content>\" | secret_scan OR secret_scan <file1> <dir/> ...")
			os.Exit(2)
		}
		findings := secrets.ScanReader(os.Stdin, "<stdin>")
		for _, f := range findings {
			printFinding(f)
		}
		os.Exit(0)
	}

	for _, arg := range args {
		info, err := os.Stat(arg)
		if err != nil {
			continue
		}
		if info.IsDir() {
			findings, err := secrets.ScanDirectory(arg, 8)
			if err == nil {
				for _, f := range findings {
					printFinding(f)
				}
			}
		} else {
			findings, err := secrets.ScanFile(arg)
			if err == nil {
				for _, f := range findings {
					printFinding(f)
				}
			}
		}
	}
	os.Exit(0)
}
