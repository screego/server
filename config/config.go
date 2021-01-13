package config

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
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

	ExternalIP []string `split_words:"true"`

	TLSCertFile string `split_words:"true"`
	TLSKeyFile  string `split_words:"true"`

	ServerTLS     bool   `split_words:"true"`
	ServerAddress string `default:":5050" split_words:"true"`
	Secret        []byte `split_words:"true"`

	TurnAddress    string `default:":3478" required:"true" split_words:"true"`
	TurnStrictAuth bool   `default:"true" split_words:"true"`
	TurnPortRange  string `split_words:"true"`
	TurnExternal   bool   `default:"false" split_words:"true"`

	TrustProxyHeaders  bool     `split_words:"true"`
	AuthMode           string   `default:"turn" split_words:"true"`
	CorsAllowedOrigins []string `split_words:"true"`
	UsersFile          string   `split_words:"true"`
	Prometheus         bool     `split_words:"true"`

	CheckOrigin  func(string) bool `ignored:"true" json:"-"`
	ExternalIPV4 net.IP            `ignored:"true"`
	ExternalIPV6 net.IP            `ignored:"true"`
}

func (c Config) parsePortRange() (uint16, uint16, error) {
	if c.TurnPortRange == "" {
		return 0, 0, nil
	}

	parts := strings.Split(c.TurnPortRange, ":")
	if len(parts) != 2 {
		return 0, 0, errors.New("must include one colon")
	}
	stringMin := parts[0]
	stringMax := parts[1]
	min64, err := strconv.ParseUint(stringMin, 10, 16)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid min: %s", err)
	}
	max64, err := strconv.ParseUint(stringMax, 10, 16)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid max: %s", err)
	}

	return uint16(min64), uint16(max64), nil
}

func (c Config) PortRange() (uint16, uint16, bool) {
	min, max, _ := c.parsePortRange()
	return min, max, min != 0 && max != 0
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

	var errs []FutureLog
	config.ExternalIPV4, config.ExternalIPV6, errs = validateExternalIP(config.ExternalIP)
	logs = append(logs, errs...)

	min, max, err := config.parsePortRange()
	if err != nil {
		logs = append(logs, futureFatal(fmt.Sprintf("invalid SCREEGO_TURN_PORT_RANGE: %s", err)))
	} else if min == 0 && max == 0 {
		// valid; no port range
	} else if min == 0 || max == 0 {
		logs = append(logs, futureFatal("invalid SCREEGO_TURN_PORT_RANGE: min or max port is 0"))
	} else if min > max {
		logs = append(logs, futureFatal(fmt.Sprintf("invalid SCREEGO_TURN_PORT_RANGE: min port (%d) is higher than max port (%d)", min, max)))
	} else if (max - min) < 40 {
		logs = append(logs, FutureLog{
			Level: zerolog.WarnLevel,
			Msg:   "Less than 40 ports are available for turn. When using multiple TURN connections this may not be enough"})
	}

	return config, logs
}

func validateExternalIP(ips []string) (net.IP, net.IP, []FutureLog) {
	if len(ips) == 0 {
		return nil, nil, []FutureLog{futureFatal("SCREEGO_EXTERNAL_IP must be set")}
	}

	first := ips[0]

	firstParsed := net.ParseIP(first)
	if firstParsed == nil || first == "0.0.0.0" {
		return nil, nil, []FutureLog{futureFatal(fmt.Sprintf("invalid SCREEGO_EXTERNAL_IP: %s", first))}
	}
	firstIsIP4 := firstParsed.To4() != nil

	if len(ips) == 1 {
		if firstIsIP4 {
			return firstParsed, nil, nil
		}
		return nil, firstParsed, nil
	}

	second := ips[1]

	secondParsed := net.ParseIP(second)
	if secondParsed == nil || second == "0.0.0.0" {
		return nil, nil, []FutureLog{futureFatal(fmt.Sprintf("invalid SCREEGO_EXTERNAL_IP: %s", second))}
	}

	secondIsIP4 := secondParsed.To4() != nil

	if firstIsIP4 == secondIsIP4 {
		return nil, nil, []FutureLog{futureFatal("invalid SCREEGO_EXTERNAL_IP: the ips must be of different type ipv4/ipv6")}
	}

	if len(ips) > 2 {
		return nil, nil, []FutureLog{futureFatal("invalid SCREEGO_EXTERNAL_IP: too many ips supplied")}
	}

	if !firstIsIP4 {
		return secondParsed, firstParsed, nil
	}

	return firstParsed, secondParsed, nil
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
