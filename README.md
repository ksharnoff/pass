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

#### encryption
- all of the information, except for the name of the entry, its tags, and if its in circulation are encrypted. 
  - the 'displayName' of the username, password, and security question fields are also encrypted.
  - there is no varaible for the display name for (entry) name, tags, notes, or if its in circulation, so that is not encrypted
- don't use on a cloud hosted computer!! or a shared computer!! 
  - some of the sensitive data is briefly stored in the memory 
- 

#### file writing
- it writes to a yaml file

#### how it looks
- i coded it for a 84x28 size window with text font of monaco, size 18. it will work at all sizes and with all texts (to my knowledge), however you may not be able to see all the buttons without scrolling or pressing tab. everything should still work, you should still be able to access everything, it just may not look all organized. 


