package monitoring

import "time"

type BufferStorage struct {
	size int
	arr  []time.Duration
}

func NewBufferStorage(size int) *BufferStorage {
	return &BufferStorage{size: size}
}

func (bs *BufferStorage) Push(t time.Duration) {
	bs.arr = append(bs.arr, t)
	if len(bs.arr) >= bs.size {
		_, bs.arr = bs.arr[0], bs.arr[1:]
	}
}

func (bs *BufferStorage) Get() []time.Duration {
	return bs.arr
}
