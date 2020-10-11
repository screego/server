package config

import (
	"crypto/rand"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"github.com/screego/server/config/mode"
)

var (
	prefix        = "screego"
	files         = []string{"screego.config.development.local", "screego.config.development", "screego.config.local", "screego.config"}
	absoluteFiles = []string{"/etc/screego/server.config"}
	osExecutable  = os.Executable
	osStat        = os.Stat
)

const (
	AuthModeTurn = "turn"
	AuthModeAll  = "all"
	AuthModeNone = "none"
)

// Config represents the application configuration.
type Config struct {
	LogLevel LogLevel `default:"info" split_words:"true"`

	ExternalIP string `split_words:"true"`

	TLSCertFile string `split_words:"true"`
	TLSKeyFile  string `split_words:"true"`

	ServerTLS     bool   `split_words:"true"`
	ServerAddress string `default:"0.0.0.0:5050" split_words:"true"`
	Secret        []byte `split_words:"true"`

	TurnAddress    string `default:"0.0.0.0:3478" required:"true" split_words:"true"`
	TurnStrictAuth bool   `default:"true" split_words:"true"`

	TrustProxyHeaders  bool     `split_words:"true"`
	AuthMode           string   `default:"turn" split_words:"true"`
	CorsAllowedOrigins []string `split_words:"true"`
	UsersFile          string   `split_words:"true"`
	Prometheus         bool     `split_words:"true"`

	CheckOrigin func(string) bool `ignored:"true" json:"-"`
}

// Get loads the application config.
func Get() (Config, []FutureLog) {
	var logs []FutureLog
	dir, log := getExecutableOrWorkDir()
	if log != nil {
		logs = append(logs, *log)
	}

	for _, file := range getFiles(dir) {
		_, fileErr := osStat(file)
		if fileErr == nil {
			if err := godotenv.Load(file); err != nil {
				logs = append(logs, futureFatal(fmt.Sprintf("cannot load file %s: %s", file, err)))
			} else {
				logs = append(logs, FutureLog{
					Level: zerolog.DebugLevel,
					Msg:   fmt.Sprintf("Loading file %s", file)})
			}
		} else if os.IsNotExist(fileErr) {
			continue
		} else {
			logs = append(logs, FutureLog{
				Level: zerolog.WarnLevel,
				Msg:   fmt.Sprintf("cannot read file %s because %s", file, fileErr)})
		}
	}

	config := Config{}
	err := envconfig.Process(prefix, &config)
	if err != nil {
		logs = append(logs,
			futureFatal(fmt.Sprintf("cannot parse env params: %s", err)))
	}

	if config.AuthMode != AuthModeTurn && config.AuthMode != AuthModeAll && config.AuthMode != AuthModeNone {
		logs = append(logs,
			futureFatal(fmt.Sprintf("invalid SCREEGO_AUTH_MODE: %s", config.AuthMode)))
	}

	if config.ExternalIP == "" {
		logs = append(logs, futureFatal("SCREEGO_EXTERNAL_IP must be set"))
	}

	if config.ServerTLS {
		if config.TLSCertFile == "" {
			logs = append(logs, futureFatal("SCREEGO_TLS_CERT_FILE must be set if TLS is enabled"))
		}

		if config.TLSKeyFile == "" {
			logs = append(logs, futureFatal("SCREEGO_TLS_KEY_FILE must be set if TLS is enabled"))
		}
	}

	var compiledAllowedOrigins []*regexp.Regexp
	for _, origin := range config.CorsAllowedOrigins {
		compiled, err := regexp.Compile(origin)
		if err != nil {
			logs = append(logs, futureFatal(fmt.Sprintf("invalid regex: %s", err)))
		}
		compiledAllowedOrigins = append(compiledAllowedOrigins, compiled)
	}

	config.CheckOrigin = func(origin string) bool {
		if origin == "" {
			return true
		}
		for _, compiledOrigin := range compiledAllowedOrigins {
			if compiledOrigin.Match([]byte(strings.ToLower(origin))) {
				return true
			}
		}
		return false
	}

	if len(config.Secret) == 0 {
		config.Secret = make([]byte, 32)
		if _, err := rand.Read(config.Secret); err == nil {
			logs = append(logs, FutureLog{
				Level: zerolog.InfoLevel,
				Msg:   "SCREEGO_SECRET unset, user logins will be invalidated on restart"})
		} else {
			logs = append(logs, futureFatal(fmt.Sprintf("cannot create secret %s", err)))
		}
	}

	if net.ParseIP(config.ExternalIP) == nil || config.ExternalIP == "0.0.0.0" {
		logs = append(logs, futureFatal(fmt.Sprintf("invalid SCREEGO_EXTERNAL_IP: %s", config.ExternalIP)))
	}

	return config, logs
}

func getExecutableOrWorkDir() (string, *FutureLog) {
	dir, err := getExecutableDir()
	// when using `go run main.go` the executable lives in th temp directory therefore the env.development
	// will not be read, this enforces that the current work directory is used in dev mode.
	if err != nil || mode.Get() == mode.Dev {
		return filepath.Dir("."), err
	}
	return dir, nil
}

func getExecutableDir() (string, *FutureLog) {
	ex, err := osExecutable()
	if err != nil {
		return "", &FutureLog{
			Level: zerolog.ErrorLevel,
			Msg:   "Could not get path of executable using working directory instead. " + err.Error()}
	}
	return filepath.Dir(ex), nil
}

func getFiles(relativeTo string) []string {
	var result []string
	for _, file := range files {
		result = append(result, filepath.Join(relativeTo, file))
	}
	homeDir, err := os.UserHomeDir()
	if err == nil {
		result = append(result, filepath.Join(homeDir, ".config/screego/server.config"))
	}
	result = append(result, absoluteFiles...)
	return result
}
