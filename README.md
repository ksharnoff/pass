# pass

## what is this project?
This is a password manager run entirely in the terminal.

The password manager stores an encrypted list of entries -- where each entry can store its name, tags, and unlimited URLs, usernames, passwords, security questions, or notes. You can search the names, tags, and URLs. You can view and easily copy to your clipboard any field of an entry. You can see a list of entries that have reused passwords. 

## starting for the first time
- Download the zip of the latest release or run `git clone https://github.com/ksharnoff/pass.git`.
- Unzip the zip and then run `go mod tidy` to install all necessary dependencies. 
- Use `createEncr.go` to create and encrypt your file of passwords the first time, `go run createEncr.go`. If you forget your password, you cannot decrypt the file later. 
	- If you would like to change your password or the key generation parameters in the future, use `changeKey.go` by `go run changeKey.go`. You will have to edit `changeKey.go` and `encrypt/encrypt.go` to change the key parameters
- You can successfully run the password manager now, `go run pass.go`. I recommend compiling it by `go build pass.go` to run it quickly as an executable, `./pass`.

The encrypted passwords will be stored in a file named `pass.yaml` in the `pass` directory. 

## TUI `tview`
I used the TUI [`tview`](https://github.com/rivo/tview). I used four types of primitives: input fields, lists, text boxes, and forms. In order to format them, I used flexes, pages, and grids. I used grids only to add borders around the primitives. 

The password manager will scale to work on any sized terminal, with any font. I recommend using a monospaced font, for example, `Monaco`.

## encryption and file writing
All of the entries are [marshaled](https://pkg.go.dev/gopkg.in/yaml.v3#Marshal) as if they were going to be written to a YAML file. Instead, that list of byte is entirely encrypted before being written to a file. Then, after reading from the file, the list of bytes is decrypted and turned into the list of entries. The password manager password (the master password) must be inputted before viewing any fields. The program verifies that the correct password was inputted if it can successfully unmarshal the data into a list of entries; the master password is never stored.

Argon2 is used to make a key and the entries are encrypted with AES-256.

This password manager is unsuitable for cloud computing or a shared computer as the decrypted information is stored in the memory. I do not believe this to be a significant risk on a single user computer. 

The encryption is done in `encrypt.go` which must be in a directory called `encrypt` inside the greater `pass` directory as `encrypt.go` is used for several other files. 

## commands
This section is about all of the actions that can be done with the password manager. `#` is written to mean any number and `str` is written to mean any set of characters. 
All of the commands are called through the command line at the bottom. The terminal in these photos is 84x28 with the font Monaco, 18pt. 

### `/open #`
![Picture of /open, opening entry 3, titled lkasdflkads. The big box to the left has each field of entry 3 written, tags, username, password, security question, notes, in circulation, date modified, date opened, and date created. The column box on the right listing all of the commands has separated out /edit and /copen to the top. There is a blue input line at the bottom.](https://github.com/ksharnoff/pass/blob/main/examples/open.jpeg)

`/open` is used to view an entry. It will include time information that is known. Passwords and security questions have their values printed in black text. Therefore, one can highlight it to see the values. 

### `/copen #`
![Picture of /copen, opening entry 3. The big box to the left is a numbered list (with letters of the alphabet instead of numbers). Each element of the list is a different field of entry 3, its name, tags, username, password, security question, notes, in circulation, date modified, date opened, and date created. The column box on the right listing all of the commands has separated out /edit and /open to the top. There is a blue input line at the bottom.](https://github.com/ksharnoff/pass/blob/main/examples/copen.jpeg)

`/copen` is also used to view an entry. It is a list that is used to copy the data to the clipboard. With `/copen`, you select one of the fields and it copies itself to your clipboard.

### `/new`
![Picture of /new filled out with information.There is an input line for a name, here it is lkasdflkads. There is an input line for tags, here is xc8ouk. The box circulated is ticked. There are buttons titled new field, save entry, quit, notes, and edit field. Under this is a numerated list (with letters) of the email, password, and security question that were already filled in. The right box has instructions for how to move around /new and what you must finish in order to save an entry.](https://github.com/ksharnoff/pass/blob/main/examples/new.jpeg)

`\new` has a form at the top with two input fields for the entry name and its tags. Then there are buttons for making a new field (username, password, or security question), saving, quitting, deleting, and making notes.
You must name the entry in order to save it. 

There is no limit to the number of usernames, passwords, or security questions you can make. They are all encrypted the same, except the values for passwords and security questions are blotted out when viewed. 

### `/copy #`
`/copy` is the same as `/new` except fields are already filled in with the information of entry #. 

### `/edit #`
![Picture of /edit, it looks much like /copen but without the dates listed. Instead, once you select one of the listed fields you can edit it. There is a numerated list showing its name, tags, username, password, security question, notes, in circulation, date modified, date opened, and date created. The column box to the right has instructions for how to move around /edit. There is a blue input line at the bottom.](https://github.com/ksharnoff/pass/blob/main/examples/edit.jpeg)

`/edit` is used for editing an entry already made. It is a list with each field of the entry. You can select a field and then edit that specific one. 

### `/find str`
![Picture of /find ak. Listed in the left side box are six entries, their numbers and indices. Some of the names do not have the letters “ak”, therefore those entries must have it in the tags. The column box on the right listing all of the commands has separated out /open, /copen, and /edit to the top. There is the blue input line at the bottom.](https://github.com/ksharnoff/pass/blob/main/examples/find%20str.jpeg)

`/find` is used to search the name and tags of all the entries for a string. It then returns the list of entries that contain that string. The entries are printed out following the same format as `/list`. The example above is searching for “ak”.

### `/list`
![Picture of /list. Listed in the left box are all of the entries, formatted into three columns. In this screenshot, there are only enough entries for one and a half columns. Each entry number and entry name is listed. Not all of the numbers are there - look at information about circulation. The column box on the right listing all of the commands has separated out /open, /copen, and /edit to the top. There is the blue input line at the bottom.](https://github.com/ksharnoff/pass/blob/main/examples/list.jpeg)

`/list` is used to list all of the entries. /list is useful to see the index number of an entry to open it. `/list` prints the entries in three columns of a fixed size, therefore the entry name can get cut off. This is done with a single text box, using some string and math trickery. 

### `/pick` and `/picc`
![Picture of /pick. In a numerated list of letters, each entry, its number, and its tags are listed. If you click on any of the elements, it takes you to /open to view that entry. The column box to the right has instructions for how to move around /pick. There is a blue input line at the bottom.](https://github.com/ksharnoff/pass/blob/main/examples/pick.jpeg)

`/pick` and `/picc` look mostly identical. They are lists of all the entries, like `/list`, except you can select and open an entry. `/pick` will `/open` an entry while `/picc` will `/copen` an entry.

### `/comp # #`
![Picture of /comp 2 24. In the big square left text box, it says the entries’ indices and names, [2] akjsdf;k and [24] cvbnncbbcveqwbcnew. Under that, it says that akjsdf;k’s security question 0 = cvbnncbbcveqwbcnew’s password. There is a column box on the right listing all of the commands and a blue input line at the bottom.](https://github.com/ksharnoff/pass/blob/main/examples/comp.jpeg)

`/comp # #` looks for duplicate passwords or answers to security questions between two entries.

### `/reused`
![Picture of /reused. In the big text box on the left, it has listed out blahg, then two entries and their indices. After that is aghikl and then two entries and their indices. After each entry is what field uses it, in these examples it is password or security question 0 or 1. There is a column box on the right listing all of the commands and a blue input line at the bottom.](https://github.com/ksharnoff/pass/blob/main/examples/reused.jpeg )

`/reused` shows any duplicate passwords or answers to security questions from all of the entries. The passwords (or answers) are printed in dark gray, but one can use their mouse to select the text to read it more clearly if needed.

## miscellaneous info

### mouse usage and clipboard
When the mouse is enabled in `tview` in order to change the focus or click buttons, you cannot select and copy any text. 

For ease of copying, there is `/copen #`, where you select one of the fields from the list and it copies to your clipboard. Once you quit out of viewing the entry, your clipboard is cleared.

### time values
Times are saved for the date created, date last modified, and the date last opened. 
Date last modified is only updated if any edits are made and saved in `/edit`.
Date last opened is modified if the entry is opened by `/open #` or  `/copen #`. 
These dates also work as security in case of irregularities.

### circulation
In each entry there is a boolean named `Circulate` which determines if the entry shows up in `/list`, `/pick`, and `/picc`. All commands that work on entries still work (`/edit`, `/open`, `/copy`, etc.). This can be used to reduce clutter of old entries without changing the entry numbers of later ones.
