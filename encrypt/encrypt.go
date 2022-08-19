/*
	FIX the adding slice to Encrypt instead of one at a time? 

	make the iv gen encrypted ran
*/

package encrypt

import (

	// for encrypting things
	"golang.org/x/crypto/argon2"
	"crypto/aes"
	"crypto/cipher"

	// iv generation
	"math/rand"
	"time"
	"os"

	// for comparing the key as the correct one
	"encoding/base64"
)

// right now the correct password is foobar!
const knownPlaintext = "trans rights R human rights 1234"
//const encryptedPlaintext = "AAAAAAAAAAAAAAAAAAAAANGYxl8FWvWBoG+/KRgGRwSSwiXNG4VXA9jIQU5gmVIh"
const encryptedPlaintext = "AAAAAAAAAAAAAAAAAAAAAPJmzNkoSo2ojkWHqU9w5GJvQgz8Q6smbAeuB8qNdexf"

// makes a key, returns a chiper block
// then checks with correctKey function if the key is correct -- if it's correct then true is returned
func KeyGeneration(password string) (cipher.Block, bool, string){

	if len([]byte(password)) < 1{
		return nil, false, "password for key generation is too short, string empty"
	}

	// salt generation is going to be the same thing every time
	salt := []byte("qwertyuiopasdfghjklzxcvbnm")

	// current parameters: 4, 2048*1024, 4, 32
	key := argon2.IDKey([]byte(password), salt, 4, 2048*1024, 4, 32)

	ciphBlock, err := aes.NewCipher(key)

	if err != nil{
		return nil, false, err.Error()
	}
	return ciphBlock, CorrectKey(ciphBlock), ""
}

func CorrectKey(ciphBlock cipher.Block) bool{
	comparison := Encrypt([]byte(knownPlaintext), ciphBlock, true)

	encoder := base64.NewEncoding("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890+/")
	encryptedComp := encoder.EncodeToString(comparison)

	if encryptedComp == encryptedPlaintext{
		return true
	}

	return false 
}

func Encrypt(plaintext []byte, ciphBlock cipher.Block, keyTest bool) []byte{
	// adds padding in form of "\n"
	// can replace 0x0A with a slice (but you must write ... at the end)
	if len(plaintext)%aes.BlockSize != 0{
		for i := len(plaintext)%aes.BlockSize; i < aes.BlockSize; i++{
			plaintext = append(plaintext, 0x0A) // 0x0A = []byte("\n")
		}
	}

	encrypt := make([]byte, aes.BlockSize+len(plaintext))

	iv := encrypt[:aes.BlockSize]

	 // if just testing the key, then the iv will be blank (same as when the ciphered plaintext was first generated)
	if !keyTest{ // this is the random iv generation for not testing the key but encrypting the file
		// IV GENERATION SHOULD BE CHANGED TO CRYPTO/RAND
		rand.Seed(time.Now().UnixNano()^int64(os.Getpid()))
		rand.Read(iv)
	}

	encryptBlock := cipher.NewCBCEncrypter(ciphBlock, iv)

	encryptBlock.CryptBlocks(encrypt[aes.BlockSize:], plaintext)

	return encrypt
}

func Decrypt(encrypted []byte, ciphBlock cipher.Block) []byte{
	iv := encrypted[:aes.BlockSize]
	encrypted = encrypted[aes.BlockSize:]

	decryptBlock := cipher.NewCBCDecrypter(ciphBlock, iv)

	decrypt := make([]byte, len(encrypted))

	decryptBlock.CryptBlocks(decrypt, encrypted) // not sure if this works, if not then write it as "encrypted, encrypted" which will for sure work -- it's being written like this for consistency

	return decrypt
}
