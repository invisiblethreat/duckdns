package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/TV4/env"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	yaml "gopkg.in/yaml.v2"
)

// Update contains everything that DuckDNS will need to update a record
type Update struct {
	Token string   `yaml:"token"`
	Names []string `yaml:"domains"`
}

// CLIOptions are to set things via CLI
type CLIOptions struct {
	Debug bool
	File  string
	Token string
	Log   string
	Names []string
}

// Valid checks that all parameters are set for an update
func (u *Update) Valid() bool {
	if len(u.Names) > 0 && u.Token != "" {
		return true
	}
	return false
}

// GetConfigCLI sets the arguments for an update if they have been passed in on
// the CLI
func getConfigCLI(c CLIOptions, log *logrus.Logger) Update {
	var u Update

	u.Token = c.Token
	log.Debugf("Set token from CLI to %s", c.Token)
	u.Names = c.Names
	log.Debugf("Set names from CLI to %s", strings.Join(c.Names, ", "))

	return u
}

// GetConfigFile reads the config for DuckDNS
func getConfigFile(existing *Update, file string, log *logrus.Logger) {

	var update Update

	yamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		log.WithError(err).Debug("error reading file")
		return
	}
	err = yaml.Unmarshal(yamlFile, &update)
	if err != nil {
		log.WithError(err).Debug("error unmarshaling YAML file")
		return
	}

	// Set the token if it's not empty and doesn't already exist
	if update.Token == "" {
		log.Debugf("the token is empty after trying to parse %s", file)
	} else if existing.Token == "" {
		existing.Token = update.Token
	}

	// Set names to if they exist and value is not already set
	if len(update.Names) == 0 {
		log.Debugf("no names/subdomains specified to update from %s", file)
	} else if len(existing.Names) == 0 {
		existing.Names = update.Names
	}

}

// GetConfigEnv is for reading items out of the environment if you didn't want
// to set them on the CLI
func getConfigEnv(u *Update, log *logrus.Logger) {
	token := env.String("DUCK_TOKEN", "")
	name := env.String("DUCK_NAMES", "")

	// Set the token if not already set
	if u.Token == "" {
		u.Token = token
		log.Debugf("Set token from environment to %s", token)
	}

	if len(u.Names) == 0 && name != "" {
		// support DUCK_NAME="domain1 domain2" from the environment
		u.Names = strings.Split(name, " ")

		log.Debugf("Set names from environment to %s",
			strings.Join(u.Names, ", "))
	}
}

func makeUpdate(update Update, log *logrus.Logger) error {
	log.Debugf("Dumping update params: %#v", update)
	if !update.Valid() {
		log.Fatal("Arguments not set for update!")
		os.Exit(1)
	}
	var errs []string
	stub := "https://www.duckdns.org/update?domains="
	tokenStub := "&token="
	ipStub := "&ip="

	for _, v := range update.Names {

		url := fmt.Sprintf("%s%s%s%s%s", stub, v, tokenStub, update.Token, ipStub)
		log.Debugf("Update string: %s", url)
		res, err := http.Get(url)
		if err != nil {
			errs = append(errs, err.Error())
			log.WithError(err).Error("Error contacting DuckDNS server")
			break
		}

		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			errs = append(errs, err.Error())
			log.WithError(err).Error("Error reading body response")
			break
		}
		res.Body.Close()

		if strings.Contains(string(bodyBytes), "KO") {
			errs = append(errs, fmt.Sprintf("Error updating %s with DuckDNS", v))
			break
		}

		log.Debugf("updated DuckDNS for name %s", v)

	}

	if len(errs) != 0 {
		return errors.New(strings.Join(errs, "\n"))
	}

	return nil
}

func main() {
	var cli CLIOptions

	pflag.BoolVarP(&cli.Debug, "debug", "d", false, "Use debug mode")
	pflag.StringVarP(&cli.File, "config", "c", "duckdns.yaml",
		"Config file location")
	pflag.StringSliceVarP(&cli.Names, "names", "n", nil,
		"Names to update with DuckDNS. Just the subdomain section. "+
			"Use the flag multiple times to set multiple values.")
	pflag.StringVarP(&cli.Token, "token", "t", "",
		"Token for updating DuckDNS")
	pflag.StringVarP(&cli.Log, "log", "l", "",
		"Log file location")

	pflag.Parse()

	log := logrus.New()
	path, err := os.OpenFile(cli.Log, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		log.Out = path
	} else {
		log.Info("No logging path set, using stderr")
	}

	if cli.Debug {
		log.SetLevel(logrus.DebugLevel)
	}

	log.Debugf("Logging level: %s", log.GetLevel().String())

	// CLI vars
	update := getConfigCLI(cli, log)

	// Set things that weren't set by the CLI
	if !update.Valid() {
		getConfigEnv(&update, log)
	}

	// File vars
	if !update.Valid() {
		getConfigFile(&update, cli.File, log)
	}

	if err := makeUpdate(update, log); err != nil {
		log.WithError(err).Fatal("error updating IP address")
		os.Exit(1)
	}
	log.Info("IP address updated successfully")
}
