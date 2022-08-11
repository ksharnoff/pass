/*

	to do: make it take the password as input so it's not ever written in a file

*/

package main

import(
	"fmt"

	"golang.org/x/crypto/argon2"
	"crypto/aes"
	"crypto/cipher"

	"encoding/base64"
)

//const encryptedPhrase = "x55FoieXrzAW/wvHP3uVZnzlWUVVM8gTmgRcq1nTN2s="

func main(){

	password := "foobar"

	// no padding needs to be done
	input := []byte("trans rights R human rights 1234")
	
	if len(input)%aes.BlockSize != 0{
		fmt.Println("plaintext is not a multiple of the block size:  ", len(input)%aes.BlockSize)
	}else{

		//fmt.Println("starting the program wooooo")

		salt := []byte("qwertyuiopasdfghjklzxcvbnm")

		// THE PARAMETERS OF THIS KEY MUST MATCH THE PARAMATERS IN ENCRYPTION.GO
		key := argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, 32) // makes a key out of the password manager password


		ciphBlock, err := aes.NewCipher(key) // makes a cipher block out of the key that can be used to encyrpt/decrypt stuff

		if err != nil{
			fmt.Println("error in making cipher block", err.Error())
		}else{
			// make IV the same each time, a blank [0 0 0 0 ...]
			encrypted := make([]byte, aes.BlockSize+len(input))

			iv := encrypted[:aes.BlockSize]


			encryptBlock := cipher.NewCBCEncrypter(ciphBlock, iv)
			
			encryptBlock.CryptBlocks(encrypted[aes.BlockSize:], input)

			skip()
			
			fmt.Println("encrypted! as byte in decimal form", encrypted)

			encoder := base64.NewEncoding("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890+/")

			encryptedStr := encoder.EncodeToString(encrypted)

			skip()
			fmt.Println(encryptedStr)

		}		
	}	
}


func skip(){
	fmt.Println("")

}
