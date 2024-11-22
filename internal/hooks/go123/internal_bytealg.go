package go123

import (
	"bytes"   //gosim:notranslate
	"strings" //gosim:notranslate
	"unsafe"
)

func InternalBytealg_IndexByteString(s string, c byte) int {
	return strings.IndexByte(s, c)
}

func InternalBytealg_Equal(a, b []byte) bool {
	return bytes.Equal(a, b)
}

func InternalBytealg_IndexByte(b []byte, c byte) int {
	return bytes.IndexByte(b, c)
}

func InternalBytealg_MakeNoZero(n int) []byte {
	return make([]byte, n)
}

func InternalBytealg_Compare(a, b []byte) int {
	return bytes.Compare(a, b)
}

func InternalBytealg_abigen_runtime_cmpstring(a, b string) int {
	return strings.Compare(a, b)
}

func InternalBytealg_Count(b []byte, c byte) int {
	return bytes.Count(b, []byte{c})
}

func InternalBytealg_CountString(s string, c byte) int {
	return strings.Count(s, string([]byte{c}))
}

func InternalBytealg_abigen_runtime_memequal(a, b unsafe.Pointer, size uintptr) bool {
	panic("gosim not implemented")
}

func InternalBytealg_abigen_runtime_memequal_varlen(a, b unsafe.Pointer) bool {
	panic("gosim not implemented")
}

func InternalBytealg_Index(a, b []byte) int {
	return bytes.Index(a, b)
}

func InternalBytealg_IndexString(a, b string) int {
	return strings.Index(a, b)
}
