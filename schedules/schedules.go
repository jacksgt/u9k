package schedules

import (
	"time"
)

// runs function f at certain intervals in the background (async)
// retains a function that when called shuts down the schedule
func schedule(f func(), interval time.Duration) func() {
	// set up timer and channel
	ticker := time.NewTicker(interval)
	quit := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				f()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	return func() { close(quit) }
}

func Init() {
	_ = schedule(func() {
		expireFiles()
	}, 24*time.Hour) // daily
	//defer expireFilesQuitter()
}
