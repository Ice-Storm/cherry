package eventhub

import (
	"reflect"
	"runtime"
	"testing"
	"time"
)

func TestSub(t *testing.T) {
	ps := New(1)
	ch1 := ps.Sub("t1")
	ch2 := ps.Sub("t1")
	ch3 := ps.Sub("t2")

	ps.Pub("hi", "t1")
	ps.Pub("hello", "t2")

	ps.Shutdown()

	checkContents(t, ch1, []string{"hi"})
	checkContents(t, ch2, []string{"hi"})
	checkContents(t, ch3, []string{"hello"})
}
func TestUnsub(t *testing.T) {
	ps := New(1)
	defer ps.Shutdown()

	ch := ps.Sub("t1")

	ps.Pub("hi", "t1")
	ps.Unsub(ch, "t1")
	checkContents(t, ch, []string{"hi"})
}

func TestClose(t *testing.T) {
	ps := New(1)
	ch1 := ps.Sub("t1")
	ch2 := ps.Sub("t1")
	ch3 := ps.Sub("t2")
	ch4 := ps.Sub("t3")

	ps.Pub("hi", "t1")
	ps.Pub("hello", "t2")
	ps.Close("t1", "t2")

	checkContents(t, ch1, []string{"hi"})
	checkContents(t, ch2, []string{"hi"})
	checkContents(t, ch3, []string{"hello"})

	ps.Pub("welcome", "t3")
	ps.Shutdown()

	checkContents(t, ch4, []string{"welcome"})
}

func TestShutdown(t *testing.T) {
	start := runtime.NumGoroutine()
	New(10).Shutdown()
	time.Sleep(1 * time.Millisecond)
	if current := runtime.NumGoroutine(); current != start {
		t.Fatalf("Goroutine leak! Expected: %d, but there were: %d.", start, current)
	}
}

func checkContents(t *testing.T, ch chan interface{}, vals []string) {
	contents := []string{}
	for v := range ch {
		contents = append(contents, v.(string))
	}

	if !reflect.DeepEqual(contents, vals) {
		t.Fatalf("Invalid channel contents. Expected: %v, but was: %v.", vals, contents)
	}
}
