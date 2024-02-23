package watcher

import (
	"os"
	"strings"
	"time"

	"github.com/JamesTiberiusKirk/hot-reloader-proxy/cmd/hrp/logger"
	"github.com/radovskyb/watcher"
)

type Config struct {
	TemplateDir   string
	Ignore        []string
	Include       []string
	IgnoreSuffix  []string
	IncludeSuffix []string
	RegexFilter   string
	EventWindow   int
	PolingTime    int
}

func (w *Config) overrideZeros() {
	if w.PolingTime == 0 {
		w.PolingTime = 10
	}

	if w.EventWindow == 0 {
		w.EventWindow = 140
	}

	// w.Ignore = append(w.Ignore, "~")
	// w.IgnoreSuffix = append(w.IgnoreSuffix, "_templ.go", "_templ.txt")
}

func StartTemplateWatcher(log logger.Logger, templatesChanged chan<- string, watcherConfig Config) {
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

	watcherConfig.overrideZeros()

	w.SetMaxEvents(1)

	if len(watcherConfig.Ignore) > 0 {
		w.AddFilterHook(nameContainsIgnoreFilterHook(true, watcherConfig.Ignore...))
	}

	if len(watcherConfig.Include) > 0 {
		w.AddFilterHook(nameContainsFilterHook(true, watcherConfig.Include...))
	}

	if len(watcherConfig.IncludeSuffix) > 0 {
		w.AddFilterHook(suffixFilterHook(false, watcherConfig.IncludeSuffix...))
	}

	if len(watcherConfig.IgnoreSuffix) > 0 {
		w.AddFilterHook(suffixIgnoreFilterHook(false, watcherConfig.IgnoreSuffix...))
	}

	previousT := time.Now()

	go func() {
		for {
			select {
			case event := <-w.Event:
				// NOTE: dont sendevents that come together within a certain time
				eventWindowHalf := time.Duration(watcherConfig.EventWindow / 2)
				modTimeBefore := event.ModTime().Add(-eventWindowHalf * time.Millisecond)
				modTimeAfter := event.ModTime().Add(+eventWindowHalf * time.Millisecond)

				if previousT.After(modTimeBefore) && previousT.Before(modTimeAfter) {
					continue
				}

				previousT = event.ModTime()

				templatesChanged <- event.Path
			case err := <-w.Error:
				log.Error("FS error: %s", err.Error()) // Print the event's info.
			case <-w.Closed:
				return
			}
		}
	}()

	// Watch test_folder recursively for changes.
	if err := w.AddRecursive(watcherConfig.TemplateDir); err != nil {
		log.Error("Add recusive error: %s", err.Error()) // Print the event's info.
	}

	// Start the watching process - it'll check for changes every 100ms.
	if err := w.Start(time.Millisecond * time.Duration(watcherConfig.PolingTime)); err != nil {
		log.Error("Error starting watcher: %s", err.Error())
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
		// Dont ignore
		return nil
	}
}

// suffixFilterHook - Extending watcher functionality with using strings.Contains
// This will add on match
func suffixFilterHook(useFullPath bool, suffixes ...string) watcher.FilterFileHookFunc {
	return func(info os.FileInfo, fullPath string) error {
		str := info.Name()

		if useFullPath {
			str = fullPath
		}

		for _, s := range suffixes {
			if strings.HasSuffix(str, s) {
				return nil
			}
		}

		// No match.
		return watcher.ErrSkip
	}
}

// suffixIgnoreFilterHook - Extending watcher functionality with using strings.Contains
// This will ignore on match
func suffixIgnoreFilterHook(useFullPath bool, suffixes ...string) watcher.FilterFileHookFunc {
	return func(info os.FileInfo, fullPath string) error {
		str := info.Name()

		if useFullPath {
			str = fullPath
		}

		for _, s := range suffixes {
			if strings.HasSuffix(str, s) {
				return watcher.ErrSkip
			}
		}

		// Dont ignore
		return nil
	}
}
