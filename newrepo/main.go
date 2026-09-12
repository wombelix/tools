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
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var logLevel slog.LevelVar
var logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
	Level: &logLevel,
}))

func main() {
	const REPOBASEURL = "git.sr.ht/~wombelix/"
	const REPOBASEURLGIT = "git@git.sr.ht:~wombelix/"
	const TPLREPONAME = "tpl"
	const NAME = "Dominik Wombacher"
	const EMAIL = "dominik@wombacher.cc"

	logLevel.Set(getLogLevelFromEnv())

	repoBaseUrlHttp := fmt.Sprintf("https://%s", REPOBASEURL)

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

	repoUrl := fmt.Sprintf("%s%s", REPOBASEURL, repoName)
	repoUrlGit := fmt.Sprintf("%s%s", REPOBASEURLGIT, repoName)
	repoTplUrlHttps := fmt.Sprintf("%s%s", repoBaseUrlHttp, TPLREPONAME)

	logger.Info("initializing newrepo", "name", repoName)

	gitBinary, err := exec.LookPath("git")
	if err != nil {
		panic(err)
	}
	logger.Debug("found git binary", "path", gitBinary)

	env := append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	runGit := func(dir string, args ...string) {
		cmd := exec.Command(gitBinary, args...)
		cmd.Env = env
		if dir != "" {
			cmd.Dir = dir
		}

		logger.Debug("executing git command", "args", args)

		out, err := cmd.CombinedOutput()
		if err != nil {
			panic(fmt.Sprintf(" git %v\n%s\n%s", args, err, string(out)))
		}
	}

	logger.Info("cloning repository", "url", repoUrlGit)
	runGit("", "clone", repoUrlGit)

	workDir := fmt.Sprintf("./%s", repoName)

	logger.Info("setting up branch and pulling template")
	runGit(workDir, "branch", "-m", "main")
	runGit(workDir, "remote", "add", "tpl", repoTplUrlHttps)
	runGit(workDir, "pull", "tpl", "main")
	runGit(workDir, "branch", "--unset-upstream")
	runGit(workDir, "remote", "remove", "tpl")

	logger.Info("updating template files with repo name")
	err = replaceStringInFile(fmt.Sprintf("%s/.build.yml", workDir), "tpl", repoName)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	err = replaceStringInFile(fmt.Sprintf("%s/README.md", workDir), "tpl", repoName)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	err = replaceStringInFile(fmt.Sprintf("%s/README.md", workDir), "Template repo with basic configs, LICENSE and README.", repoDesc)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	logger.Info("committing and pushing changes")
	runGit(workDir, "commit", "-am", "feat: update tpl files to new repo name")
	runGit(workDir, "push", "origin", "main", "-o", "skip-ci")

	logger.Info("registering with REUSE", "repo", repoUrl)
	err = reuseRegistration(NAME, EMAIL, repoUrl)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func replaceStringInFile(path, search, replace string) error {
	logger.Debug("replacing string in file", "path", path, "search", search, "replace", replace)

	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	input, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	searchBytes := []byte(search)
	replaceBytes := []byte(replace)

	if !bytes.Contains(input, searchBytes) {
		return nil
	}

	output := bytes.ReplaceAll(input, searchBytes, replaceBytes)

	err = os.WriteFile(path, output, info.Mode())
	if err != nil {
		return err
	}

	return nil
}

func reuseRegistration(name, email, repo string) error {
	logger.Debug("starting REUSE registration", "name", name, "email", email, "repo", repo)

	jar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}
	client := &http.Client{Jar: jar}

	resp, err := client.Get("https://api.reuse.software/register")
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Warn("failed to close GET response body", "error", err)
		}
	}()

	body, _ := io.ReadAll(resp.Body)

	re := regexp.MustCompile(`name="csrf_token"[^>]*value="([^"]+)"`)
	matches := re.FindSubmatch(body)
	if len(matches) < 2 {
		return fmt.Errorf("[REUSE] CSRF token not found")
	}
	token := string(matches[1])

	data := url.Values{}
	data.Set("csrf_token", token)
	data.Set("name", name)
	data.Set("confirm", email)
	data.Set("project", repo)

	resp, err = client.PostForm("https://api.reuse.software/register", data)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Warn("failed to close POST response body", "error", err)
		}
	}()

	postBody, _ := io.ReadAll(resp.Body)
	logger.Debug("REUSE registration response", "body", string(postBody))
	if !bytes.Contains(postBody, []byte("Registration successful")) {
		return fmt.Errorf("[REUSE] Registration failed, expected 'Registration successful' in response")
	}

	logger.Info("REUSE registration successful")
	return nil
}

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
