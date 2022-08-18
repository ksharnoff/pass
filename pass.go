/*
	/help
	write out what to write for /help in a google doc
	order commands list + helpinfo info alphabetically
	for commandsText in /open write that in order to edit you must go to /edit
	write about mouse usage and reason
	write it with string defined over several lines


	rename commandsText

	fix EnableMouse for the correct places

	fix SetDoneFunc for pick.list and the other lists!

	have it do something in case that file doesn't exsist, or instead have it start with that file downloaded?? 

	add another time to keep track of, of the last time opened? 

	error second text should add the extra space,, or be scooted over, so it will look good with auto generated strings

	rename pass.yaml?

	edit all the usages of pointers to slices when it is unnecessary


	!! don't have it marshalll/unmarhsall??? just encrypt the bytes lmaoooo?!?!

	remove /test from the list of functions on the left -- but keep it in??? 

	reorder commmands:
	/home
	/help
	--
	/new
	--
	/edit
	/copy
	--
	/find
	--
	/list
	/pick & picc
	--
	/open & copen

	maybe turn mouse off for /help?


	fix the colors so it is visable in flux, and it looks better loll

	send pull request for info
		tview.Pages is order dependant
		tview normal color is snow

*/

package main

import (
	"github.com/rivo/tview"
	"github.com/gdamore/tcell/v2"
	"strconv" // used to convert _ to string and string to _
	"fmt" 
	"strings"
	"github.com/atotto/clipboard" // copies the data to clipboard in /copen
	
	"time"

	// encryption thing
	"crypto/cipher"
	"pass/encrypt"

	// for writing to file
	"os"
	"gopkg.in/yaml.v3"
)

type entry struct {
	Name string
	Tags string // if search function works by looking at start of string, make tags an []string
	Usernames []Field
	Password Field
	SecurityQ []Field
	Notes [6]string // maybe make this an 8 in the future?
	Circulate bool
	Created time.Time
	Modified time.Time
}
type Field struct {
	DisplayName string
	Value string
}

type textGrid struct {
	text *tview.TextView
	grid *tview.Grid
}
type formGrid struct{
	form *tview.Form
	grid *tview.Grid
}
type listGrid struct{
	list *tview.List
	grid *tview.Grid
}
type textFormFlexGrid struct{
	text *tview.TextView
	form *tview.Form 
	flex *tview.Flex
	grid *tview.Grid
}
type inputGrid struct{
	input *tview.InputField
	grid *tview.Grid
}
type twoTextFlexGrid struct{
	title *tview.TextView 
	text *tview.TextView
	flex *tview.Flex
	grid *tview.Grid
}

