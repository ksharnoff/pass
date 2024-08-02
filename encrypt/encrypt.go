package encrypt

import (
	// for encrypting things
	"crypto/aes"
	"crypto/cipher"
	"golang.org/x/crypto/argon2"

	// for iv generation
	"crypto/rand"
	"io"
)

const FileName = "pass.yaml"
// const FileName = "realPass.yaml"

// Makes a key, then a cipher block. It returns "" if the
// password generation was successfull.
// If the following function is changed, also change it in changeKey.go
func KeyGeneration(password string) (cipher.Block, string) {

	if len([]byte(password)) < 1 {
		return nil, "password for key generation is too short, string empty"
	}

	// Salt generation must be the same thing every time.
	salt := []byte("qwertyuiopasdfghjklzxcvbnm")

	// Current parameters: 4, 2048*1024, 4, 32 -- takes about 2 seconds
	key := argon2.IDKey([]byte(password), salt, 4, 2048*1024, 4, 32)

	ciphBlock, err := aes.NewCipher(key)

	if err != nil {
		return nil, err.Error()
	}
	return ciphBlock, ""
}

func Encrypt(plaintext []byte, ciphBlock cipher.Block) []byte {
	// adds padding in form of "\n"
	if len(plaintext)%aes.BlockSize != 0 {
		for i := len(plaintext) % aes.BlockSize; i < aes.BlockSize; i++ {
			plaintext = append(plaintext, 0x0A) // 0x0A = []byte("\n")
		}
	}
	encrypt := make([]byte, aes.BlockSize+len(plaintext))

	// iv generation
	iv := encrypt[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	encryptBlock := cipher.NewCBCEncrypter(ciphBlock, iv)

	encryptBlock.CryptBlocks(encrypt[aes.BlockSize:], plaintext)

	return encrypt
}

func Decrypt(encrypted []byte, ciphBlock cipher.Block) []byte {
	iv := encrypted[:aes.BlockSize]
	encrypted = encrypted[aes.BlockSize:]

	decryptBlock := cipher.NewCBCDecrypter(ciphBlock, iv)

	decrypt := make([]byte, len(encrypted))

	decryptBlock.CryptBlocks(decrypt, encrypted)

	return decrypt
}
