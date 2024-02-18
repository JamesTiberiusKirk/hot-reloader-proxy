package watcher

import (
	"os"
	"strings"
	"time"

	"github.com/JamesTiberiusKirk/hot-reloader-proxy/cmd/hrp/logger"
	"github.com/radovskyb/watcher"
)

func StartTemplateWatcher(log logger.Logger, templatesChanged chan<- string, templateDir string, ignore []string, include []string) {
	w := watcher.New()

	// SetMaxEvents to 1 to allow at most 1 event's to be received
	// on the Event channel per watching cycle.
	//
	// If SetMaxEvents is not set, the default is to send all events.
	// w.SetMaxEvents(1)

	// Only notify rename and move events.
	// w.FilterOps(watcher.Rename, watcher.Move)

	// Only files that match the regular expression during file listings
	// will be watched.
	// r := regexp.MustCompile(".*.(css|go)")
	// w.AddFilterHook(watcher.RegexFilterHook(r, false))

	w.AddFilterHook(nameContainsIgnoreFilterHook(true, ignore...))
	w.AddFilterHook(nameContainsFilterHook(true, include...))

	previousT := time.Now()

	go func() {
		for {
			select {
			case event := <-w.Event:

				// NOTE: dont sendevents that come together within a certain time
				modTimeBefore := event.ModTime().Add(-50 * time.Millisecond)
				modTimeAfter := event.ModTime().Add(+50 * time.Millisecond)

				if previousT.After(modTimeBefore) && previousT.Before(modTimeAfter) {
					continue
				}

				previousT = event.ModTime()

				templatesChanged <- event.Path
			case err := <-w.Error:
				log.Info("FS error: ", err.Error()) // Print the event's info.
			case <-w.Closed:
				return
			}
		}
	}()

	// Watch test_folder recursively for changes.
	if err := w.AddRecursive(templateDir); err != nil {
		log.Info("Add recusive error: ", err) // Print the event's info.
	}

	// Print a list of all of the files and folders currently
	// being watched and their paths.
	// for path := range w.WatchedFiles() {
	// 	slog.Info("[HOT_RELOAD] watching", "path", path)
	// }

	// Trigger 2 events after watcher started.
	// go func() {
	// 	w.Wait()
	// 	w.TriggerEvent(watcher.Create, nil)
	// 	w.TriggerEvent(watcher.Remove, nil)
	// }()

	// Start the watching process - it'll check for changes every 100ms.
	if err := w.Start(time.Millisecond * 100); err != nil {
		log.Info("Error starting watcher: ", err)
		panic(err)
	}

}

// nameContainsFilterHook - Extending watcher functionality with using strings.Contains
// This will add on match
func nameContainsFilterHook(useFullPath bool, contains ...string) watcher.FilterFileHookFunc {
	return func(info os.FileInfo, fullPath string) error {
		str := info.Name()

		if useFullPath {
			str = fullPath
		}

		for _, c := range contains {
			// Match
			if strings.Contains(str, c) {
				return nil
			}

		}
		// No match.
		return watcher.ErrSkip
	}
}

// nameContainsIgnoreFilterHook - Extending watcher functionality with using strings.Contains
// This will ignore on match
func nameContainsIgnoreFilterHook(useFullPath bool, contains ...string) watcher.FilterFileHookFunc {
	return func(info os.FileInfo, fullPath string) error {
		str := info.Name()

		if useFullPath {
			str = fullPath
		}

		for _, c := range contains {
			// Match
			if strings.Contains(str, c) {
				return watcher.ErrSkip
			}

		}
		// No match.
		return nil
	}
}
