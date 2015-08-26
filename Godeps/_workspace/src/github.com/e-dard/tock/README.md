## tock

Because tick follows tock.

[![Build Status](https://drone.io/github.com/e-dard/tock/status.png)](https://drone.io/github.com/e-dard/tock/latest)
[![GoDoc](https://godoc.org/github.com/e-dard/tock?status.svg)](http://godoc.org/github.com/e-dard/tock)


tock provides a Ticker which is API-compatible with a time.Ticker, but also allows the caller to stop, restart, and adjust the duration with which the Ticker ticks.

Receivers on the tick channel can continue to listen on the same channel after any of the tock.Ticker operations are carried out.

Ticker is safe for use by multiple goroutines.

### Example

```go
package main

import (
	"fmt"
	"time"

	"github.com/e-dard/tock"
)

func main() {
	// Create a new Ticker.
	t := tock.NewTicker(250 * time.Millisecond)

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
```

Produces for example the following output:

```
10:32:56.567
10:32:56.817
10:32:57.067
10:32:57.317
10:32:58.368
10:32:58.418
10:32:58.468
10:32:58.518
10:32:58.568
```
