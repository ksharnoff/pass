# password manager

## what is this project?
This is a password manager run entirely in the terminal. 

## tview
- used grids to put boxes around text, lists, forms, flexes
- .


## terminal visual decisions
- there is a command line at the bottom that all inputs and made
- the list of possible commands are written in a box to the right, do /help for more details
- copy /help here into this readme?
- a lot of the functions are anonymous functions inside of func main, that is because a lot of the functions are just setting up the different Primitives (forms, lists) of tview
- i coded it for a 84x28 size window with text font of monaco, size 18. it will work at all sizes and with all texts (to my knowledge), however you may not be able to see all the buttons without scrolling or pressing tab. everything should still work, you should still be able to access everything, it just may not look all organized.
	-- i found this the best set up horizontally for /find and /list. the 


#### mouse usage and clipboard
^^ maybe move to another 
When the mouse is enabled in tview in order to change the focus or click buttons, one cannot select and copy any text. To combat this, sometimes the mouse in disabled. 

For ease of copying, there is `/copen #` which utilizes tview.List so that when you select one one of the fields it copies it to the clipboard. 

## encryption and file writing
All of the entries are [marshalled] (https://pkg.go.dev/gopkg.in/yaml.v3#Marshal) as if they were going to be written to a yaml file. Instead, that byte slice is entirely encrypted before being written to the file. 

This password manager is unsuitable for cloud computing or a shared computer as the decrypted information is stored in the memory. 

- don't use on a cloud hosted computer!! or a shared computer!! 
  - some of the sensitive data is briefly stored in the memory 
- 

homeCommands := " commands\n --------\n /home \n /help \n /new \n /find str\n /edit # \n /open # \n /copen # \n /list \n /pick \n /picc \n /copy \n /test"

## commands
This section is all of the actions that can be done with the password manager.

^^ move this section further up

#### `/home`
{image of /home right here}
`/home` is the starting screen once you’ve logged in. It is a blank tview.Box with the title of “sad, empty box” in the middle top. 

The focus is in the command line on the bottom. The text on the left details the possible commands. The mouse is enabled. 

#### `/help`
{image of /help}
`/help` is similar to this README but it is condensed and in the password manager itself for ease of access. 

Same specs (specifications) as /home. 

#### `/new`
{image of /new}
`\new` has a tview.Form at the top with has two input fields for the entry name and its tags. Then there are buttons for making a new field (username, password, or security question), saving, quitting, deleting, making notes.
You must name the entry in order to save it. 

{image of /new with an add new field open}
Once you select 

You can only make one password, but can make infinite usernames or security questions. If there is a security phrase or pin in addition to a password, that should be a security question field.


{image of /new with fields already added}
Once you have added fields, a new button appears that changes the focus to the list below to let you select a field to edit. The reason why I did this, using a button to navigate between the two sections, is because I believe that it is important to keep the mouse disabled here. (Look above at mouse usage and clipboard section) 



other various decisions:


#### in circulation
In each entry there is a boolean named `Circulate` which determines if the entry shows up in `/list` or `/pick`. All commands that work on entries still work (edit, open, copy, etc.). The reasoning for this is because when I change my password on a site I want to keep the old one. Making a new entry when you change the password while removing the old entry from circulation solves that problem. The entry will still show up in `/find` however it will 
