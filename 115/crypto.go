// refer to https://github.com/deadblue/elevengo/tree/develop/internal/crypto/m115

package _115

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"io"
)

var (
	xorKeySeed = []byte{
		0xf0, 0xe5, 0x69, 0xae, 0xbf, 0xdc, 0xbf, 0x5a,
		0x1a, 0x45, 0xe8, 0xbe, 0x7d, 0xa6, 0x73, 0x88,
		0xde, 0x8f, 0xe7, 0xc4, 0x45, 0xda, 0x86, 0x94,
		0x9b, 0x69, 0x92, 0x0b, 0x6a, 0xb8, 0xf1, 0x7a,
		0x38, 0x06, 0x3c, 0x95, 0x26, 0x6d, 0x2c, 0x56,
		0x00, 0x70, 0x56, 0x9c, 0x36, 0x38, 0x62, 0x76,
		0x2f, 0x9b, 0x5f, 0x0f, 0xf2, 0xfe, 0xfd, 0x2d,
		0x70, 0x9c, 0x86, 0x44, 0x8f, 0x3d, 0x14, 0x27,
		0x71, 0x93, 0x8a, 0xe4, 0x0e, 0xc1, 0x48, 0xae,
		0xdc, 0x34, 0x7f, 0xcf, 0xfe, 0xb2, 0x7f, 0xf6,
		0x55, 0x9a, 0x46, 0xc8, 0xeb, 0x37, 0x77, 0xa4,
		0xe0, 0x6b, 0x72, 0x93, 0x7e, 0x51, 0xcb, 0xf1,
		0x37, 0xef, 0xad, 0x2a, 0xde, 0xee, 0xf9, 0xc9,
		0x39, 0x6b, 0x32, 0xa1, 0xba, 0x35, 0xb1, 0xb8,
		0xbe, 0xda, 0x78, 0x73, 0xf8, 0x20, 0xd5, 0x27,
		0x04, 0x5a, 0x6f, 0xfd, 0x5e, 0x72, 0x39, 0xcf,
		0x3b, 0x9c, 0x2b, 0x57, 0x5c, 0xf9, 0x7c, 0x4b,
		0x7b, 0xd2, 0x12, 0x66, 0xcc, 0x77, 0x09, 0xa6,
	}

	xorClientKey = []byte{
		0x42, 0xda, 0x13, 0xba, 0x78, 0x76, 0x8d, 0x37,
		0xe8, 0xee, 0x04, 0x91,
	}

	rsaPrivateKey = []byte("-----BEGIN RSA PRIVATE KEY-----\n" +
		"MIICXAIBAAKBgQCMgUJLwWb0kYdW6feyLvqgNHmwgeYYlocst8UckQ1+waTOKHFC\n" +
		"TVyRSb1eCKJZWaGa08mB5lEu/asruNo/HjFcKUvRF6n7nYzo5jO0li4IfGKdxso6\n" +
		"FJIUtAke8rA2PLOubH7nAjd/BV7TzZP2w0IlanZVS76n8gNDe75l8tonQQIDAQAB\n" +
		"AoGANwTasA2Awl5GT/t4WhbZX2iNClgjgRdYwWMI1aHbVfqADZZ6m0rt55qng63/\n" +
		"3NsjVByAuNQ2kB8XKxzMoZCyJNvnd78YuW3Zowqs6HgDUHk6T5CmRad0fvaVYi6t\n" +
		"viOkxtiPIuh4QrQ7NUhsLRtbH6d9s1KLCRDKhO23pGr9vtECQQDpjKYssF+kq9iy\n" +
		"A9WvXRjbY9+ca27YfarD9WVzWS2rFg8MsCbvCo9ebXcmju44QhCghQFIVXuebQ7Q\n" +
		"pydvqF0lAkEAmgLnib1XonYOxjVJM2jqy5zEGe6vzg8aSwKCYec14iiJKmEYcP4z\n" +
		"DSRms43hnQsp8M2ynjnsYCjyiegg+AZ87QJANuwwmAnSNDOFfjeQpPDLy6wtBeft\n" +
		"5VOIORUYiovKRZWmbGFwhn6BQL+VaafrNaezqUweBRi1PYiAF2l3yLZbUQJAf/nN\n" +
		"4Hz/pzYmzLlWnGugP5WCtnHKkJWoKZBqO2RfOBCq+hY4sxvn3BHVbXqGcXLnZPvo\n" +
		"YuaK7tTXxZSoYLEzeQJBAL8Mt3AkF1Gci5HOug6jT4s4Z+qDDrUXo9BlTwSWP90v\n" +
		"wlHF+mkTJpKd5Wacef0vV+xumqNorvLpIXWKwxNaoHM=\n" +
		"-----END RSA PRIVATE KEY-----")
	rsaPublicKey = []byte("-----BEGIN RSA PUBLIC KEY-----\n" +
		"MIGJAoGBANHetaZ5idEKXAsEHRGrR2Wbwys+ZakvkjbdLMIUCg2klfoOfvh19vrL\n" +
		"TZgfXl47peZ4Ed1zt6QQUlQiL6zCBqdOiREhVFGv/PXr/eiHvJrbZ1wCqDX3XL53\n" +
		"pgOvggaD9DnnztQokyPfnJBVdp4VeYuUU+iQWLPi4/GGsHsEapltAgMBAAE=\n" +
		"-----END RSA PUBLIC KEY-----")

	rsaClientKey *rsa.PrivateKey
	rsaServerKey *rsa.PublicKey
)

