/*
	MIT License
	Copyright (c) 2022 Kezia Sharnoff

	changeKey.go
	Decrypts the file and then re-encrypts it with a different password
	(or with different key generation parameters)

	If you would like to change the key generation parameters, set the ones
	that you want to change to in KeyGeneration() in encrypt/encrypt.go and 
	have the keyGeneration() function in this file be the old parameters. 
	Also, set keyGenChange to true in this file. 
*/

package main

import (
	"fmt"
	"github.com/ksharnoff/pass/encrypt"
	"os"
	"strings"

	"crypto/aes"
	"crypto/cipher"
	"golang.org/x/crypto/argon2"
)

func main() {
	entries := []encrypt.Entry{}

	fmt.Println("\nFirst you will give your old password, then your new one.")
	fmt.Println("If you are just changing the key parameters, write the same password twice.")
	fmt.Println("\nIf you would like to change the key parameters go into the changeKey.go file and:")
	fmt.Println("\tChange keyGenChange to be true")
	fmt.Println("\tMake the keyGeneration function the same as it is in encrypt.go and:")
	fmt.Println("Next, go into encrypt.go")
	fmt.Println("\tChange the keyGeneration function to have different parameter\n")

	// if this is set to true then it will make the key to decrypt
	// using the keyGeneration function in this file, which will have
	// the old parameters to change away from. if you are just changing
	// what the password is and not the parameters, keep it as false
	keyGenChange := false

	var oldPass string
	fmt.Println("Write your old password: ")
	fmt.Scan(&oldPass)
	fmt.Print("\033[F\r", strings.Repeat(" ", len(oldPass)))
	fmt.Println("")

	badPass := true
	var newPass string
	for badPass {
		fmt.Println("Write your new password: ")
		fmt.Scan(&newPass)
		fmt.Print("\033[F\r", strings.Repeat(" ", len(newPass)))
		fmt.Println("")

		if (newPass == "/quit") || (newPass == "/q") ||
			(newPass == "quit") || (newPass == "q") {
			fmt.Println("Please chose a different password!\nIt cannot be /quit, /q, quit, or q\n")
			badPass = true
		} else {
			badPass = false
			break
		}
	}

	fmt.Println("THINGS ARE HAPPENING - DO NOT QUIT THE PROGRAM\n")

	var ciphBlockOld cipher.Block
	var oldKeyErr string

	if keyGenChange {
		// key generation using the old settings preserved in this file
		ciphBlockOld, oldKeyErr = keyGeneration(oldPass)
	} else {
		// key generation using the normal settings
		ciphBlockOld, oldKeyErr = encrypt.KeyGeneration(oldPass)
	}

	if oldKeyErr != "" {
		printAndExit("Error in key generation of old password: " + oldKeyErr)
	}

	readErr := encrypt.ReadFromFile(&entries, ciphBlockOld)
	
	if readErr != "" {
		printAndExit(readErr)
	}

	fmt.Println("Decrypted & unmarshled the input, success so far!")

	fmt.Println("\nTHINGS ARE HAPPENING - DO NOT QUIT THE PROGRAM\n")

	// generate new key that uses the settings chosen in encrypt/encrypt.go
	ciphBlockNew, newKeyErr := encrypt.KeyGeneration(newPass)

	if newKeyErr != "" {
		printAndExit("Error in key generation of new password: " + newKeyErr)
	}

	writeErr := encrypt.WriteToFile(entries, ciphBlockNew)

	if writeErr != "" {
		printAndExit(writeErr)
	}

	fmt.Println("Success! The passwords have been re-encrypted and written to file!")
}

// This is different than KeyGeneration in encrypt.go only so that
// it can be used to decrypt the file initially with parameters
// different than in pass/encrypt. So if you want to change the
// parameters, have the old ones here and the new ones you want to
// change to in encrypt.go.
func keyGeneration(password string) (cipher.Block, string) {
	if len([]byte(password)) < 1 {
		return nil, "Password given for key generation is zero characters"
	}

	salt := []byte("qwertyuiopasdfghjklzxcvbnm")

	// parameters currently in encrypt.go are: 1, 64*1024, 4, 32
	key := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	ciphBlock, err := aes.NewCipher(key)

	if err != nil {
		return nil, err.Error()
	}
	return ciphBlock, ""
}

// Input: error string to print.
// Then exits with status code 1. 
func printAndExit(error string) {
	fmt.Println(error)
	os.Exit(1)
}