func main(){
	app := tview.NewApplication()

	entries := []entry{}
	
	// pages is the pages set up for the left top box
	pages := tview.NewPages()

	// this is what everything is in, with it being split between the left and right
	// (the left being another flex split up and down)
	flex := tview.NewFlex()
	flexRow := tview.NewFlex().SetDirection(tview.FlexRow) // change name to flexLeft?

	// this is the text box that contains the commands, on the left and its grid (border)
	commands := textGrid{text: tview.NewTextView().SetScrollable(true), grid: tview.NewGrid().SetBorders(true)}
	homeCommands := " commands\n --------\n /home \n /help \n /new \n /find str\n /edit # \n /open # \n /copen # \n /list \n /pick \n /picc \n /copy \n /test"

	// this is the box that the page is set to when at /home
	// probably delete the title as some point, it's just like that for now tho
	sadEmptyBox := tview.NewBox().SetBorder(true).SetTitle("sad, empty box")

	// string of what is put into the command line
	inputed := ""
	// this is the commandLine as well as its grid (border)
	commandLine := inputGrid{input: tview.NewInputField().SetLabel("input: ").SetFieldWidth(60), grid: tview.NewGrid().SetBorders(true)}

	// this is the function that will do things based off the commands given to commandLine.input
	commandLineActions := func(key tcell.Key){}

	// this function is called when the focus switches back 
	// and one can type in the command line, so it says to look right 
	lookRightCommandLinePlaceholder := func(){
		commandLine.input.SetPlaceholder("psst look to the right")
	}
	// this function is called when the focus switches away and one
	// cannot type in the command line, so it says so 
	cantTypeCommandLinePlaceholder := func(){
		commandLine.input.SetPlaceholder("psst you can't type here right now")
	}

	// this is the text box that contains that list entry names and its grid (border)	
	list := twoTextFlexGrid{title: tview.NewTextView().SetWrap(false), text: tview.NewTextView().SetScrollable(true).SetWrap(false), grid: tview.NewGrid().SetBorders(true), flex: tview.NewFlex()}

	// this is a text box to print out the entire entries, to test!
	test := textGrid{text: tview.NewTextView().SetScrollable(true), grid: tview.NewGrid().SetBorders(true)}

	// this is the text box with the /help info and its grid (border)
	help := textGrid{text: tview.NewTextView().SetScrollable(true).SetText(" /help \n -----\n\n Do /new in order to put in a new entry. \n Do a;sdkfjkl  \n\n /find is case insensitve.  \n\n Do /open to view an entry. \n You will have to put in the password before you can see the information. \n Passwords and security questions will be blotted out, but they can be copied. (Or highlighted to see them) \n To delete an entry do /edit \n\n do /edit to edit fields or delete an entry. \n in /edit, all edits are permanently saved field by field as you click save \n\n the values of all fields, except the name of the entry and the tags, are equally encrypted. \n\n circulation? \n if you don't want an entry anymore, have no more use for it, but don't want to delete it, you should remove it from circulation. it will show up in /find results, but not from /list or /pick \n\n difference between /pick and /picc "), grid: tview.NewGrid().SetBorders(true)}

/*
	helpStrdksalfjk := `
hello 
HELLO 
	tab??
hdslkjfalk
	`
*/

	// text and grid for opening an entry already made, its function to format the information
	open := textGrid{text: tview.NewTextView().SetScrollable(true).SetDynamicColors(true), grid: tview.NewGrid().SetBorders(true)}
	blankOpen := func(i int) string {return "error, blankOpen(i int) didn't run"}

	copen := listGrid{list: tview.NewList().SetMainTextColor(tcell.GetColor("ColorSnow")).SetOffset(1, 1), grid: tview.NewGrid().SetBorders(true)}
	blankCopen := func(i int){}
	
	// when something happens that could give an error it will switch to here
	// and print it with what the erorr is 
	// split to two text boxes, top stays the same while the second changes
	error := twoTextFlexGrid{title: tview.NewTextView().SetText(" Uh oh! There was an error:"), text: tview.NewTextView().SetScrollable(true), flex: tview.NewFlex(), grid: tview.NewGrid().SetBorders(true)}

	switchToHome := func(){}

	switchToWriteFileErr := func() bool{return false}


	// the following variables are all for when you adding a new entry (and a new field)

	newEntry := textFormFlexGrid{form: tview.NewForm(), flex: tview.NewFlex(), grid: tview.NewGrid().SetBorders(true)}
	blankNewEntry := func(e entry){}

	blankFieldsAdded := func(){}
	newFieldsAddedList := tview.NewList().SetSelectedFocusOnly(true)
	switchToNewFieldsList := func(doSwitch bool){}

	// this is the form of adding a new field, its grid, 
	newField := formGrid{form: tview.NewForm(), grid: tview.NewGrid().SetBorders(true)}
	newFieldFlex := tview.NewFlex() // flex to situate it in the middle of page
	blankNewField := func(e *entry){} //function to set up the form, clear it each time
	newEditFieldFlex := tview.NewFlex() // flex to put it in the middle of page of when you edit one of the fields added already to /new

	// this is the form for adding a new notes, its grid, flex, and function
	newNote := formGrid{form: tview.NewForm(), grid: tview.NewGrid().SetBorders(true)}
	newNoteFlex := tview.NewFlex() // flex to put it in the middle of the page, nil on the sides
	blankNewNote := func(e *entry){}

	// these are all temporary, they are what a new entry or field is set to in
	// order to add it. to clear the following after/before use just set them 
	// equal to entry{} or field{}
	tempEntry := entry{}
	
	tempField := Field{}
	// fieldType keeps track of the type of the field, in order to add it to the correct
	// part of the entry.
	fieldType := ""

	// no matter if the formatting changes, it will be username as 0, password as 1, etc.
	// did not let me pass in [3] as input, it must be a slice
	dropDownFields := []string{"username", "password", "security question"}

	// text to be put on the left side when in /new
	// REPLACE select WITH button?????? MAYBE??? -- ask lucy :)
	newCommands := " /new \n ---- \n move: \n -tab \n -back tab \n\n select: \n -return \n\n must name \n entry to \n save it \n\n escape? \n quit"
	newFieldCommands := " /new \n ---- \n move: \n -tab \n -back tab \n\n select: \n -return \n\n must name \n field to \n save it \n\n escape? \n quit" //only change from this one to the newCommands is field vs. entry


	// the folowing variables are for editing an entry


	// these are the list, its grid, and the function to make the list when
	// /edit an entry
	edit := listGrid{list: tview.NewList().SetSelectedFocusOnly(true), grid: tview.NewGrid().SetBorders(true)}
	blankEditList := func(i int){}
	runeAlphabet := []rune{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z'}

	// if it's at the limit, the length of runeAlphabet, then it gets set back to 0, if it is not, then it plus pluses
	runeAlphabetIterate := func(i *int){}

	// this is the form and its grid and flex for editing a specific field
	// it has two functions, for editing one of the strings in the entry struct
	// and one for editing one of the fields (password, username, securityQ)
	editField := formGrid{form: tview.NewForm(), grid: tview.NewGrid().SetBorders(true)}
	editEditFieldFlex := tview.NewFlex() // flex to put it in the middle of the page, other items are nil
	editFieldStrFlex := tview.NewFlex() // flex to put the edit fields for tags and strings in middle of page as they have less buttons than the other thing
	blankEditFieldForm := func(f *Field, fieldArr *[]Field, index int, e *entry, pass, edit bool) {}
	blankEditStringForm := func (display, value string, e *entry){}

	// this is a function that solves redundancy in going back to /edit
	// (remaking the list, switching the page, setting the focus)
	// the function uses indexSelectEntry, which is that in that should be the 
	// index of the current entry
	switchToEditList := func(modified bool){}
	// this is a variable to represent what entry is being edited,
	// it is outside of any function so it can be used to call the function above,
	// to switch back to the /edit page
	indexSelectEntry := -1

	// the little popup to ask if you are sure when deleting the entry in /edit
	//flex of just the text and form
	editDelete := textFormFlexGrid{text: tview.NewTextView().SetTextAlign(tview.AlignCenter).SetText("delete entry? \nCANNOT BE UNDONE"), form: tview.NewForm(), flex: tview.NewFlex().SetDirection(tview.FlexRow), grid: tview.NewGrid().SetBorders(true)}
	editDeleteFlex := tview.NewFlex() // flex to set up the smaller flex in the center
	blankEditDeleteEntry := func(){}

	editCommands := " /edit \n ----- \n move: \n -tab \n -back tab \n -arrows keys \n\n select: \n -return \n -click" // similar to newCommands and newFieldCommands

	pick := listGrid{list: tview.NewList().SetSelectedFocusOnly(true).SetDoneFunc(switchToHome), grid: tview.NewGrid().SetBorders(true)}
	blankPickList := func(openCopen string){}


	// following variables for setting up the password screen


	var ciphBlock cipher.Block

	passActions := func(key tcell.Key){}

	passInputed := ""
	password := inputGrid{input: tview.NewInputField().SetLabel("password: ").SetFieldWidth(71).SetMaskCharacter('*'), grid: tview.NewGrid().SetBorders(true)}

	passFlex := tview.NewFlex()

	passBoxPages := tview.NewPages()

	passPages := tview.NewPages()

	passBox := tview.NewBox().SetBorder(true)

	passErr := twoTextFlexGrid{title: tview.NewTextView().SetWrap(false).SetText(" error in signing in"), text: tview.NewTextView().SetScrollable(true).SetWrap(false), grid: tview.NewGrid().SetBorders(true), flex: tview.NewFlex()}

	// ------------------------------------------------ //
	// all varaibles initialized :) function time!
	// ------------------------------------------------ //


	// right here we're going to set up the encryption key and stuff :)

	passActions = func(key tcell.Key){
		passInputed = password.input.GetText()
		passBoxPages.SwitchToPage("passBox")

		var keySuccess bool 
		var keyErr string

		ciphBlock, keySuccess, keyErr = encrypt.KeyGeneration(passInputed)

		if keyErr != ""{
			passBoxPages.SwitchToPage("passErr")
			passErr.text.SetText(keyErr)
		}else if !keySuccess{
			passBoxPages.SwitchToPage("passErr")
			passErr.text.SetText(" wrong password!")
		}else{

			readErr := readFromFile(&entries, ciphBlock)

			if readErr != ""{
				pages.SwitchToPage("err")
				error.text.SetText(readErr)
			}else{
				passPages.SwitchToPage("passManager")
				switchToHome()
			}
		}
		password.input.SetText("")
	}
	password.input.SetDoneFunc(passActions)


	// written out what commandLine input does with stuff
	commandLineActions = func(key tcell.Key){
		app.EnableMouse(true)
		switchToHome() // resets commands text and placeholder and stuff to make sure all is working

		inputed = commandLine.input.GetText() 
		inputedArr := strings.Split(inputed, " ") 

		// the following if/else statements check that the number inputed for /edit or /open
		// that there is a number, it is an int, and it corresponds to an entry 
		if (inputedArr[0] == "/open")||(inputedArr[0] == "/edit")||(inputedArr[0] == "/copen")||(inputedArr[0] == "/copy"){

			indexSelectEntry = -1 //  sets it here to remove any previous doings

			if len(inputedArr) < 2 { // so if there is no number written
				error.text.SetText(" To " + inputedArr[0][1:5] + " an entry you must write " + inputedArr[0] + " and then a number. \n With a space after " + inputedArr[0] + " \n Ex: \n\t" + inputedArr[0] + " 3")
				pages.SwitchToPage("err")
			}else{
				openEditInt, intErr := strconv.Atoi(inputedArr[1])
				if intErr != nil{ // if what passed in is not a number
					error.text.SetText(" Make sure to only use " + inputedArr[0] + " by writing a number! \n For an example do /help")
					pages.SwitchToPage("err")
				}else{
					if openEditInt >= len(entries){ // if the number passed in isn't an index
						error.text.SetText(" The number you entered does not correspond to an entry. \n Do /list to see the entries (and their numbers) that exist.")
						pages.SwitchToPage("err")
					}else{
						indexSelectEntry = openEditInt
					}
				}
			}
		}

		switch inputedArr[0] {
		case "/home":
			pages.SwitchToPage("/home")
		case "/list":
			listAllIndexes := []int{}
			for i := 0; i < len(entries); i++{
				listAllIndexes = append(listAllIndexes, i)
			}
			title, text := listEntries(entries, listAllIndexes, " /list \n -----", false)
			list.title.SetText(title)
			list.text.SetText(text).ScrollToBeginning()

			pages.SwitchToPage("/list")
		case "/find":
			if len(inputedArr) < 2 { 
				error.text.SetText(" To find entries you must write /find and then characters. \n With a space after /find. \n Ex: \n\t /find college")
				pages.SwitchToPage("err")
			}else{
				title, text := findEntries(entries, inputedArr[1])
				list.title.SetText(title)
				list.text.SetText(text).ScrollToBeginning()
				pages.SwitchToPage("/list")
			}
		case "/test":
			test.text.SetText(testAllFields(entries))
			pages.SwitchToPage("/test")
		case "/new":
			app.EnableMouse(false)
			commands.text.SetText(newCommands)
			tempEntry = entry{}
			blankNewEntry(tempEntry)
			app.SetFocus(newEntry.form)
			cantTypeCommandLinePlaceholder()
			pages.SwitchToPage("/newEntry")
		case "/help":
			pages.SwitchToPage("/help")
		case "/open":
			if indexSelectEntry > -1 {
				app.EnableMouse(false)
				open.text.SetText(blankOpen(indexSelectEntry)) // taking input, just to be safe smile -- can change that in future
				pages.SwitchToPage("/open")
			}
		case "/copen":
			if indexSelectEntry > -1{
				app.SetFocus(copen.list)
				app.EnableMouse(false)
				blankCopen(indexSelectEntry)
				pages.SwitchToPage("/copen")
			}
		case "/edit":
			if indexSelectEntry > -1 {
				app.EnableMouse(false)
				commands.text.SetText(editCommands)
				cantTypeCommandLinePlaceholder()
				switchToEditList(false)
			}
		case "/pick":
			blankPickList(inputedArr[0])
			app.SetFocus(pick.list)
			pages.SwitchToPage("/pick")
			cantTypeCommandLinePlaceholder()
		case "/copy":
			if indexSelectEntry > -1 {
				app.EnableMouse(false)
				//commands.text.SetText(newCommands)  //<--- figure out the commands to write here!
				blankNewEntry(entries[indexSelectEntry])
				app.SetFocus(newEntry.form)
				cantTypeCommandLinePlaceholder()
				pages.SwitchToPage("/newEntry")
			}
		case "/picc":
			blankPickList(inputedArr[0])
			app.SetFocus(pick.list)
			pages.SwitchToPage("/pick")
			cantTypeCommandLinePlaceholder()
		default:
			error.text.SetText(" That input doesn't match a command! \n Look to the right right to see the possible commands. \n Make sure to spell it correctly!")
			pages.SwitchToPage("err")
		}
		commandLine.input.SetText("")
	}
	// adds the function to commandLine.input so it is run when return is pressed  // move this??
	commandLine.input.SetDoneFunc(commandLineActions)

	// switch to home sets everything to rights again
	switchToHome = func(){
		pages.SwitchToPage("/home")
		app.SetFocus(commandLine.input)
		commands.text.SetText(homeCommands)
		lookRightCommandLinePlaceholder()
	}

	runeAlphabetIterate = func(i *int){
		if *i == len(runeAlphabet){ // why does this work? shouldn't this be until its one less????????????? i have this written from so long ago i think i'm working on this too late at night
			*i = 0
		}else{
			*i++
		}
	}

	// tries to write to file, if it fails then it returns false and switches to different page
	switchToWriteFileErr = func() bool{
		
		writeErr := writeToFile(entries, ciphBlock)
		if writeErr != ""{
			pages.SwitchToPage("err")
			error.text.SetText(writeErr)
			return false
		}
		return true
	}

	// ----
	// functions for making a new entry
	// ----

	// if pass in an entry then that is for copy, if making a brand new entry (in /new) then you pass in an tempEntry as an empty version of that
	blankNewEntry = func(e entry){
		newEntry.form.Clear(true)
		newFieldsAddedList.Clear()
		
		tempEntry = e
		switchToNewFieldsList(false)

		newEntry.form.
			AddInputField("name", tempEntry.Name, 40, nil, func(itemName string){
				tempEntry.Name = itemName
			}).
			AddInputField("tags", tempEntry.Tags, 40, nil, func(tagsInput string){
				tempEntry.Tags = tagsInput
			}).
			// this order of the buttons is on purpose and makes sense
			AddButton("new field", func(){
				commands.text.SetText(newFieldCommands)
				blankNewField(&tempEntry)
				pages.ShowPage("/newField")
				app.SetFocus(newField.form)
			}).
			
			// !!! make it so you can't hit save if there is no tempEntry.name
			AddButton("save entry", func(){
				if tempEntry.Name != ""{
					tempEntry.Circulate = true
					tempEntry.Created = time.Now()
					entries = append(entries, tempEntry)
					if switchToWriteFileErr(){ // if successfully wrote to file, then it switches to home, if not then it switches to error page
						switchToHome()
					}
				}
			}).
			AddButton("quit", func(){
				switchToHome()
			}). 
			AddButton("notes", func(){ 
				blankNewNote(nil) 
				// this (blankNewNote) can be deleted and written in the 
				// commandLineActions() cases section if one wants to be able to
				// hit quit of newNote but keep the info
				pages.ShowPage("/newNote")
				app.SetFocus(newNote.form)
			})
	}

	// takes in a pointer to tempEntry if in /new, takes in an entry if in /edit 
	blankNewField = func(e *entry){
		if (e.Password.DisplayName != "")&&(dropDownFields[1] == "password"){
			dropDownFields[1] = "overide written password"
		}
		edit := false

		// only adds on tags to edit it if there are no tags made already and if its in /edit
		if e != &tempEntry{
			if (e.Tags == "")&&(dropDownFields[2] != "tags"){
				dropDownFields = append(dropDownFields, "tags") // don't change the text of "tags" its used elsewhere	
			} else if len(dropDownFields) > 2{
				if (e.Tags == "")&&(dropDownFields[3] != "tags"){
					dropDownFields = append(dropDownFields, "tags")
				}
			}
			if (e.Tags != "")&&(len(dropDownFields) == 4){ // if there are tags in the entry, but also tags as an option in the dropdown it should be removed
				dropDownFields = dropDownFields[:3]
				// above should be equivilent to: slice = slice[:len(slice)-1]
			}
			if (e.Password.DisplayName != "")&&((dropDownFields[1] == "password")||(dropDownFields[1] == "overide written password")){

				copy(dropDownFields[1:], dropDownFields[1+1:])
				dropDownFields = dropDownFields[:len(dropDownFields)-1]  
			}

			edit = true
		}

		tempField = Field{}

		tempTags := "" 
		fieldType = ""
		newField.form.Clear(true)
		newField.form. 
			AddDropDown("new field:", dropDownFields, -1, func(chosenDrop string, index int){

				for newField.form.GetFormItemCount() > 1 { // removes other things put there already
					newField.form.RemoveFormItem(1)
				} 

				fieldType = chosenDrop
				if index > -1 {
					if chosenDrop != "tags" { // only if there aren't the fields already there (doesn't count buttons)
							
						if chosenDrop == "username"{
							tempField.DisplayName = "email" // in case it isn't edited, sets this as the default
							newField.form.AddInputField("display name", "email", 50, nil, func(display string){
								tempField.DisplayName = display
							})
						}else if (chosenDrop == "password")||(chosenDrop == "overide written password"){
							tempField.DisplayName = "password" // in case it isn't edited, sets this as the default
							newField.form.AddInputField("display name", "password", 20, nil, func(display string){
								tempField.DisplayName = display
							})
						}else if (chosenDrop == "security question"){
							newField.form.AddInputField("question", "", 50, nil, func(display string){
								tempField.DisplayName = display
							})
						}
						newField.form.AddInputField("value", "", 40, nil, func(value string){
							tempField.Value = value
						})
					}else {
						newField.form.AddInputField("tags", tempEntry.Tags, 40, nil, func(tags string){
							tempTags = tags
						})
					}
				}
			}).
			AddButton("save field", func(){
				if (tempField.DisplayName != "")||(tempTags != ""){ 
					switch fieldType{
					case "username":
						e.Usernames = append(e.Usernames, tempField)
					case "password":
						e.Password = tempField
					case "overide written password":
						e.Password = tempField
					case "security question":
						e.SecurityQ = append(e.SecurityQ, tempField)
					case "tags":
						e.Tags = tempTags
					}
					if !edit{
						blankFieldsAdded()
						commands.text.SetText(newCommands)
						pages.SwitchToPage("/newEntry")
						app.SetFocus(newEntry.form)
					}else{
						switchToEditList(true)
					}
				}
			}).
			AddButton("quit", func(){
				if !edit{
					commands.text.SetText(newCommands)
					pages.SwitchToPage("/newEntry")
					app.SetFocus(newEntry.form)
				}else{
					switchToEditList(false)
				}
			})
	}

	// takes in pointer to entry so it can be used in /edit of a notes. if called in /new then it would take in nil and all would be fine
	blankNewNote = func(e *entry){
		newNote.form.Clear(true)
		toAdd := [6]string{}

		if e == nil{
			toAdd = tempEntry.Notes //maybe change the size of this in the future? current size is 6 -- or rewrite as a slice
		}else{
			toAdd = e.Notes
		}

		newNote.form.
			AddInputField("notes:", toAdd[0], 0, nil, func(inputed string){
				toAdd[0] = inputed
			})

		// i := i because making a new function in a closure (for loop) it
		// has i equal to the last iteration of it (would be 6)
		for i := 1; i < 6; i++ {
			i := i
			newNote.form.AddInputField("", toAdd[i], 0, nil, func(inputed string){
				toAdd[i] = inputed
			})
		}
		
		newNote.form.
			AddButton("save", func(){
				if e == nil{ // if all this is being done in /new
					tempEntry.Notes = toAdd
					pages.SwitchToPage("/newEntry")
					app.SetFocus(newEntry.form)
				}else{ // if all this is being done in /edit
					e.Notes = toAdd
					switchToEditList(true)
				}
			}). 
			AddButton("quit", func(){
				if e == nil{ // if being done in /new
					pages.SwitchToPage("/newEntry")
					app.SetFocus(newEntry.form)
				}else{ // if being done in /edit
					switchToEditList(false)
				}
			}). 
			AddButton("delete", func(){
				if e == nil{
					tempEntry.Notes = [6]string{}
					pages.SwitchToPage("/newEntry")
					app.SetFocus(newEntry.form)
				}else{
					e.Notes = [6]string{} // assigns the whole array at once :)
					switchToEditList(true)
				}
			})
	}

	blankFieldsAdded = func(){ 
		newFieldsAddedList.Clear()
		num := 0 // iterates through the rune alphabet slice to have set up all the rune shortcuts

		if newEntry.form.GetButtonIndex("edit field") < 0{
			newEntry.form. 
				AddButton("edit field", func(){ // DON'T CHANGE LABEL NAME
					app.SetFocus(newFieldsAddedList)
			})
		}

		newFieldsAddedList.
			AddItem("move back to top", "", runeAlphabet[num], func(){
				app.SetFocus(newEntry.form)
			})

		for i := range tempEntry.Usernames {
			i := i
			u := &tempEntry.Usernames[i]
			runeAlphabetIterate(&num)

			newFieldsAddedList.AddItem(u.DisplayName + ":", u.Value, runeAlphabet[num], func(){
				blankEditFieldForm(u, &tempEntry.Usernames, i, &tempEntry, false, false)
				pages.ShowPage("/new-editField") 
				app.SetFocus(editField.form)
			})
		}
		if tempEntry.Password.DisplayName != "" {
			runeAlphabetIterate(&num)
			newFieldsAddedList.AddItem(tempEntry.Password.DisplayName + ":", "[black]" + tempEntry.Password.Value, runeAlphabet[num], func(){
				blankEditFieldForm(&tempEntry.Password, nil, -1, &tempEntry, true, false)
				pages.ShowPage("/new-editField") 
				app.SetFocus(editField.form)
			})
		}
		for i := range tempEntry.SecurityQ {
			i := i
			sq := &tempEntry.SecurityQ[i]
			runeAlphabetIterate(&num)

			newFieldsAddedList.AddItem(sq.DisplayName + ":", "[black]" + sq.Value, runeAlphabet[num], func(){
				blankEditFieldForm(sq, &tempEntry.SecurityQ, i, &tempEntry, false, false)
				pages.ShowPage("/new-editField") 
				app.SetFocus(editField.form)
			})
		}
	}

	switchToNewFieldsList = func(doSwitch bool){
		blankFieldsAdded()
		if (doSwitch) && (newFieldsAddedList.GetItemCount() > 1) {
			pages.SwitchToPage("/newEntry")
			app.SetFocus(newFieldsAddedList)
		}
		if newFieldsAddedList.GetItemCount() < 2{ // if all the fields are deleted, then:
			newFieldsAddedList.Clear()

			editFieldIndex := newEntry.form.GetButtonIndex("edit field")
			if editFieldIndex > -1 {
				newEntry.form.RemoveButton(editFieldIndex)
				pages.SwitchToPage("/newEntry")
				app.SetFocus(newEntry.form)
			}else{
				error.text.SetText("AHHHHHHH for some reason the edit field button wasn't added despite a field later trying to be deleted!!!!")
				pages.SwitchToPage("err")
			}
		}
	}

	// ----
	// function for displaying an entry -- move outside main?
	// ----

	// precondition: i > -1
	// since it's returning a strin gmaybe have it be outside func main?
	blankOpen = func(i int) string{
		e := entries[i]
		print := " "

		print += "[" + strconv.Itoa(i) + "] " + e.Name + "\n " 
		print += strings.Repeat("-", len([]rune(print))-3) + " \n" // right now it matches under the letters of title, if at -2 then it goes one out
		if e.Tags != ""{
			print += " tags: " + e.Tags + "\n"
		}
		for _, u := range e.Usernames {
			print += " " + u.DisplayName + ": " + u.Value + "[snow]\n"
		}
		if e.Password.DisplayName != "" {
			print += " " + e.Password.DisplayName + ": [black]" + e.Password.Value + "[snow]\n"
		}
		for _, sq := range e.SecurityQ {
			print += " " + sq.DisplayName + ": [black]" +sq.Value + "[snow]\n"
		}
		emptyNotes := true

		for _, n := range e.Notes{
			if n != ""{
				emptyNotes = false
				break
			}
		}
		if !emptyNotes{
			blankLines := 0

			print += " notes: " 
			for _, n := range e.Notes {
				if n == ""{
					blankLines++
				}else{
					print += strings.Repeat("\n", blankLines)
					print += "\n\t " + n
					blankLines = 0
				}
			}
		}
		print += "\n\n"

		// following is all the "metadata"
		print += " in circulation: " + strconv.FormatBool(e.Circulate) + "\n"

		if !e.Modified.IsZero(){ // if it's not jan 1, year 1
			print += " date last modified: " + fmt.Sprint(e.Modified.Date()) + "\n" // CHANGE THIS TO HAVE A COMMA AND BE BETTER
		}
		if !e.Created.IsZero(){ // if it's not jan 1, year 1
			print += " date created: " + fmt.Sprint(e.Created.Date()) // CHANGE THIS TO HAVE A COMMA AND BE BETTER
		}
		return print
	}

	blankCopen = func(i int){
		num := 0
		copen.list.Clear()
		e := entries[i]

		//print += "[" + strconv.Itoa(i) + "] " + e.name + "\n " 
		copen.list.AddItem("leave /copen", "(takes you back to /home)", runeAlphabet[num], func(){
				switchToHome()
			})
		runeAlphabetIterate(&num)
		copen.list.AddItem("name:", e.Name, runeAlphabet[num], func(){
			clipboard.WriteAll(e.Name)
		})
		if e.Tags != ""{
			runeAlphabetIterate(&num)
			copen.list.AddItem("tags:", e.Tags, runeAlphabet[num], func(){
				clipboard.WriteAll(e.Tags)
			})
		}
		for _, u := range e.Usernames{
			u := u
			runeAlphabetIterate(&num)
			copen.list.AddItem(u.DisplayName + ":", u.Value, runeAlphabet[num], func(){
				clipboard.WriteAll(u.Value)
			})
		}
		if e.Password.DisplayName != ""{
			runeAlphabetIterate(&num)
			copen.list.AddItem(e.Password.DisplayName + ":", "[black]" + e.Password.Value,  runeAlphabet[num], func(){
				clipboard.WriteAll(e.Password.Value)
			})
		}
		for _, sq := range e.SecurityQ{
			sq := sq 
			runeAlphabetIterate(&num)
			copen.list.AddItem(sq.DisplayName + ":", "[black]" + sq.Value, runeAlphabet[num], func(){
				clipboard.WriteAll(sq.Value)
			})
		}
		for _, n := range e.Notes{ 
			n := n
			if n != ""{
				runeAlphabetIterate(&num)
				copen.list.AddItem("note:", n, runeAlphabet[num], func(){
					clipboard.WriteAll(n)
				})
			}
		}
		runeAlphabetIterate(&num)
		copen.list.AddItem("in circulation:", strconv.FormatBool(e.Circulate), runeAlphabet[num], func(){
			clipboard.WriteAll(strconv.FormatBool(e.Circulate))
		})
		if !e.Created.IsZero(){
			runeAlphabetIterate(&num)
			copen.list.AddItem("date created:", fmt.Sprint(e.Created.Date()), runeAlphabet[num], func(){
				clipboard.WriteAll(fmt.Sprint(e.Created.Date()))
			})
		}
		if !e.Modified.IsZero(){
			runeAlphabetIterate(&num)
			copen.list.AddItem("date last modifed:", fmt.Sprint(e.Modified.Date()), runeAlphabet[num], func(){
				clipboard.WriteAll(fmt.Sprint(e.Modified.Date()))
			})
		}
	}

	// ----
	// functions for editing an entry
	// ----

	blankEditList = func(i int){
		edit.list.Clear()
		e := &entries[i]

		num := 0
		edit.list.AddItem("leave /edit", "(takes you back to /home)", runeAlphabet[num], func(){
			switchToHome()
		})
		runeAlphabetIterate(&num)
		edit.list.AddItem("name: ", e.Name, runeAlphabet[num], func(){
			blankEditStringForm("name", e.Name, e)
			pages.ShowPage("/editFieldStr") 
			app.SetFocus(editField.form)
		})
		if e.Tags != "" {
			runeAlphabetIterate(&num)
			edit.list.AddItem("tags:", e.Tags, runeAlphabet[num], func(){
				blankEditStringForm("tags", e.Tags, e)
				pages.ShowPage("/editFieldStr") 
				app.SetFocus(editField.form)
			})
		}
		for i := range e.Usernames {
			i := i
			u := &e.Usernames[i]
			runeAlphabetIterate(&num)
			edit.list.AddItem(u.DisplayName + ":", u.Value, runeAlphabet[num], func(){
				blankEditFieldForm(u, &e.Usernames, i, e, false, true)
				pages.ShowPage("/edit-editField") 
				app.SetFocus(editField.form)
			})
		}
		if e.Password.DisplayName != "" {
			runeAlphabetIterate(&num)
			edit.list.AddItem(e.Password.DisplayName + ":", "[black]" + e.Password.Value, runeAlphabet[num], func(){
				blankEditFieldForm(&e.Password, nil, -1, e, true, true)
				pages.ShowPage("/edit-editField") 
				app.SetFocus(editField.form)
			})
		}
		for i := range e.SecurityQ {
			i := i
			sq := &e.SecurityQ[i]
			runeAlphabetIterate(&num)

			edit.list.AddItem(sq.DisplayName + ":", "[black]" + sq.Value, runeAlphabet[num], func(){
				blankEditFieldForm(sq, &e.SecurityQ, i, e, false, true)
				pages.ShowPage("/edit-editField") 
				app.SetFocus(editField.form)
			})
		}
		condensedNotes := ""
		emptyNotes := true
		for _, n := range e.Notes{
			n := n 
			condensedNotes += n + ", "
			if n != ""{
				emptyNotes = false
			}
		}
		if !emptyNotes {
			runeAlphabetIterate(&num)
			edit.list.AddItem("notes:", condensedNotes, runeAlphabet[num], func(){
				blankNewNote(e)
				pages.ShowPage("/newNote") 
				app.SetFocus(newNote.form)
			})
		}

		newFieldStr := "usernames"

		if e.Password.DisplayName == ""{
			newFieldStr +=", password"
		}
		if e.Tags == ""{
			newFieldStr += ", tags"
		}

		runeAlphabetIterate(&num)
		edit.list.AddItem("add new field", newFieldStr + ", security questions", runeAlphabet[num], func(){
			// code copied from blankNewEntry
			commands.text.SetText(newFieldCommands)
			blankNewField(e)
			pages.ShowPage("/newField")
			app.SetFocus(newField.form)
		})

		runeAlphabetIterate(&num)
		if e.Circulate{ // if it is in circulation, option to opt out
			edit.list.AddItem("remove from circulation", "(not permanant), check /help for info", runeAlphabet[num], func(){
				e.Circulate = false
				switchToEditList(true)
			})

		}else{ // if it has been removed, option to opt back in 
			edit.list.AddItem("add back to circulation", "(not permanant), check /help for info", runeAlphabet[num], func(){
				e.Circulate = true
				switchToEditList(true)
			})
		}
		runeAlphabetIterate(&num)
		edit.list.AddItem("delete entry", "(permanant!!)", runeAlphabet[num], func(){
			blankEditDeleteEntry()
			pages.ShowPage("/editDelete")
			app.SetFocus(editDelete.form)
		})
	}

	// includes a boolean if true it is the password field
	// can pass in nil for the slice, -1 for index if it is the password field

	// take in an extra bool for it to be edit/new field form, to check where to send back to!!!
	blankEditFieldForm = func(f *Field, fieldArr *[]Field, index int, e *entry, pass, edit bool){
		editField.form.Clear(true)

		tempField = Field{} // not necessary?,, bc being set on next few lines?
		tempField.DisplayName = f.DisplayName
		tempField.Value = f.Value

		editField.form.
			AddInputField("display name:", tempField.DisplayName, 40, nil, func(input string){
				tempField.DisplayName = input
			}).
			AddInputField("value:", tempField.Value, 40, nil, func(input string){
				tempField.Value = input
			}). 
			AddButton("save", func(){
				*f = tempField
				if edit {
					switchToEditList(true)
				}else{
					switchToNewFieldsList(true)
				}
			}). 
			AddButton("quit", func(){
				if edit {
					switchToEditList(false)
				}else{
					switchToNewFieldsList(true)
				}
			}).
			AddButton("delete field", func(){
				if pass {
					e.Password = Field{}
					if edit {
						switchToEditList(true)
					}else{
						dropDownFields[1] = "password"
						switchToNewFieldsList(true)
					}
				}else{
					if (fieldArr != nil)&&(index != -1){
						// way this is going to be coded to delete it will change the
						// order of the splice, rewrite this to be in order (SLOWER)
						// if want this to stay in order (maybe)
						(*fieldArr)[index] = (*fieldArr)[len(*fieldArr)-1]
						(*fieldArr) = (*fieldArr)[:len(*fieldArr)-1]
						if edit {
							switchToEditList(true)
						}else{
							switchToNewFieldsList(true)
						}
					}else{
						error.text.SetText("AHHHHH the array given to blankEditFieldForm is nil and it shouldnt be!!!!")
						pages.SwitchToPage("err")
					}
				}
			})
	}

	// can either be name or tags
	blankEditStringForm = func (display, value string, e *entry){
		if (display != "name")&&(display != "tags"){
			error.text.SetText("AHHHH the input of display should only be tags or name!!")
			pages.SwitchToPage("err")
		}else{

			editField.form.Clear(true)
			
			tempDisplay := display 
			tempValue := value

			editField.form.
				AddInputField(tempDisplay, tempValue, 40, nil, func(changed string){
					tempValue = changed
				}). 
				AddButton("save", func(){
					if display == "name"{
						e.Name = tempValue
					}else{
						e.Tags = tempValue
					}
					switchToEditList(true)
				}). 
				AddButton("quit", func(){
					switchToEditList(false)
				})

			// does not make a delete button for editing the name, as each entry must have a name, therefore only option would be to delete tags
			if display != "name"{ 
				editField.form.AddButton("delete", func(){
					e.Tags = ""
					switchToEditList(true)
				})
			}
		}	
	}

	// maybe just have it have input as indexSelectEntry and move it outside func main? 
	switchToEditList = func(modified bool){
		if switchToWriteFileErr(){
			if modified{
				entries[indexSelectEntry].Modified = time.Now()
			}

			blankEditList(indexSelectEntry)
			pages.SwitchToPage("/edit")
			app.SetFocus(edit.list)
		}
	}

	blankEditDeleteEntry = func(){
		editDelete.form.Clear(true)
		editDelete.form.SetButtonsAlign(tview.AlignCenter)
		editDelete.form.
			AddButton("save", func(){
				switchToEditList(false)
			}).
			AddButton("delete", func(){ // this deletes it, slower version, keeps everything in order
				copy(entries[indexSelectEntry:], entries[indexSelectEntry+1:])
				entries[len(entries)-1] = entry{} // ask dada why this is here?
				entries = entries[:len(entries)-1]	
				if switchToWriteFileErr(){			
					switchToHome()
				}
			})
	}

	// ----
	// /pick and /picc
	// ----

	// function for making the list in /pick and /picc
	// if openCopen is true, then when you select an entry it takes you do /open of that entry, vs. if it is false it would take you to /copen
	blankPickList = func(openCopen string){
		// have press escape mean that you leave
		num := 0
 		pick.list.Clear()

 		pick.list.AddItem("leave " + openCopen, "(takes you back to /home", runeAlphabet[num], func(){
 			switchToHome()
 		})

 		for i, e := range entries{
 			num++
 			if num == len(runeAlphabet){
 				num = 0
 			}
 			i := i

			if (e.Circulate){ // have it include the index number as [0] maybe??? would be cool!
	    		pick.list.AddItem("name: " + e.Name, "tags: " + e.Tags, runeAlphabet[num], func(){
	    			if openCopen == "/pick"{ // to transfer to /open #
		    			// following code copied from commandLineActions function
		    			open.text.SetText(blankOpen(i)) // taking input, just to be safe smile -- can change that in future
						pages.SwitchToPage("/open")
						app.SetFocus(commandLine.input)
						lookRightCommandLinePlaceholder()
					}else{ // to transfer to /copen #
						// following code copied from commandLineActions function
						app.SetFocus(copen.list)
						app.EnableMouse(false)
						blankCopen(i)
						pages.SwitchToPage("/copen")
					}
    			})
	    	}
		}
	}

	// ------------------------------------------------ //
	// setting up the grids, pages, flexes :)
	// ------------------------------------------------ //

	passErr.flex.SetDirection(tview.FlexRow). 
		AddItem(passErr.title, 0, 1, false). 
		AddItem(passErr.text, 0, 8, false)

	passFlex. 
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow). 
			AddItem(passBoxPages, 0, 9, false). 
			AddItem(password.grid, 0, 1, false), 0, 1, false)

	newFieldFlex.
		AddItem(nil, 0, 1, false). 
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow). 
			AddItem(nil, 0, 2, false). 
			AddItem(newField.grid, 0, 3, false). 
			AddItem(nil, 0, 1, false), 0, 4, false)

	error.flex.SetDirection(tview.FlexRow). 
		AddItem(error.title, 0, 1, false).
		AddItem(error.text, 0, 8, false)

	newNoteFlex. 
		AddItem(nil, 0, 1, false). 
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow). 
			AddItem(nil, 0, 2, false). 
			// following two, 5 is the max for changing
			AddItem(newNote.grid, 0, 6, false). // 4 fits 3 input + buttons,,5 fits 4 input + buttons
			AddItem(nil, 0, 1, false), 0, 5, false) 

	newEntry.flex.SetDirection(tview.FlexRow).
		AddItem(newEntry.form, 0, 1, false). 
		AddItem(newFieldsAddedList, 0, 2, false) // 1:2 is the maximum  

	newEditFieldFlex. // for /new
		AddItem(nil, 0, 1, false). 
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow). 
			AddItem(nil, 0, 3, false). 
			AddItem(editField.grid, 0, 3, false). 
			AddItem(nil, 0, 2, false), 0, 4, false)

	editEditFieldFlex. // for /edit
		AddItem(nil, 0, 1, false). 
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow). 
			AddItem(nil, 0, 2, false). 
			AddItem(editField.grid, 0, 3, false). 
			AddItem(nil, 0, 3, false), 0, 4, false)

	editFieldStrFlex.
		AddItem(nil, 0, 1, false). 
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow). 
			AddItem(nil, 0, 3, false). 
			AddItem(editField.grid, 0, 2, false). 
			AddItem(nil, 0, 2, false), 0, 4, false)

	editDelete.flex.
		AddItem(editDelete.text, 0, 1, false). 
		AddItem(editDelete.form, 0, 1, false)

	editDeleteFlex. 
		AddItem(nil, 0, 1, false). 
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow). 
			AddItem(nil, 0, 2, false). // 2 
			AddItem(editDelete.grid, 0, 2, false). 
			AddItem(nil, 0, 2, false), 0, 1, false).  //3
		AddItem(nil, 0, 1, false)

	list.flex.SetDirection(tview.FlexRow). 
		AddItem(list.title, 0, 1, false). 
		AddItem(list.text, 0, 8, false)

	// uses a function to add each thing to its perspective grid
	grider(commandLine.input, commandLine.grid)
	grider(commands.text, commands.grid)
	grider(list.flex, list.grid)
	grider(test.text, test.grid)
	grider(newEntry.flex, newEntry.grid) // update this one!
	grider(newField.form, newField.grid)
	grider(newNote.form, newNote.grid)
	grider(help.text, help.grid)
	grider(error.flex, error.grid)
	grider(open.text, open.grid)
	grider(edit.list, edit.grid)
	grider(editField.form, editField.grid)
	grider(editDelete.flex, editDelete.grid)
	grider(pick.list, pick.grid)
	grider(copen.list, copen.grid)
	grider(password.input, password.grid)
	grider(passErr.flex, passErr.grid)

	// all the different pages are added here
	pages.
		AddPage("/home", sadEmptyBox, true, true). 
		AddPage("/list", list.grid, true, false). 
		AddPage("/test", test.grid, true, false). 
		AddPage("/edit", edit.grid, true, false). 
		AddPage("/help", help.grid, true, false). 
		AddPage("err", error.grid, true, false). 
		AddPage("/open", open.grid, true, false). 
		AddPage("/pick", pick.grid, true, false). 
		AddPage("/copen", copen.grid, true, false). 
		AddPage("/newEntry", newEntry.grid, true, false).
		AddPage("/newField", newFieldFlex, true, false). 
		AddPage("/newNote", newNoteFlex, true, false). 
		AddPage("/new-editField", newEditFieldFlex, true, false). 
		AddPage("/editFieldStr", editFieldStrFlex, true, false).
		AddPage("/editDelete", editDeleteFlex, true, false). 
		AddPage("/edit-editField", editEditFieldFlex, true, false)

	// sets up the flex row of the left side, top is the pages bottom is the commandLine.input
	// ratio of 8:1 is the maximum that it can be (9:1 and 100:1 are the same as 8:1)
	// ratio of 9:1 is good on 28x84 grid
	flexRow. 
		AddItem(pages, 0, 9, false). 
		AddItem(commandLine.grid, 0, 1, false)

	// the greater flex consisting of the left and right sides
	// 5:1 is original, 6:1 gets two more columns for realList, 14:3 for no extra columns in realList :D:D:D 
	flex. 
		AddItem(flexRow, 0, 14, false). 
		AddItem(commands.grid, 0, 3, false) 

	passBoxPages.
		AddPage("passBox", passBox, true, true). 
		AddPage("passErr", passErr.grid, true, false)

	passPages. 
		AddPage("passInput", passFlex, true, true). 
		AddPage("passManager", flex, true, false)


	// if EnableMouse is false, then can copy/paste
	// have enable mouse turn on when in /edit, /pick !!
	if err := app.SetRoot(passPages, true).SetFocus(password.input).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

