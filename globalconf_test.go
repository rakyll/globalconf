package globalconf

import (
	"flag"
	"io/ioutil"
	"os"
	"testing"
)

func TestParse_Global(t *testing.T) {
	resetForTesting("")
	flagA := flag.Bool("a", false, "")
	flagB := flag.Float64("b", 0.0, "")
	flagC := flag.String("c", "", "")

	parse(t, "./testdata/global.ini")
	if !*flagA {
		t.Errorf("flagA found %v, expected true", *flagA)
	}
	if *flagB != 5.6 {
		t.Errorf("flagB found %v, expected 5.6", *flagB)
	}
	if *flagC != "Hello world" {
		t.Errorf("flagC found %v, expected 'Hello world'", *flagC)
	}
}

func TestParse_GlobalOverwrite(t *testing.T) {
	resetForTesting("-b=7.6")
	flagB := flag.Float64("b", 0.0, "")

	parse(t, "./testdata/global.ini")
	if *flagB != 7.6 {
		t.Errorf("flagB found %v, expected 7.6", *flagB)
	}
}

func TestParse_Custom(t *testing.T) {
	resetForTesting("")
	flagB := flag.Float64("b", 5.0, "")

	name := "custom"
	custom := flag.NewFlagSet(name, flag.ExitOnError)
	flagD := custom.String("d", "dd", "")

	Register(name, custom)
	parse(t, "./testdata/custom.ini")
	if *flagB != 5.0 {
		t.Errorf("flagB found %v, expected 5.0", *flagB)
	}
	if *flagD != "Hello d" {
		t.Errorf("flagD found %v, expected 'Hello d'", *flagD)
	}
}

func TestParse_CustomOverwrite(t *testing.T) {
	resetForTesting("-b=6")
	flagB := flag.Float64("b", 5.0, "")

	name := "custom"
	custom := flag.NewFlagSet(name, flag.ExitOnError)
	flagD := custom.String("d", "dd", "")

	Register(name, custom)
	parse(t, "./testdata/custom.ini")
	if *flagB != 6.0 {
		t.Errorf("flagB found %v, expected 6.0", *flagB)
	}
	if *flagD != "Hello d" {
		t.Errorf("flagD found %v, expected 'Hello d'", *flagD)
	}
}

func TestParse_GlobalAndCustom(t *testing.T) {
	resetForTesting("")
	flagA := flag.Bool("a", false, "")
	flagB := flag.Float64("b", 0.0, "")
	flagC := flag.String("c", "", "")

	name := "custom"
	custom := flag.NewFlagSet(name, flag.ExitOnError)
	flagD := custom.String("d", "", "")

	Register(name, custom)
	parse(t, "./testdata/globalandcustom.ini")
	if !*flagA {
		t.Errorf("flagA found %v, expected true", *flagA)
	}
	if *flagB != 5.6 {
		t.Errorf("flagB found %v, expected 5.6", *flagB)
	}
	if *flagC != "Hello world" {
		t.Errorf("flagC found %v, expected 'Hello world'", *flagC)
	}
	if *flagD != "Hello d" {
		t.Errorf("flagD found %v, expected 'Hello d'", *flagD)
	}
}

func TestParse_GlobalAndCustomOverwrite(t *testing.T) {
	resetForTesting("-a=true", "-b=5", "-c=Hello")
	flagA := flag.Bool("a", false, "")
	flagB := flag.Float64("b", 0.0, "")
	flagC := flag.String("c", "", "")

	name := "custom"
	custom := flag.NewFlagSet(name, flag.ExitOnError)
	flagD := custom.String("d", "", "")

	Register(name, custom)
	parse(t, "./testdata/globalandcustom.ini")
	if !*flagA {
		t.Errorf("flagA found %v, expected true", *flagA)
	}
	if *flagB != 5.0 {
		t.Errorf("flagB found %v, expected 5.0", *flagB)
	}
	if *flagC != "Hello" {
		t.Errorf("flagC found %v, expected 'Hello'", *flagC)
	}
	if *flagD != "Hello d" {
		t.Errorf("flagD found %v, expected 'Hello d'", *flagD)
	}
}

func TestSet(t *testing.T) {
	resetForTesting()
	file, _ := ioutil.TempFile("", "")
	conf := parse(t, file.Name())
	conf.Set("default", &flag.Flag{Name: "a", Value: newFlagValue("test")})

	flagA := flag.String("a", "", "")
	parse(t, file.Name())
	if *flagA != "test" {
		t.Errorf("flagA found %v, expected 'test'", *flagA)
	}
}

func TestDelete(t *testing.T) {
	resetForTesting()
	file, _ := ioutil.TempFile("", "")
	conf := parse(t, file.Name())
	conf.Set("default", &flag.Flag{Name: "a", Value: newFlagValue("test")})
	conf.Delete("default", "a")

	flagA := flag.String("a", "", "")
	parse(t, file.Name())
	if *flagA != "" {
		t.Errorf("flagNewA found %v, expected ''", *flagA)
	}
}

func parse(t *testing.T, filename string) *GlobalConf {
	conf, err := NewWithFilename(filename)
	if err != nil {
		t.Error(err)
	}
	conf.ParseAll()
	return conf
}

// Resets os.Args and the default flag set.
func resetForTesting(args ...string) {
	os.Args = append([]string{"cmd"}, args...)
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

type flagValue struct {
	str string
}

func (f *flagValue) String() string {
	return f.str
}

func (f *flagValue) Set(value string) error {
	f.str = value
	return nil
}

func newFlagValue(val string) *flagValue {
	return &flagValue{str: val}
}
