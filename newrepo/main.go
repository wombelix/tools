// SPDX-FileCopyrightText: 2025 Dominik Wombacher <dominik@wombacher.cc>
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	const REPOBASEURL = "https://git.sr.ht/~wombelix/"
	const TPLREPONAME = "tpl"

	args := os.Args[1:]
	if len(args) < 1 || len(args) > 2 {
		fmt.Println("Usage: newrepo <name> <description>")
		os.Exit(1)
	}

	repoName := args[0]

	gitBinary, err := exec.LookPath("git")
	if err != nil {
		panic(err)
	}

	env := append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	// clone
	cmd := exec.Command(gitBinary, "clone", REPOBASEURL+repoName)
	cmd.Env = env

	out, err := cmd.CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("%s\n%s", err, out))
	}

	// rename master -> main
	cmd = exec.Command(gitBinary, "branch", "-m", "main")
	cmd.Env = env
	cmd.Dir = "./" + repoName

	out, err = cmd.CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("%s\n%s", err, out))
	}

	// add git remote tpl
	cmd = exec.Command(gitBinary, "remote", "add", "tpl", REPOBASEURL+TPLREPONAME)
	cmd.Env = env
	cmd.Dir = "./" + repoName

	out, err = cmd.CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("%s\n%s", err, out))
	}

	// pull from remote tpl
	cmd = exec.Command(gitBinary, "pull", "tpl", "main")
	cmd.Env = env
	cmd.Dir = "./" + repoName

	out, err = cmd.CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("%s\n%s", err, out))
	}

	// untrack origin master
	cmd = exec.Command(gitBinary, "branch", "--unset-upstream")
	cmd.Env = env
	cmd.Dir = "./" + repoName

	out, err = cmd.CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("%s\n%s", err, out))
	}

	// remove git remote tpl
	cmd = exec.Command(gitBinary, "remote", "remove", "tpl")
	cmd.Env = env
	cmd.Dir = "./" + repoName

	out, err = cmd.CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("%s\n%s", err, out))
	}
}
