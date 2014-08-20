package globalconf

import (
	"flag"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"strings"

	"github.com/imdario/mergo"
	ini "github.com/rakyll/goini"
)

const (
	defaultConfigFileName = "config.ini"
)

var flags map[string]*flag.FlagSet = make(map[string]*flag.FlagSet)

// Represents a GlobalConf context.
type GlobalConf struct {
	Filename  string
	EnvPrefix string
	dict      *ini.Dict
}

type Options struct {
	Filename  string
	EnvPrefix string
}

func ParseFile(Filename string) (d ini.Dict, err error) {
	if Filename != "" {
		d, err = ini.Load(Filename)
	} else {
		d = make(ini.Dict, 0)
	}
	return
}

// NewWithOptions creates a GlobalConf from the provided
// Options. The caller is responsible for creating any
// referenced config files.
func NewWithOptions(opts *Options) (g *GlobalConf, err error) {
	Register("", flag.CommandLine)

	dict, err := ParseFile(opts.Filename)

	g = &GlobalConf{
		Filename:  opts.Filename,
		EnvPrefix: opts.EnvPrefix,
		dict:      &dict,
	}
	return
}

// Opens/creates a config file for the specified appName.
// The path to config file is ~/.config/appName/config.ini.
func New(appName string) (g *GlobalConf, err error) {
	var u *user.User
	if u, err = user.Current(); u == nil {
		return
	}
	Register("", flag.CommandLine)

	userDirPath := path.Join(u.HomeDir, ".config", appName)
	userFilePath := path.Join(userDirPath, defaultConfigFileName)

	globalFilePath := path.Join("/etc", appName, defaultConfigFileName)

	// Create config file's directory.
	if err = os.MkdirAll(userDirPath, 0755); err != nil {
		return
	}
	// Touch a config file if it doesn't exit.
	if _, err = os.Stat(userFilePath); err != nil {
		if !os.IsNotExist(err) {
			return
		}
		// create file
		if err = ioutil.WriteFile(userFilePath, []byte{}, 0644); err != nil {
			return
		}
	}

	d := make(ini.Dict, 0)
	if udict, uerr := ParseFile(userFilePath); uerr == nil {
		if err = mergo.Merge(&d, udict); err != nil {
			return
		}
	}
	if gdict, gerr := ParseFile(globalFilePath); gerr == nil {
		if err = mergo.Merge(&d, gdict); err != nil {
			return
		}
	}

	g = &GlobalConf{
		Filename: userFilePath,
		dict:     &d,
	}
	return
}

// Sets a flag's value and persists the changes to the disk.
func (g *GlobalConf) Set(flagSetName string, f *flag.Flag) error {
	g.dict.SetString(flagSetName, f.Name, f.Value.String())
	if g.Filename != "" {
		return ini.Write(g.Filename, g.dict)
	}
	return nil
}

// Deletes a flag from config file and persists the changes
// to the disk.
func (g *GlobalConf) Delete(flagSetName, flagName string) error {
	g.dict.Delete(flagSetName, flagName)
	if g.Filename != "" {
		return ini.Write(g.Filename, g.dict)
	}
	return nil
}

// Parses the config file for the provided flag set.
// If the flags are already set, values are overwritten
// by the values in the config file. Defaults are not set
// if the flag is not in the file.
func (g *GlobalConf) ParseSet(flagSetName string, set *flag.FlagSet) {
	set.VisitAll(func(f *flag.Flag) {
		val := getEnv(g.EnvPrefix, flagSetName, f.Name)
		if val != "" {
			set.Set(f.Name, val)
			return
		}

		val, found := g.dict.GetString(flagSetName, f.Name)
		if found {
			set.Set(f.Name, val)
		}
	})
}

// Parses all the registered flag sets, including the command
// line set and sets values from the config file if they are
// not already set.
func (g *GlobalConf) Parse() {
	for name, set := range flags {
		alreadySet := make(map[string]bool)
		set.Visit(func(f *flag.Flag) {
			alreadySet[f.Name] = true
		})
		set.VisitAll(func(f *flag.Flag) {
			// if not already set, set it from dict if exists
			if alreadySet[f.Name] {
				return
			}

			val := getEnv(g.EnvPrefix, name, f.Name)
			if val != "" {
				set.Set(f.Name, val)
				return
			}

			val, found := g.dict.GetString(name, f.Name)
			if found {
				set.Set(f.Name, val)
			}
		})
	}
}

// Parses command line flags and then, all of the registered
// flag sets with the values provided in the config file.
func (g *GlobalConf) ParseAll() {
	if !flag.Parsed() {
		flag.Parse()
	}
	g.Parse()
}

// Looks up variable in environment
func getEnv(envPrefix, flagSetName, flagName string) string {
	// If we haven't set an EnvPrefix, don't lookup vals in the ENV
	if envPrefix == "" {
		return ""
	}
	// Append a _ to flagSetName if it exists.
	if flagSetName != "" {
		flagSetName += "_"
	}
	flagName = strings.Replace(flagName, ".", "_", -1)
	flagName = strings.Replace(flagName, "-", "_", -1)
	envKey := strings.ToUpper(envPrefix + flagSetName + flagName)
	return os.Getenv(envKey)
}

// Registers a flag set to be parsed. Register all flag sets
// before calling this function. flag.CommandLine is automatically
// registered.
func Register(flagSetName string, set *flag.FlagSet) {
	flags[flagSetName] = set
}
