/*
	you take the input that's printed out and you set that as the const encryptedPhrase in encrypt.go to match your password

	this is done and works
*/

package main

import(
	"fmt"

	"crypto/aes"

	"encoding/base64"

	"pass/encrypt"

	"strings"

	"time" // just for running tests haha
)

// example but with the the password "foobar"
//const encryptedPhrase = "AAAAAAAAAAAAAAAAAAAAANGYxl8FWvWBoG+/KRgGRwSSwiXNG4VXA9jIQU5gmVIh"


func main(){
	var password string
	fmt.Println("write your password: ")
	fmt.Scan(&password)
	fmt.Print("\033[F\r", strings.Repeat(" ", len(password)))
	fmt.Println("")

	start := time.Now()


	// no padding needs to be done
	input := []byte("trans rights R human rights 1234")
	
	if len(input)%aes.BlockSize != 0{
		fmt.Println("plaintext is not a multiple of the block size:  ", len(input)%aes.BlockSize)
	}else{

		ciphBlock, boo, str := encrypt.KeyGeneration(password)

		// need to have something done w boo lol, is not necessary
		if boo{
			fmt.Println("the new key being generated is the same as the previous key!")
		}

		if str != ""{
			fmt.Println("error in making cipher block", str)
		}else{

			encrypted := encrypt.Encrypt(input, ciphBlock, true)
			
			encoder := base64.NewEncoding("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890+/")

			encryptedStr := encoder.EncodeToString(encrypted)

			fmt.Println("")
			fmt.Println(encryptedStr)


			end := time.Now()
			fmt.Printf("\n\nthis took %v to run.\n", end.Sub(start))

		}		
	}	
}
