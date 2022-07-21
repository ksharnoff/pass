# pass

this is just a list to be edited down later to what to write


#### tview
- used grids to put boxes around text, lists, forms, flexes
- .


#### tui interface, decisions made
- there is a command line at the bottom that all inputs and made
- the list of possible commands are written in a box to the right, do /help for more details
- copy /help here into this readme?
- a lot of the functions are anonymous functions inside of func main, that is because a lot of the functions are just setting up the different Primitives (forms, lists) of tview
- this was coded for a terminal size of 78x26, as that is what looked best on my computer screen. Some of the spacing, of lines of text and of boxes of things, will be off if it's bigger or smaller but it will all still function. If it is smaller than 78x26 then you may have to scroll to see everything

#### encryption
- all of the information, except for the name of the entry, its tags, and if its in circulation are encrypted. 
  - the 'displayName' of the username, password, and security question fields are also encrypted.
  - there is no varaible for the display name for (entry) name, tags, notes, or if its in circulation, so that is not encrypted
- don't use on a cloud hosted computer!! or a shared computer!! 
  - some of the sensitive data is briefly stored in the memory 
- 

#### file writing
- it writes to 
