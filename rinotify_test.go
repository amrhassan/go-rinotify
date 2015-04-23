package rinotify

import (
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/inotify"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"testing"
	"time"
)

func TestRinotify(t *testing.T) {

	// Create a temp dir that's in the user's homedir, because xattr doesn't work on tmpfs
	user, err := user.Current()
	assert.NoError(t, err)
	tmpDir, err := ioutil.TempDir(user.HomeDir, "workerd-collectortest")
	assert.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tmpDir)
		if err != nil {
			panic(err)
		}
	}()

	watcher := RecursivelyWatch(tmpDir, inotify.IN_CREATE)

	path1 := path.Join(tmpDir, "subdir1")
	path2 := path.Join(path1, "file1")

	err = os.MkdirAll(path1, os.ModePerm)
	assert.NoError(t, err)

	// Sadly newly-created directories are not watched right away so we need to give inotify some time to catch up
	time.Sleep(1 * time.Nanosecond)

	err = ioutil.WriteFile(path2, []byte("TOUCHED"), os.ModePerm)
	assert.NoError(t, err)

	ev1 := <-watcher
	ev2 := <-watcher

	assert.Equal(t, path1, ev1.Name)
	assert.Equal(t, path2, ev2.Name)
}
