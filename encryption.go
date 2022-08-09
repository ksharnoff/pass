
/*
	- update parameters in keygeneration

*/


package main 

import (
	// for writing/reading from/to file 
	"os" 
	"gopkg.in/yaml.v3" 

	// for encrypting things
	"golang.org/x/crypto/argon2"
	"strconv"
	"crypto/aes"
	"crypto/cipher"
)

//const comparison = "trans rights R human rights"

// writes to the pass.yaml file, if it fails then it returns a string with errors
func writeToFile(entries []entry) string{
	output, marshErr := yaml.Marshal(entries)
	if marshErr != nil{
		return "error in yaml.marshal the entries \n" + marshErr.Error()
	}else{
		// conventions of writing to a temp file is write to .tmp
		writeErr := os.WriteFile("pass.yaml.tmp", output, 0600) // 0600 is the permissions, that only this user can read/write/excet to this file
		os.Rename("pass.yaml.tmp", "pass.yaml") // only will do this if the previous thing worked correctly, helps to save the data :)

		if writeErr != nil{
			return "error in os.writeFile \n" + writeErr.Error()
		}else{
			return ""
		}
	}
}

// if it works then it should return "", if not then it will return the errors in a string format
func readFromFile(entries *[]entry) string{
	input, inputErr := os.ReadFile("pass.yaml")
	if inputErr != nil{
		return "error in os.ReadFile \n" + inputErr.Error()
	}else{
		unmarshErr := yaml.Unmarshal(input, &entries)
		if unmarshErr != nil{
			return "error in yaml.Unmarshal \n" + unmarshErr.Error()
		}else{
			return ""		
		}
	}
}


// right now the correct password is foobar

// makes a key, returns a chiper block
// then checks with correctKey function if the key is correct -- if it's correct then true is returned
func keyGeneration(password string) (Cipher.Block, bool, string){

	if len([]byte(password)) < 1{
		return nil, false, "password for key generation is empty"
	}

	// salt generation is going to be the same thing every time

	salt := []byte("qwertyuiopasdfghjklzxcvbnm")

	// THE PARAMETERS MUST BE ADJUSTED
	key := argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, 32)

	ciphBlock, err := aes.NewCipher(key)

	if err != nil{
		return nil, false, err.Error()
	}
	return chipBlock, correctKey(ciphBlock), ""

}


func correctKey(block Cipher.Block) bool{


	return true

}







