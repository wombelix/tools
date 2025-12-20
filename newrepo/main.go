// SPDX-FileCopyrightText: 2025 Dominik Wombacher <dominik@wombacher.cc>
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strconv"
)

var logLevel slog.LevelVar
var logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
	Level: &logLevel,
}))

func main() {
	const REPOBASEURL = "https://git.sr.ht/~wombelix/"
	const TPLREPONAME = "tpl"
	const NAME = "Dominik Wombacher"
	const EMAIL = "dominik@wombacher.cc"

	logLevel.Set(getLogLevelFromEnv())

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

	logger.Debug("[gitBinary]")
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

		logger.Debug(fmt.Sprintf("[runGit] git %v", args))

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

	replaceStringInFile(workDir+"/.build.yml", "tpl", repoName)
	replaceStringInFile(workDir+"/README.md", "tpl", repoName)
	replaceStringInFile(workDir+"/README.md", "Template repo with basic configs, LICENSE and README.", repoDesc)

	runGit(workDir, "commit", "-am", "feat: update tpl files to new repo name")

	reuseRegistration(NAME, EMAIL, REPOBASEURL+repoName)
}

	logger.Debug(fmt.Sprintf("[replaceStringInFile] path: %s, search: %s, replace: %s", path, search, replace))

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

	logger.Debug(fmt.Sprintf("[reuseRegistration] name: %s, email, %s, repo: %s", name, email, repo))

	if err != nil {
		slog.Error(err.Error())
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("[REUSE] GET - Failed to close response body: " + err.Error())
		}
	}()

	body, _ := io.ReadAll(resp.Body)

	re := regexp.MustCompile(`name="csrf_token"[^>]*value="([^"]+)"`)
	matches := re.FindSubmatch(body)
	if len(matches) < 2 {
		slog.Error("[REUSE] CSRF token not found")
		return
	}
	token := string(matches[1])

	data := url.Values{}
	data.Set("csrf_token", token)
	data.Set("name", name)
	data.Set("confirm", email)
	data.Set("project", repo)

	resp, err = http.PostForm("https://api.reuse.software/register", data)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("[REUSE] POST - Failed to close response body: " + err.Error())
		}
	}()

	postBody, _ := io.ReadAll(resp.Body)
	logger.Debug(fmt.Sprintf("[reuseRegistration] POST Response Body - %s", string(postBody)))
func getLogLevelFromEnv() slog.Level {
	levelStr := os.Getenv("LOG_LEVEL")

	switch strings.ToLower(levelStr) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
