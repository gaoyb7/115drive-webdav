package _115

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"unsafe"
)

//#cgo CFLAGS: -I${SRCDIR}
//#cgo LDFLAGS: -L${SRCDIR} -lencode115
//
//#include "encode115.h"
import "C"

type Key [16]byte

func init() {
	C.m115_edinit()
	C.m115_xorinit()
}

func GenerateKey() string {
	var key [16]byte
	_, _ = io.ReadFull(rand.Reader, key[:])
	return hex.EncodeToString(key[:])
}

func Encode(input []byte, key []byte) ([]byte, error) {
	out := make([]byte, 2048)
	outlen := int32(-1)

	C.m115_encode((*C.uchar)(unsafe.Pointer(&input[0])),
		C.uint(len(input)),
		(*C.uchar)(unsafe.Pointer(&out[0])),
		(*C.uint)(unsafe.Pointer(&outlen)),
		(*C.uchar)(unsafe.Pointer(&key[0])),
		(*C.uchar)(unsafe.Pointer(nil)))

	return out[:outlen], nil
}

func Decode(input []byte, key []byte) ([]byte, error) {
	out := make([]byte, 2048)
	outlen := int32(-1)
	keyout := make([]byte, 128)

	C.m115_decode((*C.uchar)(unsafe.Pointer(&input[0])),
		C.uint(len(input)),
		(*C.uchar)(unsafe.Pointer(&out[0])),
		(*C.uint)(unsafe.Pointer(&outlen)),
		(*C.uchar)(unsafe.Pointer(&key[0])),
		(*C.uchar)(unsafe.Pointer(&keyout[0])))

	return out[:outlen], nil
}
