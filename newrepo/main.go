// SPDX-FileCopyrightText: 2025 Dominik Wombacher <dominik@wombacher.cc>
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"fmt"
	"log/slog"
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
	repoDesc := ""
	if len(args) == 2 {
		repoDesc = args[1]
	}

	gitBinary, err := exec.LookPath("git")
	if err != nil {
		panic(err)
	}

	env := append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	runGit := func(dir string, args ...string) {
		cmd := exec.Command(gitBinary, args...)
		cmd.Env = env
		if dir != "" {
			cmd.Dir = dir
		}

		out, err := cmd.CombinedOutput()
		if err != nil {
			panic(fmt.Sprintf(" git %v\n%s\n%s", args, err, string(out)))
		}
	}

	runGit("", "clone", REPOBASEURL+repoName)

	workDir := "./" + repoName

	runGit(workDir, "branch", "-m", "main")
	runGit(workDir, "remote", "add", "tpl", REPOBASEURL+TPLREPONAME)
	runGit(workDir, "pull", "tpl", "main")
	runGit(workDir, "branch", "--unset-upstream")
	runGit(workDir, "remote", "remove", "tpl")

	ReplaceStringInFile(workDir+"/.build.yml", "tpl", repoName)
	ReplaceStringInFile(workDir+"/README.md", "tpl", repoName)
	ReplaceStringInFile(workDir+"/README.md", "Template repo with basic configs, LICENSE and README.", repoDesc)

	runGit(workDir, "commit", "-am", "feat: update tpl files to new repo name")
}

func ReplaceStringInFile(path, search, replace string) {
	info, err := os.Stat(path)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	input, err := os.ReadFile(path)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	searchBytes := []byte(search)
	replaceBytes := []byte(replace)

	if !bytes.Contains(input, searchBytes) {
		return
	}

	output := bytes.ReplaceAll(input, searchBytes, replaceBytes)

	err = os.WriteFile(path, output, info.Mode())
	if err != nil {
		slog.Error(err.Error())
		return
	}
}
