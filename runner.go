package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"text/template"

	"github.com/codegangsta/cli"
	"github.com/op/go-logging"
	"gopkg.in/yaml.v2"
)

type Language struct {
	Command string `yaml:"command"`
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
	path          = fmt.Sprintf("%s/%s", home, pathName)
	defaultConfig = Config{
		Languages: map[string]Language{
			"golang": Language{
				Command: "go run {{file}}",
			},
			"node": Language{
				Command: "node {{file}}",
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
		// log.Info("%#v", config)

		startApp(&config, c)
	}

	app.Run(os.Args)
}

func startApp(config *Config, c *cli.Context) {
	var language = c.String("lang")
	command := new(bytes.Buffer)
	if language == "" {
		language = config.DefaultLang
	}
	t := template.Must(template.New("command").Parse(config.Languages[language].Command))
	data := map[string]interface{}{
		"file": c.Args().First(),
	}
	t.Execute(command, data)
	log.Info("running %s", command.String())
	tmux := exec.Command("tmux", "new", "-d", fmt.Sprintf("-s %s", language), fmt.Sprintf("vim %s", c.Args().First()))
	// cmd := exec.Command("tmux", "n")
	key := exec.Command("tmux", "bind-key", "-n", "'C-x'", "kill-window", fmt.Sprintf("-t %s", language))
  docker := exec.Command("tmux", "split", fmt.Sprintf("-t %s", language), "-h", fmt.Sprintf("docker run --name '%s' -t -v %s:/app:ro  watch $(eval echo \$$lang)")
	attach := exec.Command("tmux", "attach", fmt.Sprintf("-t %s", language))
	// log.Info("%#v", cmd)
	attach.Stdout = os.Stdout
	attach.Stdin = os.Stdin

	tmux.Run()
	key.Run()
	attach.Run()
}

func Setupdir() {
	log.Info("Checking: %s", path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Info("Creating: %s", path)
		err := os.MkdirAll(path, 0777)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func GetConfig() Config {
	var file = fmt.Sprintf("%s/%s", path, configName)
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
