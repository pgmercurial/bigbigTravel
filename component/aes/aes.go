/*
	AES加密解密算法，附带测试方法testAes
*/

package aes

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
)

func testAes() {
	// AES-128。key的长度：16, 24, 32 bytes 对应 AES-128, AES-192, AES-256
	key := "sfe023f_gaoziwen"
	token, err := AesEncrypt("1031294046@qq.com", key)
	if err != nil {
		panic(err)
	}

	fmt.Println(token)
	origData, err := AesDecrypt(token, key)
	//origData, err := AesDecrypt(bts, key)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(origData))
	os.Exit(1)
}

// encode
func AesEncrypt(origStr, keyStr string) (string, error) {
	origData := []byte(origStr)
	key := []byte(keyStr)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	blockSize := block.BlockSize()
	origData = PKCS5Padding(origData, blockSize)
	// origData = ZeroPadding(origData, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	// 根据CryptBlocks方法的说明，如下方式初始化crypted也可以
	// crypted := origData
	blockMode.CryptBlocks(crypted, origData)
	return base64.StdEncoding.EncodeToString(crypted), nil
}

// decode
func AesDecrypt(token string, keyStr string) (string, error) {
	if token == "" || keyStr == "" {
		return "", errors.New(fmt.Sprintf("param empty token:%s key %s", token, keyStr))
	}
	key := []byte(keyStr)
	crypted, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(crypted))
	// origData := crypted
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS5UnPadding(origData)
	// origData = ZeroUnPadding(origData)
	return string(origData), nil
}

func ZeroPadding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{0}, padding)
	return append(ciphertext, padtext...)
}

func ZeroUnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	// 去掉最后一个字节 unpadding 次
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
