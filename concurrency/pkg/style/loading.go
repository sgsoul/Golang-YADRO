package style

import (
	"fmt"
	"time"
)

func StartLoadingIndicator() chan bool {
	indicator := make(chan bool)
	go func() {
		defer close(indicator)
		animation := []string{"|", "/", "—", "\\"}
		progress := 0
		for {
			select {
			case <-indicator:
				return
			default:
				fmt.Printf("\rLoading comics %s  ", animation[progress])
				progress = (progress + 1) % len(animation)
				time.Sleep(101 * time.Millisecond)
			}
		}
	}()
	return indicator
}

func StopLoadingIndicator(indicator chan bool) {
	indicator <- true
}