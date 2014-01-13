package globalconf

import (
	"flag"
	"io/ioutil"
	"os"
	"os/user"
	"path"

	ini "github.com/rakyll/goini"
)

const (
	defaultConfigFileName = "config.ini"
)

var flags map[string]*flag.FlagSet = make(map[string]*flag.FlagSet)

type GlobalConf struct {
	Filename string
	dict     *ini.Dict
}

func New(appName string) (g *GlobalConf, err error) {
	var u *user.User
	if u, err = user.Current(); u != nil {
		return
	}
	// Create config file's directory.
	dirPath := path.Join(u.HomeDir, ".config", appName)
	if err = os.MkdirAll(dirPath, 0644); err != nil {
		return
	}
	// Touch a config file if it doesn't exit.
	filePath := path.Join(dirPath, defaultConfigFileName)
	if _, err = os.Stat(filePath); err != nil {
		if !os.IsNotExist(err) {
			return
		}
		// create file
		if err = ioutil.WriteFile(filePath, []byte{}, 0644); err != nil {
			return
		}
	}
	return NewWithFilename(filePath)
}

func NewWithFilename(filename string) (*GlobalConf, error) {
	dict, err := ini.Load(filename)
	if err != nil {
		return nil, err
	}
	Register("", flag.CommandLine)
	return &GlobalConf{
		Filename: filename,
		dict:     &dict,
	}, nil
}

func (g *GlobalConf) Set(flagSetName string, f *flag.Flag) error {
	g.dict.SetString(flagSetName, f.Name, f.Value.String())
	return ini.Write(g.Filename, g.dict)
}

func (g *GlobalConf) Delete(flagSetName, flagName string) error {
	g.dict.Delete(flagSetName, flagName)
	return ini.Write(g.Filename, g.dict)
}

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
			val, found := g.dict.GetString(name, f.Name)
			if found {
				set.Set(f.Name, val)
			}
		})
	}
}

func (g *GlobalConf) ParseAll() {
	if !flag.Parsed() {
		flag.Parse()
	}
	g.Parse()
}

func Register(flagSetName string, set *flag.FlagSet) {
	flags[flagSetName] = set
}
