/*
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

const KnownPlaintext = "trans rights R human rights 1234"
//const encryptedPlaintext = "AAAAAAAAAAAAAAAAAAAAAPJmzNkoSo2ojkWHqU9w5GJvQgz8Q6smbAeuB8qNdexf"
const encryptedPlaintext = "AAAAAAAAAAAAAAAAAAAAAPpAr+smJE48hV/gLXQ3+Nu9CjZHnVMpc4d9RPyhN4AZ"


// Makes a key, then a cipher block. It also returns a boolea
// for if the key is the correct key, by checking with CorrectKey function. 
func KeyGeneration(password string) (cipher.Block, bool, string){

	if len([]byte(password)) < 1{
		return nil, false, "password for key generation is too short, string empty"
	}

	// Salt generation is going to be the same thing every time. 
	salt := []byte("qwertyuiopasdfghjklzxcvbnm")

	// Current parameters: 4, 2048*1024, 4, 32 -- takes about 2 seconds
	key := argon2.IDKey([]byte(password), salt, 4, 2048*1024, 4, 32)

	ciphBlock, err := aes.NewCipher(key)

	if err != nil{
		return nil, false, err.Error()
	}
	return ciphBlock, CorrectKey(ciphBlock), ""
}

// If the key is correct then it returns true. 
func CorrectKey(ciphBlock cipher.Block) bool{
	comparison := Encrypt([]byte(KnownPlaintext), ciphBlock, true)

	encoder := base64.NewEncoding("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890+/")
	encryptedComp := encoder.EncodeToString(comparison)

	if encryptedComp == encryptedPlaintext{
		return true
	}

	return false 
}

func Encrypt(plaintext []byte, ciphBlock cipher.Block, keyTest bool) []byte{
	// adds padding in form of "\n"
	if len(plaintext)%aes.BlockSize != 0{
		for i := len(plaintext)%aes.BlockSize; i < aes.BlockSize; i++{
			plaintext = append(plaintext, 0x0A) // 0x0A = []byte("\n")
		}
	}
	encrypt := make([]byte, aes.BlockSize+len(plaintext))

	iv := encrypt[:aes.BlockSize]

	// If just testing the key, then the iv will be blank, in order
	// to compare it to the known plaintext. 
	if !keyTest{ 
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

	decryptBlock.CryptBlocks(decrypt, encrypted)
	
	return decrypt
}
