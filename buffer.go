package SmartBuffer

import (
	"errors"
	"io"
)

var (
	ErrBufNotEnough = errors.New("buffer not enough")
)

type Buffer struct {
	buf         []byte
	read_offset int // read at &buf[read_offset:], write at &buf[len(buf):]
	cap         int

	level level
	user  *user
	block *block

	next *Buffer // 实现链表结构
}

// 有些时候，需要将 Buffer.buf 暴露出去给它方进行 read([]byte) 这样的操作
// 使用 RestBytes() , 交出当前 buf 的空余部分

func (b *Buffer) Bytes() []byte {
	return b.buf
}

func (b *Buffer) RestBytes() ([]byte, int) {
	return b.buf[len(b.buf):], b.RestSize()
}

func (b *Buffer) RestSize() int {
	return b.cap - len(b.buf)
}

func (b *Buffer) UsedSize() int {
	return len(b.buf)
}

// 清空

func (b *Buffer) Reset() {
	b.buf = b.buf[:0]
	b.read_offset = 0
}

// 扩容

func (this *Buffer) LevelUP() bool {
	return this.user.levelup(this)
}

func (b *Buffer) Read(p []byte) (n int, err error) {
	if b.read_offset >= b.cap {
		b.Reset()
		return 0, io.EOF
	}
	n = copy(p, b.buf[b.read_offset:])
	b.read_offset += n
	return
}

func (b *Buffer) Write(p []byte) (n int, err error) {
	need_size := len(p) + len(b.buf)
	if need_size > b.cap {
		if !b.LevelUP() {
			return 0, ErrBufNotEnough
		}
	}
	b.buf = b.buf[:need_size]
	n = copy(b.buf[len(b.buf):need_size], p)
	return n, nil
}

func (b *Buffer) ReadAll(r io.Reader) ([]byte, error) {
	// ReadAll() 将 r 的数据全部写入 b
	if b.read_offset >= b.cap {
		b.Reset()
	}
	for {
		m, err := r.Read(b.buf[len(b.buf):b.cap])
		b.buf = b.buf[:len(b.buf)+m]
		if err == io.EOF {
			break
		}
		if err != nil {
			return b.Bytes(), err
		}
		if m == 0 && len(b.buf) == b.cap {
			if !b.LevelUP() {
				return b.Bytes(), ErrBufNotEnough
			}
		}
	}
	return b.Bytes(), nil // err is EOF, so return nil explicitly
}

//

func (b *Buffer) GoHome(){
	if b.user != nil{ // levelup 时取得的 swap_buffer 是没有 b.user 的
		b.user.stat_hook(b)
	}
	b.block.lock.Lock()
	if b.block.dead {
		b.block = nil
		b.block.lock.Unlock()
		return
	}
	b.Reset()
	b.next = b.block.port
	b.block.port = b
	b.block.rest++
	b.block.lock.Unlock()
}
