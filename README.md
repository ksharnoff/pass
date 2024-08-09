# pass

## what is this project?
This is a password manager run entirely in the terminal. 

In the manager, and in this README, I have used `#` to represent a number and `str` to represent any random characters. 

## starting for the first time
Download all of the files (except the example folder of screenshots) into a folder called `pass`. Make sure there is a folder within `pass` called `encrypt` that `encrypt.go` is in. This is necessary as `encrypt.go` is imported for several different files.

Use the createEncr.go file to create and encrypt your file of password the first time. In all future times, you can just run pass.go and all should work.

There is also changeKey.go for decrypting the file and then encrypting it with a different key, in order to change your password or key parameters. 

## tview and visuals
I used the TUI [tview](https://github.com/rivo/tview). I used four types of primitives: input fields, lists, text boxes, and forms. In order to format them, I used flexes, pages, and grids. I used grids only to add borders around the primitives. 

The majority of the code is anonymous functions inside of func main in order to set up the primitives.

I coded it for a 84x28 window size with a text font of monaco, size 18. I chose this window size because it best fit the three columns for `/list` and `/find` with that font size.

It will work with all fonts (to my knowledge), however you may not be able to see all the items without scrolling or pressing the tab. If your font is bigger than monaco size 18, then you should use a bigger window size. 

## encryption and file writing
All of the entries are [marshaled](https://pkg.go.dev/gopkg.in/yaml.v3#Marshal) as if they were going to be written to a yaml file. Instead, that byte slice is entirely encrypted before being written to the file. Then, when reading from the file, the byte slice is decrypted and turned into the slice of entries. Therefore, the password to the password manager must be put in at the beginning before accessing any of the commands.

Argon2 is used to make a key and then the entries are encrypted with AES-256.

The way that the program knows if you put in the right password is if it can unmarshal the data successfully.

This password manager is unsuitable for cloud computing or a shared computer as the decrypted information is stored in the memory.

The encryption is in the file encrypt.go which must be in a folder called encrypt inside the greater pass folder as that is how the imports work. encrypt.go gets imported into not just pass.go but the files for setting up the program.

## commands
This section is about all of the actions that can be done with the password manager.
All of the commands are called through the command line at the bottom. 

### `/home`
![Picture of /home, a black screen with white dotted lines. There is a big empty box to the left, a column box on the right listing all the commands, and a blue input line at the bottom labeled “input: psst look to the right”.](https://github.com/ksharnoff/pass/blob/main/examples/%3Ahome%20Medium.jpeg)

`/home` is the starting screen once you’ve logged in. There’s nothing going on yet. The text on the right details the possible commands.

### `/help`
![Picture of /help, a black screen with white dotted lines. The big box to the left is full of white text, detailing information about the manager and specifics about each command type. There is a column box on the right listing all of the commands and a blue input line at the bottom.](https://github.com/ksharnoff/pass/blob/main/examples/%3Ahelp%20Medium.jpeg)

`/help` is similar to this README but it is condensed and in the password manager itself for ease of access. 

### `/open #`
![Picture of /open, opening entry 3, titled lkasdflkads. The big box to the left has each field of entry 3 written, tags, username, password, security question, notes, in circulation, date modified, date opened, and date created. The column box on the right listing all of the commands has separated out /edit and /copen to the top. There is a blue input line at the bottom.](https://github.com/ksharnoff/pass/blob/main/examples/%3Aopen%20Medium.jpeg)

`/open` is used to view an entry. It will include time information that is known. Passwords and security questions have their values printed in black text. Therefore, one can highlight it to see the values. 

### `/copen #`
![Picture of /copen, opening entry 3. The big box to the left is a numbered list (with letters of the alphabet instead of numbers). Each element of the list is a different field of entry 3, its name, tags, username, password, security question, notes, in circulation, date modified, date opened, and date created. The column box on the right listing all of the commands has separated out /edit and /open to the top. There is a blue input line at the bottom.](https://github.com/ksharnoff/pass/blob/main/examples/%3Acopen%20Medium.jpeg)

`/copen` is also used to view an entry. It is a list that is used to copy the data to the clipboard. With `/copen`, you select one of the fields and it copies itself to your clipboard.

### `/new`
![Picture of /new filled out with information.There is an input line for a name, here it is lkasdflkads. There is an input line for tags, here is xc8ouk. The box circulated is ticked. There are buttons titled new field, save entry, quit, notes, and edit field. Under this is a numerated list (with letters) of the email, password, and security question that were already filled in. The right box has instructions for how to move around /new and what you must finish in order to save an entry.](https://github.com/ksharnoff/pass/blob/main/examples/%3Anew%20Medium.jpeg)

`\new` has a form at the top with two input fields for the entry name and its tags. Then there are buttons for making a new field (username, password, or security question), saving, quitting, deleting, and making notes.
You must name the entry in order to save it. 

There is no limit to the number of usernames, passwords, or security questions you can make. They are all encrypted the same, except the values for passwords and security questions are blotted out when viewed. 

### `/copy #`
`/copy` is the same as `/new` except fields are already filled in with the information of entry #. 

### `/edit #`
![Picture of /edit, it looks much like /copen but without the dates listed. Instead, once you select one of the listed fields you can edit it. There is a numerated list showing its name, tags, username, password, security question, notes, in circulation, date modified, date opened, and date created. The column box to the right has instructions for how to move around /edit. There is a blue input line at the bottom.](https://github.com/ksharnoff/pass/blob/main/examples/%3Aedit%20Medium.jpeg)

`/edit` is used for editing an entry already made. It is a list with each field of the entry. You can select a field and then edit that specific one. 

### `/find str`
![Picture of /find ak. Listed in the left side box are six entries, their numbers and indices. Some of the names do not have the letters “ak”, therefore those entries must have it in the tags. The column box on the right listing all of the commands has separated out /open, /copen, and /edit to the top. There is the blue input line at the bottom.](https://github.com/ksharnoff/pass/blob/main/examples/%3Afind%20str%20Medium.jpeg)

`/find` is used to search the name and tags of all the entries for a string. It then returns the list of entries that contain that string. The entries are printed out following the same format as `/list`. The example above is searching for “ak”.

### `/list`
![Picture of /list. Listed in the left box are all of the entries, formatted into three columns. In this screenshot, there are only enough entries for one and a half columns. Each entry number and entry name is listed. Not all of the numbers are there - look at information about circulation. The column box on the right listing all of the commands has separated out /open, /copen, and /edit to the top. There is the blue input line at the bottom.](https://github.com/ksharnoff/pass/blob/main/examples/%3Alist%20Medium.jpeg)

`/list` is used to list all of the entries. /list is useful to see the index number of an entry to open it. `/list` prints the entries in three columns of a fixed size, therefore the entry name can get cut off. This is done with a single text box, using some string and math trickery. 

### `/pick` and `/picc`
![Picture of /pick. In a numerated list of letters, each entry, its number, and its tags are listed. If you click on any of the elements, it takes you to /open to view that entry. The column box to the right has instructions for how to move around /pick. There is a blue input line at the bottom.](https://github.com/ksharnoff/pass/blob/main/examples/%3Apick%20Medium.jpeg)

`/pick` and `/picc` look mostly identical. They are lists of all the entries, like `/list`, except you can select and open an entry. `/pick` will `/open` an entry while `/picc` will `/copen` an entry.

### `/comp # #`
![Picture of /comp 2 24. In the big square left text box, it says the entries’ indices and names, [2] akjsdf;k and [24] cvbnncbbcveqwbcnew. Under that, it says that akjsdf;k’s security question 0 = cvbnncbbcveqwbcnew’s password. There is a column box on the right listing all of the commands and a blue input line at the bottom.](https://github.com/ksharnoff/pass/blob/main/examples/%3Acomp%20Medium.jpeg)

`/comp # #` looks for duplicate passwords or answers to security questions between two entries.

### `/reused`
![Picture of /reused. In the big text box on the left, it has listed out blahg, then two entries and their indices. After that is aghikl and then two entries and their indices. After each entry is what field uses it, in these examples it is password or security question 0 or 1. There is a column box on the right listing all of the commands and a blue input line at the bottom.](https://github.com/ksharnoff/pass/blob/main/examples/%3Areused%20Medium.jpeg )

`/reused` shows any duplicate passwords or answers to security questions from all of the entries. The passwords (or answers) are printed in dark gray, but one can use their mouse to select the text to read it more clearly if needed.

## miscellaneous info

### mouse usage and clipboard
When the mouse is enabled in tview in order to change the focus or click buttons, one cannot select and copy any text. 

For ease of copying, there is `/copen #` which uses tview.List, where you select one of the fields and it copies to your clipboard.

### time values
Times are saved for the date created, date last modified, and the date last opened. 
Date last modified is only updated if any edits are made and saved in `/edit`.
Date last opened is modified if the entry is opened by `/open #` or  `/copen #`. 
These dates also work as security in case you notice irregularities.

### circulation
In each entry there is a boolean named `Circulate` which determines if the entry shows up in `/list`, `/pick`, and `/picc`. All commands that work on entries still work (`/edit`, `/open`, `/copy`, etc.). This can be used to reduce clutter of old entries without changing the entry numbers of later ones.
