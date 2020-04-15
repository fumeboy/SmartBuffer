package SmartBuffer

import (
	"bytes"
	"fmt"
	"math/rand"
	"sync"
	"testing"
)

func makeBytes(n int) []byte {
	p := make([]byte, n)
	for i := 0; i < len(p); i++ {
		p[i] = byte(i % 256)
	}
	return p
}

func BenchmarkLoops(b *testing.B) {
	u := NewUser(LV_1024x16)
	p := makeBytes(LV_1024x16.size())
	for i := 0; i < b.N; i++ {
		size := LV_1024.size()
		buf := u.Get()
		_, err := buf.ReadAll(bytes.NewReader(p[:size]))
		if err != nil {
			b.Fatalf("readall failed:%v", err)
		}
		buf.GoHome()
	}
}

func BenchmarkParallel2(b *testing.B) {
	u := NewUser(LV_1024x16)
	p := makeBytes(LV_1024x16.size())
	wg := sync.WaitGroup{}
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			size := LV_1024.size()
			buf := u.Get()
			_, err := buf.ReadAll(bytes.NewReader(p[:size]))
			if err != nil {
				b.Fatalf("readall failed:%v", err)
			}
			buf.GoHome()
			wg.Done()
		}()
	}
	wg.Wait()
}

func BenchmarkParallelResize2(b *testing.B) {
	u := NewUser(LV_1024x16)
	p := makeBytes(LV_1024x16.size())
	wg := sync.WaitGroup{}
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			size := level(rand.Intn(int(LV_1024x16))).size()
			buf := u.Get()
			_, err := buf.ReadAll(bytes.NewReader(p[:size]))
			if err != nil {
				b.Fatalf("readall failed:%v", err)
			}
			buf.GoHome()
			wg.Done()
		}()
	}
	wg.Wait()
}

func BenchmarkParallel(b *testing.B) {
	u := NewUser(LV_1024x16)
	p := makeBytes(LV_1024x16.size())
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			size := LV_1024.size()
			buf := u.Get()
			_, err := buf.ReadAll(bytes.NewReader(p[:size]))
			if err != nil {
				b.Fatalf("readall failed:%v", err)
			}
			buf.GoHome()
		}
	})
	// for i := 0; i < len(pools.pools); i++ {
	// 	fmt.Println(pools.pools[i].blocks_cap_sum)
	// }
	// fmt.Println("---")
}

func BenchmarkLoopsResize(b *testing.B) {
	u := NewUser(LV_1024x16)
	p := makeBytes(LV_1024x16.size())

	for i := 0; i < b.N; i++ {
		size := level(rand.Intn(int(LV_1024x16))).size()
		buf := u.Get()
		_, err := buf.ReadAll(bytes.NewReader(p[:size]))
		if err != nil {
			b.Fatalf("readall failed:%v", err)
		}
		buf.GoHome()
	}
}

func BenchmarkParallelResize(b *testing.B) {
	u := NewUser(LV_1024x16)
	p := makeBytes(LV_1024x16.size())
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			size := level(rand.Intn(int(LV_1024x16))).size()
			buf := u.Get()
			_, err := buf.ReadAll(bytes.NewReader(p[:size]))
			if err != nil {
				b.Fatalf("readall failed:%v", err)
			}
			buf.GoHome()
		}
	})
}

func TestGetPut(t *testing.T) {
	u := NewUser(LV_1024x16)
	p := makeBytes(LV_1024.size())
	size := rand.Intn(LV_1024.size())
	buf := u.Get()
	_, err := buf.ReadAll(bytes.NewReader(p[:size]))
	if err != nil {
		fmt.Println("readall failed:", err)
	}
	fmt.Println("if all read", buf.UsedSize() == size)
	buf.GoHome()
}

func TestBufferWriteRead(t *testing.T) {
	u := NewUser(LV_1024x16)
	buf := u.Get()
	b := make([]byte, 3)
	n := 0

	buf.Write([]byte("abcdefg"))

	n, _ = buf.Read(b)
	fmt.Println(b, n)
	n, _ = buf.Read(b)
	fmt.Println(b, n)
	n, _ = buf.Read(b)
	fmt.Println(b, n)
	buf.GoHome()
}