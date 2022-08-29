/*	
	Used to create a file and encrypt it with a specific password
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
	entries := []entry{entry{Name: "demo", Circulate: true},}

	var password string
	fmt.Println("write your password: ")
	fmt.Scan(&password)
	fmt.Print("\033[F\r", strings.Repeat(" ", len(password)))
	fmt.Println("")

	output, outputErr := yaml.Marshal(entries)

	if outputErr == nil{
		ciphBlock, str := encrypt.KeyGeneration(password)

		if str == ""{
			encryptedOutput := encrypt.Encrypt(output, ciphBlock, false)

			// if the file doesn't exsist, os.WriteFile creates it
			writeErr := os.WriteFile(encrypt.FileName + ".tmp", encryptedOutput, 0600)
			os.Rename(encrypt.FileName + ".tmp", encrypt.FileName)

			if writeErr != nil{
				fmt.Println("error in os.writeFile \n", writeErr.Error())
			}else{
				fmt.Println("success, file written!")
			}
		}else{
			fmt.Println("error in key generation: ", str)
		}
	}else{
		fmt.Println(outputErr.Error())
	}
}
