/*
	MIT License
	Copyright (c) 2022 Kezia Sharnoff

	createEncr.go
	Creates & encrypts a new file to store the passwords for the password manager.
*/

package main

import (
	"fmt"
	"github.com/ksharnoff/pass/encrypt"
	"os"
	"strings"
)

func main() {
	_, statErr := os.Stat(encrypt.FileName) // os.Stat gets info about file
	// if there was no error getting the file, then it must already exist
	if statErr == nil {
		fmt.Println("A file already exists under the name " + encrypt.FileName)
		fmt.Println("Please:")
		fmt.Println("\t1) move that file to a different directory")
		fmt.Println("OR")
		fmt.Println("\t2) change the fileName variable in encrypt.go")
		fmt.Println("\nThis is protection so your data is not written over")
		os.Exit(1)
	}

	entries := []encrypt.Entry{encrypt.Entry{Name: "Demo", Circulate: true}}

	password := "/quit"

	for password == "/quit" {
		fmt.Println("\n----------")
		fmt.Println("Please write your password to encrypt your password manager.")
		fmt.Println("If you forget it, there will be no way to access your passwords.")
		fmt.Println("After you press return, the password will disappear from terminal.")
		fmt.Scan(&password)
		fmt.Print("\033[F\r", strings.Repeat(" ", len(password)))
		fmt.Println("")

		if (password == "/quit") || (password == "/q") ||
			(password == "quit") || (password == "q") {
			fmt.Println("Please chose a different password!\nIt cannot be /quit, /q, quit, or q")
			password = "/quit"
		}
	}

	fmt.Println("\nTHINGS ARE HAPPENING - DO NOT QUIT THE PROGRAM\n")

	ciphBlock, keyErr := encrypt.KeyGeneration(password)

	if keyErr != "" {
		printAndExit(fmt.Sprint("Error in key generation:\n" + keyErr))
	}

	writeErr := encrypt.WriteToFile(entries, ciphBlock)

	if writeErr != "" {
		printAndExit(writeErr)
	}

	fmt.Println("Success, file written!\nYou can run the password manager now.")
}

func printAndExit(error string) {
	fmt.Println(error)
	os.Exit(1)
}