// this is the function used to put any type of primitive
// into a grid, as in using the grid to make a border
func grider(prim tview.Primitive, grid *tview.Grid){
	grid.AddItem(prim, 0, 0, 1, 1, 0, 0, false)
}

// this finds all the entries that has a certain str in its name or tags
func findEntries(entries []entry, str string) (string,string){
	indexes := []int{}
	str = strings.ToLower(str)
	for i, e := range entries {
		if (strings.Contains(strings.ToLower(e.Name), str))||(strings.Contains(strings.ToLower(e.Tags), str)) {
			indexes = append(indexes, i)
		}
	}

	// trims out the thing that prints so it won't go all the way to the other side
	// still used the full string as the search before this point though
	// MAYBE MAKE IT A LITTLE SHORTER AND ADD A "..." AT THE END? 
	trimmedStr := str
	if len([]byte(trimmedStr)) > 59{
		trimmedStr = trimmedStr[:59]

	}

	if len(indexes) > 0{
		return listEntries(entries, indexes, " /find " + trimmedStr + " \n " + strings.Repeat("-", len([]rune(trimmedStr))+6), true)
	}else{ 
		return " /find " + trimmedStr + " \n " + strings.Repeat("-", len([]rune(trimmedStr))+6), "no entries found"
	}
}

