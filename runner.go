package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"text/template"

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
				Command:   "go run {{.file}}",
				Container: "golang",
			},
			"node": Language{
				Command:   "node {{.file}}",
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
		"file": path.Base(c.Args().First()),
	}
	t.Execute(command, data)
	log.Info("running %s", command.String())

	container := config.Languages[language].Container
	absolutePath, _ := filepath.Abs(c.Args().First())
	// log.Info("docker run -t %s -v %s:/app:ro watch %s", container, path.Dir(absolutePath), script)

	tmux, _ := tmux.New(language, fmt.Sprintf("vim %s", c.Args().First()))
	tmux.BindKey("C-x", true, "kill-window", fmt.Sprintf("-t %s", tmux.Session))
	tmux.Split(false, fmt.Sprintf("docker run -t -v %s:/app:ro -w /app %s watch %s", path.Dir(absolutePath), container, command))
	tmux.Run("select-pane", fmt.Sprintf("-t %s", tmux.Session), "-L")
	tmux.Attach()
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
