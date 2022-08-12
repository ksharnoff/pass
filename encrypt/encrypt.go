/*
	- update parameters in keygeneration


	the way it's going to work::
		-- have the thing at the start and make the key then use the key as input into the writing/reading file

	style is historically that you don't pass around the cipher you pass the generated key
	change encrypt and decrypt to take the aes key
	literally just styalistic and that's it  
	because if you look at the function in the vacuum then it is less confusing to take in the key
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
const encryptedPlaintext = "AAAAAAAAAAAAAAAAAAAAAMeOBaInl79wFv9Lxz0rlWZ96VlFVTO5E6oEXKtJ1zdb"
/*
// writes to the pass.yaml file, if it fails then it returns a string with errors
func WriteToFile(entries []entry, ciphBlock cipher.Block) string{
	output, marshErr := yaml.Marshal(entries)
	if marshErr != nil{
		return "error in yaml.marshal the entries \n" + marshErr.Error()
	}else{

		encryptedOutput := Encrypt(output, ciphBlock, false)

		// conventions of writing to a temp file is write to .tmp
		writeErr := os.WriteFile("pass.yaml.tmp", encryptedOutput, 0600) // 0600 is the permissions, that only this user can read/write/excet to this file
		os.Rename("pass.yaml.tmp", "pass.yaml") // only will do this if the previous thing worked correctly, helps to save the data :)

		if writeErr != nil{
			return "error in os.writeFile \n" + writeErr.Error()
		}else{
			return ""
		}
	}
}

// if it works then it should return "", if not then it will return the errors in a string format
func ReadFromFile(entries *[]entry, ciphBlock cipher.Block) string{
	input, inputErr := os.ReadFile("pass.yaml")
	if inputErr != nil{
		return " error in os.ReadFile \n" + inputErr.Error()
	}else{
		// first we decrypt it!
		decryptedInput := Decrypt(input, ciphBlock)

		unmarshErr := yaml.Unmarshal(decryptedInput, &entries)
		if unmarshErr != nil{
			return " error in yaml.Unmarshal \n" + unmarshErr.Error()
		}else{
			return ""		
		}
	}
}
*/

// makes a key, returns a chiper block
// then checks with correctKey function if the key is correct -- if it's correct then true is returned
func KeyGeneration(password string) (cipher.Block, bool, string){

	if len([]byte(password)) < 1{
		return nil, false, "password for key generation is too short, string empty"
	}

	// salt generation is going to be the same thing every time
	salt := []byte("qwertyuiopasdfghjklzxcvbnm")

	// max's argon parameters are 5, 1,000,000, 1 (for first three numbers)
	// time cost is number of passes over memory
	// memory used is in kilobytes
	// memory recc: 2*1024*1024

	// try and see how much time it takes to do stuff and how long you can stand it being slow and then do it 

	// parallelism factor allows it to be run in paralell which is good for me because i am one computer -- bad for someone trying to crack it -- probably leave as 4 -max
	// some cpu run several threads on one core but apple silicon doesn't do that 
	// my memory usage is currently 64 megabytes
	// THE PARAMETERS MUST BE ADJUSTED -- make sure they are the same as settingUpKeys.go
	key := argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, 32)

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
