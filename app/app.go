package app

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hx/logs"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type App struct {
	Config
	server http.Server
	logger *logs.Logger
}

func NewApp(config Config) *App {
	app := App{
		Config: config,
		server: http.Server{
			Addr: config.BindAddress,
		},
		logger: logs.NewStdoutLogger(logs.Verbose),
	}
	handler := mux.NewRouter()
	handler.
		HandleFunc("/deploy", app.handleDeployRequest).
		Methods("POST")
	handler.
		PathPrefix("/").
		HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			writer.Header().Set("Content-Type", "text/plain; charset=UTF-8")
			writer.WriteHeader(404)
			writer.Write([]byte("🤨\n"))
		})
	app.server.Handler = handler
	return &app
}

func (a *App) Run() error {
	a.logger.Info("Running DeployBot in %s on %s for pushes to %s branches.", a.RepoDir, a.BindAddress, a.GitBranch)
	if a.GitHubSecret != "" {
		a.logger.Info("Signing secret: %s", a.GitHubSecret)
	}
	return a.server.ListenAndServe()
}

func (a *App) handleDeployRequest(writer http.ResponseWriter, request *http.Request) {
	fail := func(message string, status int, err string, params ...interface{}) {
		a.error(writer, request, message, status, err, params...)
	}

	eventName := request.Header.Get("X-GitHub-Event")
	if eventName != "push" {
		fail("Unsupported event", 400, "did not expect the '%s' event", eventName)
		return
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		fail("Error reading request body", 500, err.Error())
		return
	}

	if a.GitHubSecret != "" {
		actual, expected := request.Header.Get("X-Hub-Signature-256"), a.signature256(body)
		if actual != expected {
			fail("Incorrect signature", 401, "signature %s did not match expected %s", actual, expected)
			return
		}
	}

	payload := new(PushPayload)
	if err = json.Unmarshal(body, payload); err != nil {
		fail("Malformed JSON body", 400, err.Error())
		return
	}

	dir := a.repoDir(payload.RepoName())
	if dir == "" {
		fail("Repo not found", 404, "no repo named '%s'", payload.RepoName())
		return
	}

	scriptPath := a.findBuildScript(dir)
	if scriptPath == "" {
		fail("Cannot be built", 404, "no build script found in %s", dir)
		return
	}

	if payload.Branch() == a.GitBranch {
		go a.runScript(scriptPath)
	} else {
		a.logger.Debug("Push to '%s' branch ignored", payload.Branch())
	}

	writer.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	writer.WriteHeader(202)
	writer.Write([]byte("Thank you.\n"))
}

func (a *App) runScript(path string) {
	a.logger.Info("Running %s", path)

	cmd := exec.Command(path)
	cmd.Dir = filepath.Dir(path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		a.logger.Error("Run %s failed: %s", path, err)
	}
}

func (a *App) findBuildScript(dir string) string {
	pattern := filepath.Join(dir, "build.*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		a.logger.Error("Cannot glob %s: %s", pattern, err)
		return ""
	}
	for _, path := range matches {
		if stat, err := os.Stat(path); err == nil && !stat.IsDir() && stat.Mode()&0111 != 0 {
			return path
		}
	}
	return ""
}

func (a *App) repoDir(name string) string {
	dir := filepath.Join(a.RepoDir, name)
	stat, err := os.Stat(dir)
	if err != nil {
		return ""
	}
	if !stat.IsDir() {
		return ""
	}
	dir = filepath.Clean(dir)
	if filepath.Dir(dir) != strings.TrimSuffix(a.RepoDir, string([]byte{os.PathSeparator})) {
		return ""
	}
	return dir
}

func (a *App) signature256(body []byte) string {
	mac := hmac.New(sha256.New, []byte(a.GitHubSecret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func (a *App) error(response http.ResponseWriter, request *http.Request, message string, status int, err string, params ...interface{}) {
	if response != nil {
		http.Error(response, message, status)
	}
	msg := fmt.Sprintf("[%d] %s %s - %s", status, request.Method, request.URL.RequestURI(), message)
	if err != "" {
		msg += fmt.Sprintf(": "+err, params...)
	}
	a.logger.Error(msg)
}
