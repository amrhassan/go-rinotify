// Recursive inotify watching
package rinotify

import (
	"golang.org/x/exp/inotify"
	"io/ioutil"
	"os"
	"path"
)

// Flags that must be mixed in for the recursive watcher to work
const requiredFlags = inotify.IN_ISDIR | inotify.IN_CREATE | inotify.IN_DELETE_SELF

// Recursively watch the given path and its children (if it's a directory) for inotify events.
// The passed flags are as described in the inotify package.
//
// Be warned: creating nested directories in rapid succession (as you would do with `mkdir -p dir1/dir2/dir3`)
// may result in the grandchildren (dir2 and dir3 in this example) and their descendants to be undetected by
// this watcher. This is caused by inotifier having not given enough time to watch a directory to detect its child
// before actually creating the child.
func RecursivelyWatch(fsPath string, flags uint32, eventBufferSize uint) <-chan *inotify.Event {

	output := make(chan *inotify.Event, eventBufferSize)

	watcher, err := inotify.NewWatcher()
	if err != nil {
		panic(err)
	}

	err = watcher.AddWatch(fsPath, flags|requiredFlags)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			select {

			case ev := <-watcher.Event:

				if isDeleteSelf(ev) {

					close(output)
					watcher.Close()
					return

				} else if isNewChildDir(ev) {

					// Watch it asynchronously, propagate received events upwards until child dies
					go watchChild(output, RecursivelyWatch(ev.Name, flags, eventBufferSize))
				}

				// Also if appropriate, propagate up the stack
				if (ev.Mask & flags) > 0 { // At least one requested event type exists
					output <- ev
				}

			case err := <-watcher.Error:
				panic(err)
			}
		}
	}()

	// If the fsPath is a directory, watch over pre-existing children
	stat, err := os.Stat(fsPath)
	if err != nil {
		panic(err)
	}

	if stat.IsDir() {
		children, err := ioutil.ReadDir(fsPath)
		if err != nil {
			panic(err)
		}
		for _, child := range children {
			go watchChild(output, RecursivelyWatch(path.Join(fsPath, child.Name()), flags, eventBufferSize))
		}
	}

	return output
}

func isDeleteSelf(event *inotify.Event) bool {
	return (event.Mask & inotify.IN_DELETE_SELF) == inotify.IN_DELETE_SELF
}

func isNewChildDir(event *inotify.Event) bool {
	m := inotify.IN_CREATE | inotify.IN_ISDIR
	return (event.Mask & m) == m
}

// Pipes events from a child to its parent until the child dies
func watchChild(parent chan<- *inotify.Event, child <-chan *inotify.Event) {
	for childEvent := range child {
		parent <- childEvent
	}
}
