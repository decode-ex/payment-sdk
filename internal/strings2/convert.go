package strings2

import "unsafe"

func FromBytesNoAlloc(b []byte) string {
	return unsafe.String(
		unsafe.SliceData(b),
		len(b),
	)
}

func ToBytesNoAlloc(s string) []byte {
	return unsafe.Slice(
		unsafe.StringData(s),
		len(s),
	)
}
