package tock

import (
	"math"
	"testing"
	"time"
)

func TestNewTicker(t *testing.T) {
	NewTicker(time.Second)
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("NewTicker did not panic")
			}
		}()
		// NewTicker panics with non-positive duration.
		NewTicker(0)
	}()
}

func TestTicker(t *testing.T) {
	ticker := NewTicker(time.Millisecond)
	c := ticker.C

	// It ticks every millisecond
	d := measureTicks(c, 10)
	durationsWithin(d, 10*time.Millisecond, time.Millisecond, t)
}

func TestTicker_Stop_Resume(t *testing.T) {
	ticker := NewTicker(time.Millisecond)
	c := ticker.C

	// Stop the ticker and it doesn't tick anymore
	ticker.Stop()

	// In 30ms resume the ticker
	time.AfterFunc(30*time.Millisecond, ticker.Resume)
	d := measureTicks(c, 10)
	// It waits 30ms before it starts ticking again.
	durationsWithin(d, 40*time.Millisecond, 2*time.Millisecond, t)
}

func TestTicker_Adjust(t *testing.T) {
	ticker := NewTicker(time.Millisecond)
	c := ticker.C

	// Slowing the ticker down increases the time it takes to tick.
	err := ticker.Adjust(5 * time.Millisecond)
	if err != nil {
		t.Errorf("expected %v, got %v", nil, err)
	}

	d := measureTicks(c, 10)
	durationsWithin(d, 50*time.Millisecond, 2*time.Millisecond, t)

	// It returns an error if a non-positive duration is provided.
	if err = ticker.Adjust(-1); err == nil {
		t.Errorf("expected %v, got %v", ErrDuration, err)
	}
}

// measureTicks is a helper function to measure the time taken to
// receive n ticks on c.
func measureTicks(c <-chan time.Time, n int) time.Duration {
	now := time.Now()
	for _ = range c {
		n--
		if n == 0 {
			break
		}
	}
	return time.Since(now)
}

// durationsWithin checks that two durations are within epsilon of each
// other.
func durationsWithin(a, b, epsilon time.Duration, t *testing.T) {
	if math.Abs(float64(a-b)) > float64(epsilon) {
		t.Errorf("first duration %v and second duration %v differ by more than %v", a, b, epsilon)
	}
}

func BenchmarkAllocTicker(b *testing.B) {
	var t *time.Ticker
	for i := 0; i < b.N; i++ {
		t = time.NewTicker(15 * time.Millisecond)
		<-t.C
		<-t.C
		<-t.C
		t.Stop()
	}
}

func BenchmarkAllocAfter(b *testing.B) {
	var c <-chan time.Time
	for i := 0; i < b.N; i++ {
		c = time.After(15 * time.Millisecond)
		<-c
		c = time.After(15 * time.Millisecond)
		<-c
		c = time.After(15 * time.Millisecond)
		<-c
	}
}
