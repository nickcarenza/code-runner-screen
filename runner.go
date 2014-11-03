package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path"
	"path/filepath"
	"syscall"

	"github.com/ChrisMckenzie/code-runner/pkg/tmux"
	"github.com/codegangsta/cli"
	"github.com/op/go-logging"
	"gopkg.in/yaml.v2"
)

type Language struct {
	Command   string `yaml:"command"`
	Container string `yaml:"container"`
}

type Config struct {
	DefaultLang string              `yaml:"default"`
	Languages   map[string]Language `yaml:"languages,omitempty"`
}

const (
	pathName   = ".coderunner"
	configName = "config.yml"
	format     = "%{color}[%{level}]%{color:reset} %{message}"
)

var (
	log           = logging.MustGetLogger("runner")
	usr, _        = user.Current()
	home          = usr.HomeDir
	Path          = fmt.Sprintf("%s/%s", home, pathName)
	defaultConfig = Config{
		Languages: map[string]Language{
			"golang": Language{
				Command:   "go run %s",
				Container: "golang",
			},
			"node": Language{
				Command:   "node %s",
				Container: "node",
			},
		},
	}
)

func main() {
	logBackend := logging.NewLogBackend(os.Stderr, "", 0)
	syslogBackend, err := logging.NewSyslogBackend("")
	if err != nil {
		log.Fatal(err)
	}
	logging.SetBackend(logBackend, syslogBackend)
	logging.SetFormatter(logging.MustStringFormatter(format))
	app := cli.NewApp()
	app.Name = "Runner"
	app.Usage = "Run some code..."
	app.Version = "0.0.1a"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "lang, l",
			Value: "golang",
			Usage: "language for runner",
		},
	}
	app.Action = func(c *cli.Context) {
		Setupdir()
		config := GetConfig()
		startApp(&config, c)
	}

	app.Run(os.Args)
}

func startApp(config *Config, c *cli.Context) {
	language := c.String("lang")
	if language == "" {
		language = config.DefaultLang
	}

	command := fmt.Sprintf(config.Languages[language].Command, path.Base(c.Args().First()))
	log.Debug("Running: %s", command)

	container := config.Languages[language].Container
	absolutePath, _ := filepath.Abs(c.Args().First())

	// Create a detached tmux session
	tmux, _ := tmux.New(language, fmt.Sprintf("vim %s", c.Args().First()))
	// Bind CTRL-X to exit.
	tmux.BindKey("C-x", true, "kill-window", fmt.Sprintf("-t %s", tmux.Session))
	// Split the window vertical
	tmux.Split(false, fmt.Sprintf("docker run --name %s -t -v %s:/app:ro -w /app %s watch %s", container, path.Dir(absolutePath), container, command))
	// Set Active pane to left pane
	tmux.Run("select-pane", fmt.Sprintf("-t %s", tmux.Session), "-L")
	// Attach the tmux session.
	err := tmux.Attach()
	// log.Debug("%#v", err)
	if err == nil {
		kill := exec.Command("docker", "kill", container)
		rm := exec.Command("docker", "rm", container)
		kill.Run()
		rm.Run()
	}
}

func Setupdir() {
	log.Info("Checking: %s", Path)
	if _, err := os.Stat(Path); os.IsNotExist(err) {
		log.Info("Creating: %s", Path)
		err := os.MkdirAll(Path, 0777)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func GetConfig() Config {
	var file = fmt.Sprintf("%s/%s", Path, configName)
	if _, err := os.Stat(file); os.IsNotExist(err) {
		// Set Default Config
		return defaultConfig
	} else {
		var config = Config{}
		file, _ := ioutil.ReadFile(file)
		yaml.Unmarshal(file, &config)
		// log.Info("%#v \n %#v", config, string(file))
		defaultConfig.DefaultLang = config.DefaultLang
		appendMap(defaultConfig.Languages, config.Languages)
		return defaultConfig
	}
}

func appendMap(dest map[string]Language, src map[string]Language) {
	for lang, val := range src {
		dest[lang] = val
	}
}

func sigHandler(cs chan bool) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP,
		syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGCHLD)

	signal := <-c
	// logEvent(lognotice, sys, "Signal received: "+signal.String())

	switch signal {
	case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGCHLD:
		cs <- true
	case syscall.SIGHUP:
		cs <- false
	}
}
