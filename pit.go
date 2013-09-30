package pit

import (
	"fmt"
	"io/ioutil"
	"launchpad.net/goyaml"
	"os"
	"path/filepath"
	"strings"
)

var (
	directory   string = filepath.Join(os.Getenv("HOME"), ".pit")
	defaultConf []byte = []byte(`---
"profile": 'default'
`)
)

type Requires []string
type Profile map[string]string

type Config struct {
	Profile string
}

type pit struct {
	conf       *Config
	configFile string
}

func newPit() (*pit, error) {
	pit := new(pit)

	pit.configFile = filepath.Join(directory, "pit.yaml")

	bytes, err := ioutil.ReadFile(pit.configFile)
	if err != nil {
		if os.IsNotExist(err) {
			bytes = defaultConf
		} else {
			return nil, err
		}
	}

	conf := new(Config)
	err = goyaml.Unmarshal(bytes, conf)
	if err != nil {
		return nil, err
	}

	pit.conf = conf

	return pit, err
}

func Get(name string, requires Requires) (*Profile, error) {
	pit, err := newPit()
	if err != nil {
		return nil, err
	}

	profiles, err := pit.ReadProfiles()
	if err != nil {
		return nil, err
	}

	profile, ok := (*profiles)[name]
	if !ok {
		profile = Profile{}
	}

	if len(requires) > 0 {
		missing := []string{}

		for i := range requires {
			_, ok := profile[requires[i]]
			if !ok {
				missing = append(missing, requires[i])
			}
		}

		if len(missing) > 0 {
			err = fmt.Errorf("missing configs: %s", strings.Join(missing, ", "))
			return nil, err
		}
	}

	return &profile, nil
}

func Set(name string, data Profile) (err error) {
	pit, err := newPit()
	if err != nil {
		return
	}

	profiles, err := pit.ReadProfiles()
	if err != nil {
		return
	}

	(*profiles)[name] = data

	bytes, err := goyaml.Marshal(profiles)
	if err != nil {
		return
	}

	tmpfile, err := ioutil.TempFile("", "")
	if err != nil {
		return
	}

	err = ioutil.WriteFile(tmpfile.Name(), bytes, 0644)
	if err != nil {
		return
	}
	tmpfile.Close()

	err = os.Rename(tmpfile.Name(), pit.ProfileFile())
	if err != nil {
		return
	}

	return
}

func (pit *pit) ProfileFile() string {
	return filepath.Join(directory, (*pit.conf).Profile+".yaml")
}

func (pit *pit) ReadProfiles() (*map[string]Profile, error) {
	var profiles map[string]Profile

	profileFile := pit.ProfileFile()
	bytes, err := ioutil.ReadFile(profileFile)
	if err != nil {
		if os.IsNotExist(err) {
			profiles = map[string]Profile{}
		} else {
			return nil, err
		}
	} else {
		err = goyaml.Unmarshal(bytes, &profiles)
		if err != nil {
			return nil, err
		}
	}

	return &profiles, nil
}
