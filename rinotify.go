// Recursive inotify watching
package rinotify

import (
	"golang.org/x/exp/inotify"
)

func RecursivelyWatch(path string, flags uint32) <-chan *inotify.Event {

	output := make(chan *inotify.Event)

	watcher, err := inotify.NewWatcher()
	if err != nil {
		panic(err)
	}

	err = watcher.Watch(path)

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
					go watchChild(output, RecursivelyWatch(ev.Name, flags))
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
