# password manager

## what is this project?
This is a password manager run entirely in the terminal. 

To quit running the program you should press control+c or terminate the terminal window that is running the program. 

In the manager, and in this README, I have used # to represent a number. 

## tview and visuals
I used the TUI [tview](https://github.com/rivo/tview). I used four types of primitives: input fields, lists, text boxes, and forms. In order to format them, I used flexes, pages, and grids. I used grids only to add borders around the primitives. 

All of the commands that are done are done through the command line at the bottom. 

The majority of the code is anonymous functions inside of func main in order to set up the primitives.

I coded it for a 84x28 window size with a text font of monaco, size 18. I chose this window size because it best fit the three columns for `/list` and `/find` with that font size.

It will work with all fonts (to my knowledge), however you may not be able to see all the items without scrolling or pressing the tab. Everything should still work and you should be able to access everything, it just may not look as organized. If your font is smaller than monaco size 18, then you should have a bigger window size. 

## encryption and file writing
All of the entries are [marshaled] (https://pkg.go.dev/gopkg.in/yaml.v3#Marshal) as if they were going to be written to a yaml file. Instead, that byte slice is entirely encrypted before being written to the file. Then, when reading from the file the byte slice is decrypted and then turned into the slice of entries. 
Therefore, the password to the password manager must be put in at the beginning before accessing any of the commands. 

This password manager is unsuitable for cloud computing or a shared computer as the decrypted information is stored in the memory. 

The encryption is in the file encrypt.go which must be in a folder called encrypt  inside the greater pass folder as that is how the imports work. encrypt.go gets imported into not just pass.go but the files for setting up the program. 

## commands
This section is about all of the actions that can be done with the password manager.

#### `/home`
{image of /home right here}
`/home` is the starting screen once you’ve logged in. There’s nothing going on yet. 

The focus is in the command line on the bottom. The text on the left details the possible commands. The mouse is enabled. 

#### `/help`
{image of /help}
`/help` is similar to this README but it is condensed and in the password manager itself for ease of access. 

#### `/open`
![example image of /open](https://github.com/ksharnoff/pass/blob/main/examples/:home%20Medium.jpeg)
`/open` is used to view an entry. 

#### `/copen`
{image of /copen}
`/copen` is also used to view an entry. 


#### `/new`
{image of /new}
`\new` has a form at the top with two input fields for the entry name and its tags. Then there are buttons for making a new field (username, password, or security question), saving, quitting, deleting, and making notes.
You must name the entry in order to save it. 

There is no limit to the number of usernames, passwords, or security questions you can make. They are all encrypted the same, except the values for passwords and security questions are blotted out when viewed. 

{image of /new with fields already added}
Once you have added fields, a new button appears that changes the focus to the list below to let you select a field to edit. The reason why I have a button to switch between the two sections is because I believe it is important to disable the mouse in `/new`. (Look at mouse usage and clipboard section) 


## starting for the first time
Use the createEncr.go file to create and encrypt your file with your password the first time. There is also changeKey.go for decrypting the file and then then encrypting it with a different key, in order to change your password or key parameters. 

In this repo, they are in their own file called starting for clarity, but they should be in the same folder as pass.go and pass.yaml when you run them. 

## miscellaneous info


#### mouse usage and clipboard
When the mouse is enabled in tview in order to change the focus or click buttons, one cannot select and copy any text. To combat this, sometimes the mouse in disabled. 

For ease of copying, there is `/copen #` which utilizes tview.List so that when you select one one of the fields it copies it to the clipboard.

#### time values
The time (as in time.Now()) are saved for the date created, date last modified, and the date last opened. 
Date last modified is only updated if any real edits are done in `/edit`, just opening the entry does not suffice.
Date last opened is modified if the entry is opened by `/open #` or  `/copen #`. 
Keeping these dates also works as security in case you notice irregularities.

#### in circulation
In each entry there is a boolean named `Circulate` which determines if the entry shows up in `/list` or `/pick`. All commands that work on entries still work (edit, open, copy, etc.). This can be used to reduce clutter of old entries.
