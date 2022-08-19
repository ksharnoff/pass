/*	
	Used to create a file and encrypt it with a specific password

	!!!
	!!!change the name of the file later for os.WriteFile and os.Rename

	currently is not working, unmarshalling is not working in another file
	^^ issue may be with reading rather than writing
*/

package main 


import(
	"fmt"

	"os"

	"pass/encrypt"

	"time"

	"gopkg.in/yaml.v3"
)

type entry struct {
	Name string
	Tags string
	Usernames []Field
	Passwords []Field
	SecurityQ []Field
	Notes [6]string // maybe make this an 8 in the future?
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
	entries := []entry{entry{Name: "demo", Tags: "this is needed"}, entry{Name: "perhaps needed?", Circulate: true},}

	password := "foobar" // set this as the password that you want to encrypt with

	output, outputErr := yaml.Marshal(entries)

	if outputErr == nil{
		ciphBlock, boo, str := encrypt.KeyGeneration(password)

		fmt.Println("\n\n does the key match the key const in encrypt.go?", boo)

		if str == ""{
			encryptedOutput := encrypt.Encrypt(output, ciphBlock, false)

			// if the file doesn't exsist, os.WriteFile creates it
			writeErr := os.WriteFile("passteee.yaml.tmp", encryptedOutput, 0600)
			os.Rename("passteee.yaml.tmp", "passteee.yaml")

			if writeErr != nil{
				fmt.Println("error in os.writeFile \n", writeErr.Error())
			}else{
				fmt.Println("success, written!")
			}
		}else{
			fmt.Println("error in key generation: ", str)
		}
	}else{
		fmt.Println(outputErr.Error())
	}
}
