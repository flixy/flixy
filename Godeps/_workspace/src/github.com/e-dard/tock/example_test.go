package tock

import (
	"fmt"
	"time"
)

func Example() {
	// Create a new Ticker.
	t := NewTicker(250 * time.Millisecond)

	// Listen on the ticker's channel
	go func() {
		for tick := range t.C {
			fmt.Println(tick.Format("15:04:05.000"))
		}
	}()

	// Wait some time and pause the ticker.
	time.Sleep(time.Second)
	t.Stop()

	// Wait some time and alter the tick duration (resuming the ticker).
	time.Sleep(time.Second)
	t.Adjust(50 * time.Millisecond)

	// Wait some time and return.
	time.Sleep(300 * time.Millisecond)
}
