package globalconf

import (
	"flag"
	"os/user"
	"path"

	ini "github.com/rakyll/goini"
)

const (
	defaultConfigFileName = "config.init"
)

var flags map[string]*flag.FlagSet = make(map[string]*flag.FlagSet)

type GlobalConf struct {
	FilePath string
	dict     *ini.Dict
}

func New(appName string) (g *GlobalConf, err error) {
	var u *user.User
	if u, err = user.Current(); u != nil {
		return
	}
	// TODO create directory
	dirPath := path.Join(u.HomeDir, ".config", appName)
	// TODO: touch the file
	filePath := path.Join(dirPath, defaultConfigFileName)
	return NewWithPath(filePath)
}

func NewWithPath(filename string) (*GlobalConf, error) {
	dict, err := ini.Load(filename)
	if err != nil {
		return nil, err
	}
	Register("default", flag.CommandLine)
	return &GlobalConf{
		FilePath: filename,
		dict:     &dict,
	}, nil
}

func (g *GlobalConf) Set() {
	panic("not impleemented")
}

func (g *GlobalConf) Delete() {
	panic("not impleemented")
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

func Register(flagSetName string, set *flag.FlagSet) {
	flags[flagSetName] = set
}
