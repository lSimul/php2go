package std

import (
	"io/ioutil"
	"os"

	"github.com/lSimul/php2go/std/array"
)

// FileExists checks if the file with
// the given name exists.
// It does the same thing as PHP file_exists.
func FileExists(f string) bool {
	i, err := os.Stat(f)
	if os.IsNotExist(err) {
		return false
	}
	return !i.IsDir()
}

// Scandir returns every file and folder
// found in the directory.
// If d is not valid directory, empty array
// is returned.
// It does the same thing as PHP scandir,
// here it is just wrapper to "hide" extra
// information, only name is required.
// Right now it does not return "." and "..",
// but these directories does not make much
// sense anyways.
func Scandir(d string) array.String {
	entries, err := ioutil.ReadDir(d)
	res := array.NewString()
	if err != nil {
		return res
	}
	for _, e := range entries {
		res.Add(e.Name())
	}
	return res
}
