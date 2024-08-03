/*
	Creates & encrypts a new file to store the passwords for the password manager.
*/

package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"pass/encrypt"
	"strings"
	"time"
)

type entry struct {
	Name      string
	Tags      string
	Usernames []Field
	Passwords []Field
	SecurityQ []Field
	Notes     [6]string
	Circulate bool
	Created   time.Time
	Modified  time.Time
	Opened    time.Time
}
type Field struct {
	DisplayName string
	Value       string
}

func main() {
	_, statErr := os.Stat(encrypt.FileName) // os.Stat gets info about file
	if statErr == nil { // gives error if file doesn't exist
		fmt.Println("A file already exists under the name " + encrypt.FileName + "\nPlease:\n\t1) move that file to a different directory\nOR\n\t2) change the fileName variable in encrypt.go\n\nThis is protection so your data is not written over")
		os.Exit(1)
	}

	entries := []entry{entry{Name: "Demo", Circulate: true}}

	var password string
	fmt.Println("\nPlease write your password to encrypt your password manager.\nIf you forget it, there will be no way to access your passwords.\nAfter you press return, the password will disapear from terminal.")
	fmt.Scan(&password)
	fmt.Print("\033[F\r", strings.Repeat(" ", len(password)))
	fmt.Println("")
	fmt.Println("Do not quit the program! Things are happening!\n")

	output, outputErr := yaml.Marshal(entries)

	if outputErr != nil {
		printAndExit("Error in marshaling to yaml: " + outputErr.Error())
	}

	ciphBlock, keyErr := encrypt.KeyGeneration(password)

	if keyErr != "" {
		printAndExit("Error in key generation: " + keyErr)
	}

	encryptedOutput := encrypt.Encrypt(output, ciphBlock)

	writeErr := os.WriteFile(encrypt.FileName + ".tmp", encryptedOutput, 0600)
	os.Rename(encrypt.FileName + ".tmp", encrypt.FileName)

	if writeErr != nil {
		printAndExit("Error in os.writeFile: " + writeErr.Error())
	}

	fmt.Println("Success, file written!\nYou can run the password manager now.")
}

func printAndExit(error string) {
	fmt.Println(error)
	os.Exit(1)
}
