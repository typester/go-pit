package pit

import (
	"errors"
	"io"
	"io/ioutil"
	"launchpad.net/goyaml"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

var (
	directory   string = filepath.Join(os.Getenv("HOME"), ".pit")
	defaultConf []byte = []byte(`---
"profile": 'default'
`)
)

type Profile map[string]string
type Requires Profile

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

func edit(file string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		if runtime.GOOS == "windows" {
			editor = "notepad"
		} else {
			editor = "vim"
		}
	}

	var stdin *os.File
	var shell, shellcflag string
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		stdin, _ = os.Open("CONIN$")
		shell = os.Getenv("COMSPEC")
		if shell == "" {
			shell = "cmd"
		}
		shellcflag = "/c"
		cmd = exec.Command(shell, shellcflag, editor, file)
	} else {
		stdin = os.Stdin
		shell = os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/sh"
		}
		shellcflag = "-c"
		cmd = exec.Command(shell, shellcflag, editor+" "+file)
	}
	cmd.Stdin = stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func copy(src, dst string) (err error) {
	s, err := os.Open(src)
	if err != nil {
		return
	}
	defer s.Close()

	d, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return
	}
	defer d.Close()

	_, err = io.Copy(d, s)
	return
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

	if requires != nil && len(requires) > 0 {
		missing := false
		for k, v := range requires {
			if _, ok := profile[k]; !ok {
				profile[k] = v
				missing = true
			}
		}

		if missing {
			bytes, err := goyaml.Marshal(profile)
			if err != nil {
				return nil, err
			}

			tmpfile, err := ioutil.TempFile("", "")
			if err != nil {
				return nil, err
			}
			tmpfile.Close()
			defer os.Remove(tmpfile.Name())

			filename := tmpfile.Name()

			err = ioutil.WriteFile(filename, bytes, 0600)
			if err != nil {
				return nil, err
			}

			fi, err := os.Stat(filename)
			if err != nil {
				return nil, err
			}

			modtime := fi.ModTime()

			err = edit(filename)
			if err != nil {
				return nil, err
			}

			fi, err = os.Stat(filename)
			if err != nil {
				return nil, err
			}

			if modtime == fi.ModTime() {
				return nil, errors.New("No changes.")
			}

			bytes, err = ioutil.ReadFile(filename)
			if err != nil {
				return nil, err
			}
			err = goyaml.Unmarshal(bytes, &profiles)
			if err != nil {
				return nil, err
			}
			err = Set(name, profile)
			if err != nil {
				return nil, err
			}
		}
	}

	return &profile, nil
}

func Set(name string, data Profile) error {
	pit, err := newPit()
	if err != nil {
		return err
	}

	profiles, err := pit.ReadProfiles()
	if err != nil {
		return err
	}

	(*profiles)[name] = data

	bytes, err := goyaml.Marshal(profiles)
	if err != nil {
		return err
	}

	tmpfile, err := ioutil.TempFile("", "")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(tmpfile.Name(), bytes, 0644)
	if err != nil {
		return err
	}
	tmpfile.Close()

	filename := tmpfile.Name()

	_, err = os.Stat(directory)
	if err != nil {
		err = os.Mkdir(directory, 0700)
		if err != nil {
			return err
		}
	}

	_, err = os.Stat(directory)
	if err != nil {
		err = os.Mkdir(directory, 0700)
		if err != nil {
			return err
		}
	}

	profilefile := pit.ProfileFile()

	if _, err = os.Stat(profilefile); err == nil {
		err = os.Remove(profilefile)
		if err != nil {
			return err
		}
	}

	err = os.Rename(filename, profilefile)
	if err != nil {
		err = copy(filename, profilefile)
		if err != nil {
			return err
		}
	}

	return nil
}

func (pit *pit) ProfileFile() string {
	return filepath.Join(directory, (*pit.conf).Profile+".yaml")
}

func (pit *pit) ReadProfiles() (*map[string]Profile, error) {
	profiles := map[string]Profile{}

	profileFile := pit.ProfileFile()
	bytes, err := ioutil.ReadFile(profileFile)
	if err != nil {
		if !os.IsNotExist(err) {
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
