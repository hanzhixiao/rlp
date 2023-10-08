package rlp

import (
	"io"
	"math/big"
	"reflect"
	"sync"
)

type encBuffer struct {
	str     []byte     // string data, contains everything except list headers
	lheads  []listhead // all list headers
	lhsize  int        // sum of sizes of all encoded list headers
	sizebuf [9]byte    // auxiliary buffer for uint encoding
}

func (buf *encBuffer) Write(b []byte) (int, error) {
	buf.str = append(buf.str, b...)
	return len(b), nil
}
func (b *encBuffer) copyTo(dst []byte) {
	strpos := 0
	pos := 0
	for _, head := range b.lheads {
		n := copy(dst[pos:], b.str[strpos:head.offset])
		strpos += n
		pos += n
		enc := head.encode(dst[pos:])
		pos += len(enc)
	}
	copy(dst[pos:], b.str[strpos:])
}

func (buf *encBuffer) writeTo(w io.Writer) (err error) {
	strpos := 0
	for _, head := range buf.lheads {
		// write string data before header
		if head.offset-strpos > 0 {
			n, err := w.Write(buf.str[strpos:head.offset])
			strpos += n
			if err != nil {
				return err
			}
		}
		// write the header
		enc := head.encode(buf.sizebuf[:])
		if _, err = w.Write(enc); err != nil {
			return err
		}
	}
	if strpos < len(buf.str) {
		// write string data after the last list header
		_, err = w.Write(buf.str[strpos:])
	}
	return err
}

func (b *encBuffer) writeUint64(u uint64) {
	if u == 0 {
		b.str = append(b.str, 0x80)
	} else if u < 128 {
		b.str = append(b.str, byte(u))
	} else {
		l := putint(b.sizebuf[1:], u)
		b.sizebuf[0] = byte(l) + 0x80
		b.str = append(b.str, b.sizebuf[:l+1]...)
	}
}
func (b *encBuffer) writeBool(bool2 bool) {
	if bool2 {
		b.str = append(b.str, 0x01)
	} else {
		b.str = append(b.str, 0x80)
	}
}
func (b *encBuffer) writeBytes(bet []byte) {
	if len(bet) == 1 && bet[0] <= 0x7f {
		b.str = append(b.str, bet[0])
	} else {
		b.encodeStringHeader(len(bet))
		b.str = append(b.str, bet...)
	}
}

func (b *encBuffer) encodeStringHeader(l int) {
	if l < 56 {
		b.str = append(b.str, 0x80+byte(l))
	} else {
		s := putint(b.sizebuf[1:], uint64(l))
		b.sizebuf[0] = 0x80 + byte(s)
		b.str = append(b.str, b.sizebuf[:s+1]...)
	}
}

func (b *encBuffer) list() int {
	b.lheads = append(b.lheads, listhead{
		offset: len(b.str),
		size:   b.lhsize,
	})
	return len(b.lheads) - 1
}

func (b *encBuffer) endlist(offset int) {
	lh := &b.lheads[offset]
	//lh.size包含头的长度的总长度
	lh.size = b.size() - lh.offset - lh.size
	if lh.size <= 56 {
		b.lhsize++
	} else {
		b.lhsize += 1 + intsize(uint64(lh.size))
	}
}

func (b *encBuffer) size() int {
	return len(b.str) + b.lhsize
}
func (w *encBuffer) writeBigInt(i *big.Int) {
	bitlen := i.BitLen()
	if bitlen <= 64 {
		w.writeUint64(i.Uint64())
		return
	}
	// Integer is larger than 64 bits, encode from i.Bits().
	// The minimal byte length is bitlen rounded up to the next
	// multiple of 8, divided by 8.
	length := ((bitlen + 7) & -8) >> 3
	w.encodeStringHeader(length)
	w.str = append(w.str, make([]byte, length)...)
	index := length
	buf := w.str[len(w.str)-length:]
	for _, d := range i.Bits() {
		for j := 0; j < wordBytes && index > 0; j++ {
			index--
			buf[index] = byte(d)
			d >>= 8
		}
	}
}

func (b *encBuffer) encode(val interface{}) error {
	rval := reflect.ValueOf(val)
	writer, err := cachedWriter(rval.Type())
	if err != nil {
		return err
	}
	return writer(rval, b)
}

func (buf *encBuffer) reset() {
	buf.str = buf.str[:0]
	buf.lhsize = 0
	buf.lheads = buf.lheads[:0]
}

// 不知道干什么的
type EncoderBuffer struct {
	buf *encBuffer
	dst io.Writer

	ownBuffer bool
}

func (w EncoderBuffer) Write(b []byte) (int, error) {
	return w.buf.Write(b)
}

func encBufferFromWriter(w io.Writer) *encBuffer {
	switch w := w.(type) {
	case EncoderBuffer:
		return w.buf
	case *EncoderBuffer:
		return w.buf
	case *encBuffer:
		return w
	default:
		return nil
	}
}

// 不知道啥意思
const wordBytes = (32 << (uint64(^big.Word(0)) >> 63)) / 8

var encBufferPool = sync.Pool{
	New: func() interface{} { return new(encBuffer) },
}

func getEncBuffer() *encBuffer {
	buf := encBufferPool.Get().(*encBuffer)
	buf.reset()
	return buf
}