// this is the function that formats each entry all together " [0] twitter", to be printed out in /find or in /list
// the str taken in will be: " /find str \n-----" or " /list \n -----"
// bool taken in differentiates from /list or /find, to show or not show the ones that are not in circulation. If not in circulation, but is found in /find, it puts (rem) on the title if it is not in circulation
func listEntries (entries []entry, indexes []int, str string, showOld bool) (string,string){

	printStr := ""
	printEntries := []entry{}

	if showOld{
		for _,i := range indexes{
			printEntries = append(printEntries, entries[i]) // equivilent to entries[i] is entries[indexes[i]]
		}
	}else{
		indexes = nil
		for i,e := range entries{
			if e.Circulate{
				printEntries = append(printEntries, e)
				indexes = append(indexes, i)
			}
		}
	}
	third := len(indexes)/3
	if third < 21 {
		third = 21
	}

	for i := 0; i < third; i++{
		if i >= len(indexes){
			break
		}
		printStr += " " + indexName(indexes[i], entries)
		if len(indexes) > i+third{
			printStr += indexName(indexes[i+third], entries)
		}
		if len(indexes) > i+third+third{
			printStr += indexName(indexes[i+third], entries)
		}
		if i != third-1 { //so it doesn't do it on the last one
			printStr += "\n"
		}
	}
	return str, printStr // first string is title, second is the body of the text
}
// this puts a single index from entries to " [0] twitterDEMO       ", with those exact spaces as they are what makes it into a good column shape
func indexName (index int, entries []entry) string{
	str := "[" + strconv.Itoa(index) + "] " 
	
	if !entries[index].Circulate{
		str += "(rem) "
	}

	str += entries[index].Name

	len := len([]rune(str))
	if len > 21{
		str = str[0:21] // second number in not inclusive
		str += " "
	}else{
		str += strings.Repeat(" ", 22-len)
	}
	return str
}

