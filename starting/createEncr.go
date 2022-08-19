/*	
	Used to create a file and encrypt it with a specific password

	!!!
	!!!change the name of the file later for os.WriteFile and os.Rename
	!!! ^^ because pass.go reads from pass.yaml
*/

package main 


import(
	"fmt"
	"os"
	"pass/encrypt"
	"time"
	"gopkg.in/yaml.v3"
	"strings"
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
	entries := []entry{entry{Name: "demo", Circulate: true},}

	var password string
	fmt.Println("write your password: ")
	fmt.Scan(&password)
	fmt.Print("\033[F\r", strings.Repeat(" ", len(password)))
	fmt.Println("")

	output, outputErr := yaml.Marshal(entries)

	if outputErr == nil{
		ciphBlock, boo, str := encrypt.KeyGeneration(password)

		fmt.Println("\ndoes the key match the key const in encrypt.go?", boo)

		if str == ""{
			encryptedOutput := encrypt.Encrypt(output, ciphBlock, false)

			// if the file doesn't exsist, os.WriteFile creates it
			writeErr := os.WriteFile("newpass.yaml.tmp", encryptedOutput, 0600)
			os.Rename("newpass.yaml.tmp", "newpass.yaml")

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
