/*
	This decrypted the file and then reencrypts it with a different password 
*/

package main

import(
	"fmt"
	"os"
	"gopkg.in/yaml.v3"
	"pass/encrypt"
	"strings"
	"crypto/cipher"
	"golang.org/x/crypto/argon2"
	"crypto/aes"
	"time"
)

type entry struct {
	Name string
	Tags string
	Usernames []Field
	Passwords []Field
	SecurityQ []Field
	Notes [6]string 
	Circulate bool
	Created time.Time
	Modified time.Time
	Opened time.Time
}
type Field struct {
	DisplayName string
	Value string
}

func main(){
	entries := []entry{}

	var oldPass string
	fmt.Println("write your old password: ")
	fmt.Scan(&oldPass)
	fmt.Print("\033[F\r", strings.Repeat(" ", len(oldPass)))
	fmt.Println("")

	var newPass string
	fmt.Println("write your new password: ")
	fmt.Scan(&newPass)
	fmt.Print("\033[F\r", strings.Repeat(" ", len(newPass)))
	fmt.Println("")

	// if this is set to true then it will make the key to decrypt 
	// using the keygeneration function in this file, which will have
	// the old parameters to change away from
	keyGenChange := false 

	var ciphBlockOld cipher.Block 
	var strOld string

	if keyGenChange{
		ciphBlockOld, strOld = keyGeneration(oldPass)
	}else{
		ciphBlockOld, strOld = encrypt.KeyGeneration(oldPass)
	}
	if strOld == ""{
		input, inputErr := os.ReadFile(encrypt.FileName)
		if inputErr != nil{
			fmt.Println("error in os.ReadFile \n", inputErr.Error())
		}else{
			decryptedInput := encrypt.Decrypt(input, ciphBlockOld)

			unmarshErr := yaml.Unmarshal(decryptedInput, &entries)
			if unmarshErr != nil{
				fmt.Println("error in yaml.Unmarshal, wrong password mayhaps? \nor check that keyGenChange is false or true as needed \n", unmarshErr.Error())
			}else{
				fmt.Println("successfully unmarshaled the input, success so far.")	

				ciphBlockNew, strNew := encrypt.KeyGeneration(newPass)

				if strNew != ""{
					fmt.Println(" erroro!!" + strNew)
				}else{
					output, marshErr := yaml.Marshal(entries)
					if marshErr != nil{
						fmt.Println("error in yaml.marshal the entries \n", marshErr.Error())
					}else{
						encryptedOutput := encrypt.Encrypt(output, ciphBlockNew, false)

						writeErr := os.WriteFile(encrypt.FileName+".tmp", encryptedOutput, 0600)
						os.Rename(encrypt.FileName+".tmp", encrypt.FileName)

						if writeErr != nil{
							fmt.Println("error in os.writeFile \n", writeErr.Error())
						}else{
							fmt.Println("success! changed the password, wrote to the file!")
						}
					}
				}
			}
		}
	}else{
		fmt.Println("errror: " + strOld)
	}
}

// This is different than KeyGeneration in encrypt.go only so that
// this can be used to decrypt the file initially with parameters
// different than in pass/encrypt. So if you want to change the 
// parameters, have the old ones here and the new ones you want to 
// change in encrypt.go. 
func keyGeneration(password string) (cipher.Block, string){
	if len([]byte(password)) < 1{
		return nil, "password for key generation is too short, string empty"
	}
	salt := []byte("qwertyuiopasdfghjklzxcvbnm")

	// parameters currently in encrypt.go are: 4, 2048*1024, 4, 32
	key := argon2.IDKey([]byte(password), salt, 4, 2048*1024, 4, 32)

	ciphBlock, err := aes.NewCipher(key)

	if err != nil{
		return nil, err.Error()
	}
	return ciphBlock, ""
}