// call this to be printed out to see all the things in the struct - to test!
// maybe turn on text wrapping for the text box with this?
func testAllFields(entries []entry) string{
	allValues := "Test of all fields that are known:"

	for _, e := range entries {
		allValues += "\n\n " + fmt.Sprint(e)
	}
	return allValues
}


// writes to the pass.yaml file, if it fails then it returns a string with errors
func writeToFile(entries []entry, ciphBlock cipher.Block) string{
	output, marshErr := yaml.Marshal(entries)
	if marshErr != nil{
		return "error in yaml.marshal the entries \n" + marshErr.Error()
	}else{

		encryptedOutput := encrypt.Encrypt(output, ciphBlock, false)

		// conventions of writing to a temp file is write to .tmp
		writeErr := os.WriteFile("pass.yaml.tmp", encryptedOutput, 0600) // 0600 is the permissions, that only this user can read/write/excet to this file
		os.Rename("pass.yaml.tmp", "pass.yaml") // only will do this if the previous thing worked correctly, helps to save the data :)

		if writeErr != nil{
			return "error in os.writeFile \n" + writeErr.Error()
		}else{
			return ""
		}
	}
}

// if it works then it should return "", if not then it will return the errors in a string format
func readFromFile(entries *[]entry, ciphBlock cipher.Block) string{
	input, inputErr := os.ReadFile("pass.yaml")
	if inputErr != nil{
		return " error in os.ReadFile \n" + inputErr.Error()
	}else{
		// first we decrypt it!
		decryptedInput := encrypt.Decrypt(input, ciphBlock)

		unmarshErr := yaml.Unmarshal(decryptedInput, &entries)
		if unmarshErr != nil{
			return " error in yaml.Unmarshal \n" + unmarshErr.Error()
		}else{
			return ""		
		}
	}
}
