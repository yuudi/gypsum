package cli

import (
	"fmt"
	"os"
	"os/exec"
)

func daemon() {
	for {
		cmd := exec.Command(os.Args[0], "run")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			if _, ok := err.(*exec.ExitError); ok {
				// never mind
			} else {
				fmt.Println(err)
				os.Exit(1)
			}
		}
		if cmd.ProcessState.ExitCode() != 5 {
			os.Exit(cmd.ProcessState.ExitCode())
		}
	}
}
