package main

import (
	"fmt"
	"github.com/hx/deploybot/app"
	"os"
)

func main() {
	err := app.NewApp(app.Config{
		BindAddress:  option("BIND_ADDRESS", "127.0.0.1:5555"),
		GitBranch:    option("GIT_BRANCH", "master"),
		RepoDir:      option("REPO_DIR", "*"),
		GitHubSecret: option("GITHUB_SECRET", ""),
	}).Run()
	if err != nil {
		fail(err.Error())
	}
}

func option(name, defaultVal string) string {
	fullName := "DEPLOYBOT_" + name
	if val, ok := os.LookupEnv(fullName); ok {
		return val
	}
	if defaultVal == "*" {
		fail("environment variable " + fullName + " must be set")
	}
	return defaultVal
}

func fail(reason string) {
	fmt.Fprintln(os.Stderr, reason)
	os.Exit(1)
}
