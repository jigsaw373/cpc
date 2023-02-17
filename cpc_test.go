package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	mathrand "math/rand"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCopyBlockwise(t *testing.T) {
	td := t.TempDir()

	sizes := []int64{0, 1}
	for n := 1; n < runtime.NumCPU(); n++ {
		for delta := -1; delta <= 1; delta++ {
			sizes = append(sizes, int64(n*(4<<10)+delta))
		}
	}

	for _, size := range sizes {
		for run := 1; run <= 3; run++ {
			name := fmt.Sprintf("size%d-run%d", size, run)
			t.Run(name, func(t *testing.T) {
				testCopyBlockwise(t, td, size, run)
			})
		}
	}
}

func testCopyBlockwise(t *testing.T, td string, size int64, run int) {
	ss := fmt.Sprint(size)
	src := filepath.Join(td, "input-"+ss)
	dst := filepath.Join(td, "output-"+ss)

	want := randBytes(int(size))
	if err := ioutil.WriteFile(src, want, 0644); err != nil {
		t.Fatal(err)
	}

	if run == 3 {
		testDirtyCopy(t, size, want, dst)
	}

	st, err := cpBlockwise(loggerDiscard, src, dst)
	if err != nil {
		t.Fatalf("cpblockwise: %v", err)
	}

	testCopyResults(t, size, run, st)

	got, err := ioutil.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(got, want) {
		t.Fatalf("bytes didn't equal; dst len = %v; want len %v", len(got), size)
	}
}

func testDirtyCopy(t *testing.T, size int64, want []byte, dst string) {
	if size == 0 {
		t.Skip("n/a")
	}

	dirtyCopy := append([]byte(nil), want...)
	dirtyCopy[mathrand.Intn(int(size))] = byte(mathrand.Intn(256))
	if err := ioutil.WriteFile(dst, dirtyCopy, 0644); err != nil {
		t.Fatal(err)
	}
}

func testCopyResults(t *testing.T, size int64, run int, st *stats) {
	if run == 1 && st.PagesUnmodified != 0 {
		t.Errorf("initial unmodified pages = %v; want 0", st.PagesUnmodified)
	}
	if run == 2 && st.PagesWritten > 0 {
		t.Errorf("second run written pages = %v; want 0", st.PagesWritten)
	}
	if run == 3 {
		if st.PagesWritten != 1 {
			t.Errorf("PagesWritten = %v; want 1", st.PagesWritten)
		}
		if size > 4<<10 && st.PagesUnmodified == 0 {
			t.Errorf("PagesUnmodified = %v; want >0", st.PagesUnmodified)
		}
	}
}

// loggerDiscard is a Logf that throws away the logs given to it.
func loggerDiscard(string, ...interface{}) {}

func randBytes(n int) []byte {
	ret := make([]byte, n)
	rand.Read(ret)
	return ret
}
