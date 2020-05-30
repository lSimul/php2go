package std

import (
	"io/ioutil"
	"os"

	"github.com/lSimul/php2go/std/array"
)

func FileExists(f string) bool {
	i, err := os.Stat(f)
	if os.IsNotExist(err) {
		return false
	}
	return !i.IsDir()
}

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
