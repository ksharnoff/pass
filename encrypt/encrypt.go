/*
	MIT License
	Copyright (c) 2022 Kezia Sharnoff

	encrypt.go
	Encrypts the slice of bytes, create keys, and write to files.
	The structs for entries are defined here so that they can be imported
	across different files and used in writing to files.
*/

package encrypt

import (
	// for encrypting things
	"crypto/aes"
	"crypto/cipher"
	"golang.org/x/crypto/argon2"

	// for iv generation
	"crypto/rand"
	"io"

	// for file reading and writing
	"gopkg.in/yaml.v3"
	"os"

	// for the entry and field structs
	"time"
)

// An entry represents an account or site
type Entry struct {
	Name      string
	Tags      string
	Usernames []Field
	Passwords []Field
	SecurityQ []Field
	// notes is 6 because that looks the best in /new as individual inputs,
	// textArea has unpredictable copy and pasting
	Notes     [6]string
	Circulate bool
	Urls      []string
	Created   time.Time
	Modified  time.Time
	Opened    time.Time
}
type Field struct {
	DisplayName string
	Value       string
}

const FileName = "pass.yaml" 

// Input: the user's password.
// Output: cipher block made from a key from the password.
// If the following function is changed, also change it in changeKey.go
func KeyGeneration(password string) (cipher.Block, string) {

	if len([]byte(password)) < 1 {
		return nil, "Password given for key generation is zero characters"
	}

	// Salt generation must be the same thing every time.
	salt := []byte("qwertyuiopasdfghjklzxcvbnm")

	// Current parameters: 1, 64*1024, 4, 32
	key := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	
	ciphBlock, err := aes.NewCipher(key)

	if err != nil {
		return nil, err.Error()
	}
	return ciphBlock, ""
}

// Input: the plaintext slice and a cipher block.
// Return: the encrypted bytes.
// Encrypts a slice of bytes using the cipher block (from the key)
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

// Input: the encrypted slice and a cipher block made from the same original
// password.
// Return: a decrypted slice.
// Decrypts the encrypted bytes using the cipher block
// If the decryption failed (wrong password), nonsense will be returned.
func Decrypt(encrypted []byte, ciphBlock cipher.Block) []byte {
	iv := encrypted[:aes.BlockSize]
	encrypted = encrypted[aes.BlockSize:]

	decryptBlock := cipher.NewCBCDecrypter(ciphBlock, iv)

	decrypt := make([]byte, len(encrypted))

	decryptBlock.CryptBlocks(decrypt, encrypted)

	return decrypt
}

// Input: the decrypted entries slice and cipher block made from the master
// password.
// Return: an error string, empty "" if no error
func WriteToFile(entries []Entry, ciphBlock cipher.Block) string {
	output, marshErr := yaml.Marshal(entries)

	if marshErr != nil {
		return " Error in yaml.Marshal\n\n " + marshErr.Error()
	}

	encryptedOutput := Encrypt(output, ciphBlock)

	// 0600 is the permissions that only this user can read/write to this file
	writeErr := os.WriteFile(FileName+".tmp", encryptedOutput, 0600)

	if writeErr != nil {
		return "Error in os.WriteFile\n\n " + writeErr.Error()
	}

	// Only will do this if the previous writing to a file worked, keeps it safe.
	os.Rename(FileName+".tmp", FileName)

	return ""
}

// Input: entries slice to write into and cipher block from the master password
// Return: an error string, empty "" if no error
func ReadFromFile(entries *[]Entry, ciphBlock cipher.Block) string {
	input, inputErr := os.ReadFile(FileName)

	if inputErr != nil {
		return " Error in os.ReadFile\n Make sure that a file named " +
			FileName +
			" exists.\n If there isn't one, run createEncr.go\n\n " +
			inputErr.Error()
	}

	decryptedInput := Decrypt(input, ciphBlock)
	unmarshErr := yaml.Unmarshal(decryptedInput, &entries)

	if unmarshErr != nil {
		return " Error in yaml.Unmarshal\n Make sure you write the correct password.\n\n " +
			unmarshErr.Error()
	}
	return ""
}