type Key [16]byte

func init() {
	block, _ := pem.Decode(rsaPrivateKey)
	rsaClientKey, _ = x509.ParsePKCS1PrivateKey(block.Bytes)
	block, _ = pem.Decode(rsaPublicKey)
	rsaServerKey, _ = x509.ParsePKCS1PublicKey(block.Bytes)
}

func GenerateKey() Key {
	key := Key{}
	_, _ = io.ReadFull(rand.Reader, key[:])
	return key
}

func Encode(input []byte, key Key) (output string) {
	buf := make([]byte, 16+len(input))
	copy(buf, key[:])
	copy(buf[16:], input)
	xorTransform(buf[16:], xorDeriveKey(key[:], 4))
	reverseBytes(buf[16:])
	xorTransform(buf[16:], xorClientKey)
	output = base64.StdEncoding.EncodeToString(rsaEncrypt(buf))
	return
}

func Decode(input string, key Key) (output []byte, err error) {
	data, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return
	}
	data = rsaDecrypt(data)
	output = make([]byte, len(data)-16)
	copy(output, data[16:])
	xorTransform(output, xorDeriveKey(data[:16], 12))
	reverseBytes(output)
	xorTransform(output, xorDeriveKey(key[:], 4))
	return
}

func xorDeriveKey(seed []byte, size int) []byte {
	key := make([]byte, size)
	for i := 0; i < size; i++ {
		key[i] = (seed[i] + xorKeySeed[size*i]) & 0xff
		key[i] ^= xorKeySeed[size*(size-i-1)]
	}
	return key
}

func xorTransform(data []byte, key []byte) {
	dataSize, keySize := len(data), len(key)
	mod := dataSize % 4
	if mod > 0 {
		for i := 0; i < mod; i++ {
			data[i] ^= key[i%keySize]
		}
	}
	for i := mod; i < dataSize; i++ {
		data[i] ^= key[(i-mod)%keySize]
	}
}

func rsaEncrypt(input []byte) []byte {
	plainSize, blockSize := len(input), rsaServerKey.Size()-11
	buf := bytes.Buffer{}
	for offset := 0; offset < plainSize; offset += blockSize {
		sliceSize := blockSize
		if offset+sliceSize > plainSize {
			sliceSize = plainSize - offset
		}
		slice, _ := rsa.EncryptPKCS1v15(
			rand.Reader, rsaServerKey, input[offset:offset+sliceSize])
		buf.Write(slice)
	}
	return buf.Bytes()
}

func rsaDecrypt(input []byte) []byte {
	output := make([]byte, 0)
	cipherSize, blockSize := len(input), rsaServerKey.Size()
	for offset := 0; offset < cipherSize; offset += blockSize {
		sliceSize := blockSize
		if offset+sliceSize > cipherSize {
			sliceSize = cipherSize - offset
		}
		slice, _ := rsa.DecryptPKCS1v15(
			rand.Reader, rsaClientKey, input[offset:offset+sliceSize])
		output = append(output, slice...)
	}
	return output
}

func reverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}
