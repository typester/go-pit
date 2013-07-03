package pit

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func MockDir() (string, func()) {
	original := directory

	dir, err := ioutil.TempDir("", "")
	if err != nil {
		panic(err)
	}

	reset := func() {
		directory = original
		os.RemoveAll(dir)
	}
	directory = dir

	return dir, reset
}

func TestMock(t *testing.T) {
	func() {
		if directory != filepath.Join(os.Getenv("HOME"), ".pit") {
			t.Error("directory is not original")
			return
		}

		dir, reset := MockDir()
		defer reset()

		if directory != dir {
			t.Error("failed to mock")
			return
		}
	}()

	if directory != filepath.Join(os.Getenv("HOME"), ".pit") {
		t.Error("failed to reset")
		return
	}
}

func TestSimpleGet(t *testing.T) {
	_, reset := MockDir()
	defer reset()

	profile, err := Get("example.com", Requires{})
	if err != nil {
		t.Error("failed to Get: ", err)
		return
	}

	if profile == nil {
		t.Error("profile is nil")
		return
	}

	if len(*profile) != 0 {
		t.Error("profile is not empty")
		return
	}
}

func TestSetGet(t *testing.T) {
	_, reset := MockDir()
	defer reset()

	err := Set("example.com", Profile{
		"username": "example-user",
		"password": "example-password",
	})
	if err != nil {
		t.Error("failed to Set: ", err)
	}

	profile, err := Get("example.com", Requires{})
	if err != nil {
		t.Error("failed to Get: ", err)
		return
	}

	username, ok := (*profile)["username"]
	if !ok {
		t.Error("username does not exist")
		return
	}
	if username != "example-user" {
		t.Error("username is wrong")
		return
	}

	password, ok := (*profile)["password"]
	if !ok {
		t.Error("password does not exist")
		return
	}
	if password != "example-password" {
		t.Error("password is wrong")
		return
	}
}

func TestRequireCheck(t *testing.T) {
	_, reset := MockDir()
	defer reset()

	_, err := Get("example.com", Requires{"username"})
	if err == nil {
		t.Error("Get should be fail")
		return
	}

	if err.Error() != "missing configs: username" {
		t.Error("error is wrong")
		return
	}

	_, err = Get("example.com", Requires{"username", "password"})
	if err == nil {
		t.Error("Get should be fail")
		return
	}

	if err.Error() != "missing configs: username, password" {
		t.Error("error is wrong")
		return
	}

	err = Set("example.com", Profile{
		"username": "hoge",
	})
	if err != nil {
		t.Error("Set fail")
		return
	}

	_, err = Get("example.com", Requires{"username", "password"})
	if err == nil {
		t.Error("Get should be fail")
		return
	}

	if err.Error() != "missing configs: password" {
		t.Error("error is wrong")
		return
	}
}
