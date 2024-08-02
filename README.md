# pass

## what is this project?
This is a password manager run entirely in the terminal. 

To quit running the program you should press control+c or terminate the terminal window that is running the program. 

In the manager, and in this README, I have used # to represent a number. 

## tview and visuals
I used the TUI [tview](https://github.com/rivo/tview). I used four types of primitives: input fields, lists, text boxes, and forms. In order to format them, I used flexes, pages, and grids. I used grids only to add borders around the primitives. 

The majority of the code is anonymous functions inside of func main in order to set up the primitives.

I coded it for a 84x28 window size with a text font of monaco, size 18. I chose this window size because it best fit the three columns for `/list` and `/find` with that font size.

It will work with all fonts (to my knowledge), however you may not be able to see all the items without scrolling or pressing tab. Everything should still work and you should be able to access everything, it just may not look as organized. If your font is smaller than monaco size 18, then you should have a bigger window size. 

## encryption and file writing
All of the entries are [marshaled](https://pkg.go.dev/gopkg.in/yaml.v3#Marshal) as if they were going to be written to a yaml file. Instead, that byte slice is entirely encrypted before being written to the file. Then, when reading from the file the byte slice is decrypted and then turned into the slice of entries. 
Therefore, the password to the password manager must be put in at the beginning before accessing any of the commands. 

Argon2 is used to make the key and the entries are encrypted with AES-256. 

The way that the program knows if you put in the right password is if it can unmarshal the data successfully.

This password manager is unsuitable for cloud computing or a shared computer as the decrypted information is stored in the memory. 

The encryption is in the file encrypt.go which must be in a folder called encrypt inside the greater pass folder as that is how the imports work. encrypt.go gets imported into not just pass.go but the files for setting up the program. 

## commands
This section is about all of the actions that can be done with the password manager.
All of the commands are called through the command line at the bottom. 

#### `/home`
![Picture of /home](https://github.com/ksharnoff/pass/blob/main/examples/:home%20Medium.jpeg)

`/home` is the starting screen once you’ve logged in. There’s nothing going on yet. The text on the left details the possible commands. A command that was written on the side after this picture is `/quit` which quits the application. 

#### `/help`
![Picture of /help](https://github.com/ksharnoff/pass/blob/main/examples/:help%20Medium.jpeg)

`/help` is similar to this README but it is condensed and in the password manager itself for ease of access. 

#### `/open #`
![Example of /open](https://github.com/ksharnoff/pass/blob/main/examples/:open%20Medium.jpeg)

`/open` is used to view an entry. It will include time information that is known. Passwords and security questions will also have their values printed, except they’ll be printed out in black text. Therefore, one can highlight it to see the values. 

#### `/copen #`
![Example of /copen](https://github.com/ksharnoff/pass/blob/main/examples/:copen%20%20Medium.jpeg)
`/copen` is also used to view an entry. It is a list that is used to copy data to the clipboard more easily. With `/copen` you select one of the fields and it copies itself to your clipboard.

#### `/new`
![Picture of blank /new](https://github.com/ksharnoff/pass/blob/main/examples/:new%20Medium.jpeg)

`\new` has a form at the top with two input fields for the entry name and its tags. Then there are buttons for making a new field (username, password, or security question), saving, quitting, deleting, and making notes.
You must name the entry in order to save it. 

There is no limit to the number of usernames, passwords, or security questions you can make. They are all encrypted the same, except the values for passwords and security questions are blotted out when viewed. 

![Picture of /new with fields](https://github.com/ksharnoff/pass/blob/main/examples/:new%20fields%20Medium.jpeg)

Once you have added fields, a new button appears that changes the focus to the list below to let you select a field to edit. The reason why I have a button to switch between the two sections is because I believe it is important to disable the mouse in `/new`. (Look at mouse usage and clipboard section) 

#### `/copy #`
`/copy` is the same as `/new` except fields are already filled in with the information of entry #. 

#### `/edit #`
![Example of /edit](https://github.com/ksharnoff/pass/blob/main/examples/:edit%20Medium.jpeg)

`/edit` is used for editing an entry already made. It is a list with each field of the entry. You can select a field and then edit that specific one. 

#### `/find str`
![Example of /find](https://github.com/ksharnoff/pass/blob/main/examples/:find%20str%20Medium.jpeg)

`/find` is used to search the name and tags of all the entries for a string. It then returns the list of entries that contain that string. The entries are printed out following the same format as `/list`.

#### `/list`
![Example of /list](https://github.com/ksharnoff/pass/blob/main/examples/:list%20Medium.jpeg)

`/list` is used to list all of the entries. You look at /list to see the index number of an entry to open it. `/list` prints the entries in three columns of a fixed size, therefore the entry name can get cut off. This is done with a single text box, using some string and math trickery. 

#### `/pick` and `/picc`
![Example of /pick](https://github.com/ksharnoff/pass/blob/main/examples/:pick%20Medium.jpeg)

`/pick` and `/picc` look mostly identical. They are lists of all the entries, like `/list`, except you can select and open an entry. 

## starting for the first time
Use the createEncr.go file to create and encrypt your file with your password the first time. There is also changeKey.go for decrypting the file and then then encrypting it with a different key, in order to change your password or key parameters. If you just run createEncr.go then run pass.go, all should work. 

## miscellaneous info


#### mouse usage and clipboard
When the mouse is enabled in tview in order to change the focus or click buttons, one cannot select and copy any text. To combat this, sometimes the mouse in disabled. 

For ease of copying, there is `/copen #` which utilizes tview.List so that when you select one one of the fields it copies it to the clipboard.

#### time values
Times are saved for the date created, date last modified, and the date last opened. 
Date last modified is only updated if any real edits are done in `/edit`, just opening the entry does not suffice.
Date last opened is modified if the entry is opened by `/open #` or  `/copen #`. 
Keeping these dates also works as security in case you notice irregularities.

#### in circulation
In each entry there is a boolean named `Circulate` which determines if the entry shows up in `/list` and `/pick`. All commands that work on entries still work (edit, open, copy, etc.). This can be used to reduce clutter of old entries.
