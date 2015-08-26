// Package tock provides a Ticker which is API-compatible with a
// time.Ticker, but also allows the caller to stop, restart, and adjust
// duration with which the Ticker ticks.
//
// Receivers can listen on the same channel for all of the above
// operations. Ticker is safe for use by multiple goroutines.
package tock

import (
	"errors"
	"time"
)

// ErrDuration is returned if a non-positive duration is given to a
// Ticker.
var ErrDuration = errors.New("non-positive interval provided")

// A Ticker holds a channel that delivers ticks of a clock at
// intervals.
//
// A Ticker's tick interval can be adjusted, and the Ticker can be
// restarted once stopped. Sends to C will cease when the Ticker is
// stopped and begin again when the Ticker is restarted.
type Ticker struct {
	C  <-chan time.Time // The channel on which the ticks are delivered.
	c  chan time.Time
	t  *time.Ticker
	dc chan time.Duration
}

// NewTicker returns a new Ticker containing a channel that will send
// the time with a period specified by the duration argument.
//
// It adjusts the intervals or drops ticks to make up for slow
// receivers. The duration d must be greater > 0; if not,
// NewTicker will panic.
//
// Adjust the ticker to change the tick duration; Restart the Ticker to
// start a previously stopped Ticker.
func NewTicker(d time.Duration) *Ticker {
	if d <= 0 {
		panic(errors.New("non-positive interval for NewTicker"))
	}

	t := &Ticker{
		c:  make(chan time.Time, 1),
		t:  time.NewTicker(d),
		dc: make(chan time.Duration, 1),
	}
	t.C = t.c
	go t.start(d)
	return t
}

func (t *Ticker) start(initial time.Duration) {
	var (
		tick *time.Ticker
		d    = initial
	)
	for {
		tick = t.t
		select {
		case nextd := <-t.dc:
			if t.t != nil {
				t.t.Stop()
			}
			// We wish to resume or change the Ticker.
			if nextd >= 0 {
				if nextd > 0 {
					// Adjusting the Ticker to a new duration.
					d = nextd
				}

				if d > 0 {
					// Create a new time.Ticker with the given duration.
					t.t = time.NewTicker(d)
				}
			}
		case tm := <-tick.C:
			// non-blocking send will drop tm on the floor if there are
			// slow receivers
			select {
			case t.c <- tm:
			default:
			}
		}
	}
}

// Stop turns off a Ticker.
//
// Stop does not close Ticker.C, to prevent a read from the channel
// succeeding incorrectly.
func (t *Ticker) Stop() {
	t.dc <- time.Duration(-1)
}

// Resume resumes a previously stopped Ticker.
//
// When a channel is resumed, ticks will continue to be sent down C.
func (t *Ticker) Resume() {
	t.dc <- time.Duration(0)
}

// Adjust changes the duration period that the ticker sends the
// time on.
//
// Adjust returns an error if d is non-positive. The existing duration
// is maintained in this case that an invalid duration is provided.
func (t *Ticker) Adjust(d time.Duration) error {
	if d <= 0 {
		return ErrDuration
	}
	t.dc <- d
	return nil
}
