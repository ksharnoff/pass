/*

	to do: make it take the password as input so it's not ever written in a file

*/

package main

import(
	"fmt"

	"crypto/aes"

	"encoding/base64"

	"pass/encrypt"
)

//const encryptedPhrase = "x55FoieXrzAW/wvHP3uVZnzlWUVVM8gTmgRcq1nTN2s="
// AAAAAAAAAAAAAAAAAAAAAMeOBaInl79wFv9Lxz0rlWZ96VlFVTO5E6oEXKtJ1zdb


func main(){

	password := "foobar"

	// no padding needs to be done
	input := []byte("trans rights R human rights 1234")
	
	if len(input)%aes.BlockSize != 0{
		fmt.Println("plaintext is not a multiple of the block size:  ", len(input)%aes.BlockSize)
	}else{

		ciphBlock, boo, str := encrypt.KeyGeneration(password)

		if boo{
			boo = !boo
		}

		if str != ""{
			fmt.Println("error in making cipher block", str)
		}else{

			encrypted := encrypt.Encrypt(input, ciphBlock, true)
			
			fmt.Println("encrypted! as byte in decimal form", encrypted)

			encoder := base64.NewEncoding("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890+/")

			encryptedStr := encoder.EncodeToString(encrypted)

			skip()
			fmt.Println(encryptedStr)

		}		
	}	
}

/*

	// makes a key, returns a chiper block
	// then checks with correctKey function if the key is correct -- if it's correct then true is returned
	func keyGeneration(password string) (cipher.Block, string){

		if len([]byte(password)) < 1{
			return nil, "password for key generation is too short, string empty"
		}

		// salt generation is going to be the same thing every time
		salt := []byte("qwertyuiopasdfghjklzxcvbnm")

		// THE PARAMETERS MUST BE ADJUSTED -- make sure they are the same as settingUpKeys.go
		key := argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, 32)

		ciphBlock, err := aes.NewCipher(key)

		if err != nil{
			return nil, err.Error()
		}
		return ciphBlock, ""
	}

	func encrypt(plaintext []byte, ciphBlock cipher.Block) []byte{
		// adds padding in form of "/n"
		if len(plaintext)%aes.BlockSize != 0{
			for i := len(plaintext)%aes.BlockSize; i < aes.BlockSize; i++{
				plaintext = append(plaintext, 0x0A) // 0x0A = []byte("\n")
			}
		}

		encrypt := make([]byte, aes.BlockSize+len(plaintext))

		iv := encrypt[:aes.BlockSize]

		encryptBlock := cipher.NewCBCEncrypter(ciphBlock, iv)

		encryptBlock.CryptBlocks(encrypt[aes.BlockSize:], plaintext)

		return encrypt
	}

	func decrypt(encrypted []byte, ciphBlock cipher.Block) []byte{
		iv := encrypted[:aes.BlockSize]
		encrypted = encrypted[aes.BlockSize:]

		decryptBlock := cipher.NewCBCDecrypter(ciphBlock, iv)

		decrypt := make([]byte, len(encrypted))

		decryptBlock.CryptBlocks(decrypt, encrypted) // not sure if this works, if not then write it as "encrypted, encrypted" which will for sure work -- it's being written like this for consistency

		return decrypt
	}
*/


func skip(){
	fmt.Println("")

}
