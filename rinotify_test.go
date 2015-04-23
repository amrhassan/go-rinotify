package rinotify

import (
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/inotify"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"
)


func TestRinotify(t *testing.T) {

	touch := func(filePath string) {
		err := ioutil.WriteFile(filePath, []byte{}, os.ModePerm)
		assert.NoError(t, err)
	}

	mkDirAll := func(path string) {
		err := os.MkdirAll(path, os.ModePerm)
		assert.NoError(t, err)
	}

	tmpDir, err := ioutil.TempDir("", "rnotifytest")
	assert.NoError(t, err)

	subdir0 := path.Join(tmpDir, "subdir0")
	mkDirAll(subdir0)

	watcher := RecursivelyWatch(tmpDir, inotify.IN_CREATE)

	subdir1 := path.Join(tmpDir, "subdir1")
	file0 := path.Join(tmpDir, "file0")
	subdir1file1 := path.Join(subdir1, "file1")
	subdir0file2 := path.Join(subdir0, "file2")

	// Everything starting here should be caught by inotify

	mkDirAll(subdir1)
	// Sadly newly-created directories are not watched right away so we need to give inotify some time to catch up
	time.Sleep(1 * time.Millisecond)

	touch(subdir1file1)
	touch(subdir0file2)
	touch(file0)

	ev1Name := (<-watcher).Name
	ev2Name := (<-watcher).Name
	ev3Name := (<-watcher).Name
	ev4Name := (<-watcher).Name

	receivedNames := []string{ev1Name, ev2Name, ev3Name, ev4Name}

	assert.Contains(t, receivedNames, subdir1)
	assert.Contains(t, receivedNames, subdir1file1)
	assert.Contains(t, receivedNames, subdir0file2)
	assert.Contains(t, receivedNames, file0)
}
