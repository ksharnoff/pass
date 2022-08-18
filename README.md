# pass

WORK OF PROGRESS, NOT FINISHED

this is just a list to be edited down later to what to write


#### tview
- used grids to put boxes around text, lists, forms, flexes
- .


#### tui interface, decisions made
- there is a command line at the bottom that all inputs and made
- the list of possible commands are written in a box to the right, do /help for more details
- copy /help here into this readme?
- a lot of the functions are anonymous functions inside of func main, that is because a lot of the functions are just setting up the different Primitives (forms, lists) of tview

#### encryption
- all of the information, except for the name of the entry, its tags, and if its in circulation are encrypted. 
  - the 'displayName' of the username, password, and security question fields are also encrypted.
  - there is no varaible for the display name for (entry) name, tags, notes, or if its in circulation, so that is not encrypted
- don't use on a cloud hosted computer!! or a shared computer!! 
  - some of the sensitive data is briefly stored in the memory 
- 


## functions
here are the different functions and commands that can be used

### /new
/new is how you can make a new entry 

#### /help
/help has condensed information from this readme, made easier so you don't have to leave

##### /list
/list lists all of the entries made so you can see the options to open them

###### /pick
/pick is like /list but it uses tview.List and lets you chose one,



