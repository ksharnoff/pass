/*
change from 26x78 to 31x91? 
^^^^^^^^^^

order commands list + helpinfo info alphabetically

for /new should have a limit of how many new fields you can make
// maybe have infinite number of notes that can be made????


make a better way to edit the notes? maybe change so notes field is actually
an array of strings?
^^^ make so if you press the new notes button again in the same time of making a new entry that it works well // deletes it all or doesnt'?


idea to implement:
a way to remove entries from the main index but then have them be in other array that keeps them, would show up in /find as (OLD) but not in /list

!!what to write in commands box for (new)notes???

also make the three column showing the /find

move text, grid, flex to structs!!! -- cleaner!

rename commandsText
rename grider, inputer
*/

package main

import (
	"github.com/rivo/tview"
	"github.com/gdamore/tcell/v2"
	"strconv" //used to convert from int to string for the index in nameToString()
	"fmt" //used to convert struct to string for testing functions
	"strings"
)

type entry struct {
	name string
	tags string // if search function works by looking at start of string, make tags an []string
	usernames []field
	password field
	securityQ []field
	notes string
}
type field struct {
	displayName string
	value string
	secret bool
}

/*
type HAHAHAHAH struc {
	textbox tview.TextView
}
*/


func main(){

	app := tview.NewApplication()

	// this is the list of the demo entries, real entries will be in a file
	entries := []entry{
		entry{
			name: "twitterDEMO",
			tags: "socials, demo",
			usernames: []field{
				{
					displayName: "username",
					value: "yellowyaks",
					secret: false,
				},
				{
					displayName: "email",
					value: "a;ksdjfkad@gmail.com",
					secret: false,
				},
			},
			password: field{
				displayName: "password",
				value: "0349",
				secret: true,
			},
			notes: "last changed password summer 2021\n\thii \n\thello \n\theyy",
		},
		entry{
			name: "college boardDEMO",
			tags: "college, demo",
			usernames: []field{
				{
					displayName: "email",
					value: "a;ksdjfkad@gmail.com",
					secret: false,
				},
			},
			password: field{
				displayName: "password",
				value: "23984",
				secret: true,
			},
			securityQ: []field{
				{
					displayName: "what was your first car?", 
					value: ";aiodkj",
					secret: true,
				},
			},
			notes: "need to keep email short to write on test day",
		},
		entry{
			name: "wooo myACT",
			tags: "college, demo",
			usernames: []field{
				{
					displayName: "email",
					value: "a;ksharnof2333@gmail.com",
					secret: false,
				},
			},
			password: field{
				displayName: "password",
				value: "02983490832",
				secret: true,
			},
			securityQ: []field{
				{
					displayName: "what was your first CLARINET?", 
					value: ";buffet coprmodn",
					secret: true,
				},
			},
			notes: "netest test ets, \n\treal phone given ",
		},
		entry{
			name: "libary A",
			tags: "library, demo, overdrive",
			usernames: []field{
				{
					displayName: "card",
					value: "9873458974398795843",
					secret: false,
				},
			},
			password: field{
				displayName: "pin",
				value: "09128",
				secret: true,
			},
			notes: "from google doc! ",
		},
		entry{
			name: "libary B",
			tags: "library, demo, overdrive",
			usernames: []field{
				{
					displayName: "card",
					value: "12354126357812",
					secret: false,
				},
			},
			password: field{
				displayName: "pin",
				value: "12356",
				secret: true,
			},
			notes: "from google doc from dada ",
		},
	}

	for i := 0; i < 20; i++{ // put at 52 makes it show the max amount (when 5 already in entries)
		entries = append(entries, entry{name: "test",tags: "demo, test!, smiles"})
	}
	
	// pages is the pages set up for the left top box
	pages := tview.NewPages()

	// this is what everything is in, with it being split between the left and right
	// (the left being another flex split up and down)
	flex := tview.NewFlex()
	flexRow := tview.NewFlex().SetDirection(tview.FlexRow) // change name to flexLeft?

	// this is the text box that contains the commands, on the left and its grid (border)
	commandsText := tview.NewTextView().SetScrollable(true)
	commandsTextGrider := tview.NewGrid().SetBorders(true)
	commands := " commands\n --------\n /home \n /help \n /new \n /find str\n /edit # \n /open # \n /list \n /test"

	// this is the box that the page is set to when at /home
	// probably delete the title as some point, it's just like that for now tho
	emptySadBox := tview.NewBox().SetBorder(true).SetTitle("sad, empty box")

	// string of what is put into the command line
	inputed := ""
	// this is the commandLine as well as its grid (border)
	// maybe change the placeholder?
	// !!!!!!!!!! when the focus is changed and you can't type, have the placeholder say that
	inputer := tview.NewInputField().
		SetLabel("input: ").
		SetFieldWidth(55)
	inputerGrider := tview.NewGrid().SetBorders(true)

	// this function is called when the focus switches back 
	// and one can type in the command line, so it says to look right 
	lookRightInputerPlaceholder := func(){
		inputer.SetPlaceholder("psst look to the right")
	}
	// this function is called when the focus switches away and one
	// cannot type in the command line, so it says so 
	cantTypeInputerPlaceholder := func(){
		inputer.SetPlaceholder("psst you can't type here right now")
	}

	// this is the text box that contains that list entry names and its grid (border)
	listTextRight := tview.NewTextView().SetScrollable(true).SetWrap(false)
	listTextMiddle := tview.NewTextView().SetScrollable(true).SetWrap(false)
	listTextLeft := tview.NewTextView().SetScrollable(true).SetWrap(false)
	listTextFlex := tview.NewFlex()
	listTextFlexGrider := tview.NewGrid().SetBorders(true)

	// this is a text box to print out the entire entries, to test!
	testText := tview.NewTextView().SetScrollable(true)
	testTextGrider := tview.NewGrid().SetBorders(true)

	// this is the function that will do things based off the commands given to inputer
	commandLine := func(key tcell.Key){}


	// this is the text box with the /help info and its grid (border)
	helpText := tview.NewTextView().SetScrollable(true).SetText(" /help \n -----\n\n Do /new in order to put in a new entry. \n Do a;sdkfjkl  \n\n /find is case insensitve.  \n\n Do /open to view an entry. \n You will have to put in the password before you can see the information. \n Passwords and security questions will be blotted out, but they can be copied. (Or highlighted to see them) \n To delete an entry do /edit \n\n do /edit to edit fields or delete an entry. \n in /edit, all edits are permanently saved field by field as you click save \n\n an idea is that you can remove an entry from circulation instead of deleting it. \n\n the values of all fields, except the name of the entry and the tags, are equally encrypted. \n ")
	helpTextGrider := tview.NewGrid().SetBorders(true)

	// for viewing something that you've opened
	// will include a button for editing it ???? !!!!!!!make it a form to also have a button?????
	// the text box with its grider and its function to format the information
	openEntryText := tview.NewTextView().SetScrollable(true)
	openEntryTextGrider := tview.NewGrid().SetBorders(true)
	openEntry := func(i int) string {return "error, openEntry didn't run"}
	

	// when something happens that could give an error it will switch to here
	// and print it and which it's guess of the error.
	// has the upper bit which will say this is the error box
	errorTitleText := tview.NewTextView().SetText(" Uh oh! There was an error:")
	errorText := tview.NewTextView().SetScrollable(true)
	errorTextFlex := tview.NewFlex()
	errorTextFlexGrider := tview.NewGrid().SetBorders(true)


	switchToHome := func(){}


	// the following variables are all for when you adding a new entry (and a new field)


	// this is the form of the new entry area, its grider, its function to set it up
	newEntryForm := tview.NewForm()
	blankNewEntry := func(){}

	newEntryFlex := tview.NewFlex() // to put both the form above and the list of the fields made already
	newEntryFlexGrider := tview.NewGrid().SetBorders(true)
	blankFieldsAdded := func(){}
	newFieldsAddedList := tview.NewList().SetSelectedFocusOnly(true)
	switchToNewFieldsList := func(){}

	// this is the form of adding a new field, its grider, 
	// its flex to situate it, function to set it up.
	newFieldForm := tview.NewForm()
	newFieldFormGrider := tview.NewGrid().SetBorders(true)
	newFieldFormFlex := tview.NewFlex()
	blankNewField := func(){}

	// this is the form for adding a new notes, its grider, flex, and function
	newNoteForm := tview.NewForm()
	newNoteFormGrider := tview.NewGrid().SetBorders(true)
	newNoteFormFlex := tview.NewFlex()
	blankNewNote := func(){}

	// these are all temporary, they are what a new entry or field is set to in
	// order to add it. to clear the following after/before use just set them 
	// equal to entry{} or field{}
	tempEntry := entry{}
	
	tempField := field{}
	// fieldType keeps track of the type of the field, in order to add it to the correct
	// part of the entry.
	fieldType := ""

	// no matter if the formatting changes, it will be username as 0, password as 1, etc.
	// did not let me pass in [3] as input, it must be a slice
	dropDownFields := []string{"username", "password", "security Question"}

	// text to be put on the left side when in /new
	// REPLACE select WITH button?????? MAYBE??? -- ask lucy :)
	newCommands := " /new \n ---- \n move: \n -tab \n -back tab \n -click \n\n select: \n -return \n -click\n\n must name \n entry to \n save it \n\n escape? \n quit"
	newFieldCommands := " /new \n ---- \n move: \n -tab \n -back tab \n -click \n\n select: \n -return \n -click\n\n must name \n field to \n save it \n\n escape? \n quit" //only change from this one to the newCommands is field vs. entry


	// the folowing varialbes are for editing an entry


	// these are the list, its grid, and the function to make the list when
	// /edit an entry
	editList := tview.NewList().SetSelectedFocusOnly(true)
	editListGrider := tview.NewGrid().SetBorders(true)
	blankList := func(i int){}
	runeAlphabet := []rune{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z'}

	// this is the form and its grid and flex for editing a specific field
	// it has two functions, for editing one of the strings in the entry struct
	// and one for editing one of the fields (password, username, securityQ)
	editFieldForm := tview.NewForm()
	editFieldFormGrider := tview.NewGrid().SetBorders(true)
	editFieldFormFlex := tview.NewFlex()
	blankEditFieldForm := func(f *field, fieldArr *[]field, index int, e *entry, pass, edit bool) {}
	blankEditStringForm := func (display, value string, e *entry){}

	// this is a function that solves redundancy in going back to /edit
	// (remaking the list, switching the page, setting the focus)
	// the function uses indexSelectEntry, which is that in that should be the 
	// index of the current entry
	switchToEditList := func(){}
	// this is a variable to represent what entry is being edited,
	// it is outside of any function so it can be used to call the function above,
	// to switch back to the /edit page
	indexSelectEntry := -1

	// the little popup to ask if you are sure when deleting something
	editDeleteText := tview.NewTextView().SetTextAlign(tview.AlignCenter).SetText("delete entry? \nCANNOT BE UNDONE")
	editDeleteForm := tview.NewForm()
	editDeleteFormFlex := tview.NewFlex().SetDirection(tview.FlexRow) //flex of just the text and form
	editDeleteFormFlexGrid := tview.NewGrid().SetBorders(true)
	editDeleteFlex := tview.NewFlex() // flex to set up the smaller flex in the center
	blankEditDeleteEntry := func(){}


	editCommands := " /edit \n ----- \n move: \n -tab \n -back tab \n -arrows keys \n\n select: \n -return \n -click" // similar to newCommands and newFieldCommands

	// ------------------------------------------------ //
	// all varaibles initialized :) function time!
	// ------------------------------------------------ //

	// written out what inputer does with stuff
	commandLine = func(key tcell.Key){
		inputed = inputer.GetText() 
		inputedArr := strings.Split(inputed, " ") 
		returnedList := []string{"error!", "failed \n to get \n anything", "AHHHHHHHHHHHHH"}

		// the following if/else statements check that the number inputed for /edit or /open
		// that there is a number, it is an int, and it corresponds to an entry 
		if (inputedArr[0] == "/open")||(inputedArr[0] == "/edit"){

			indexSelectEntry = -1 //  sets it here to remove any previous doings

			if len(inputedArr) < 2 { // so if there is no number written
				errorText.SetText(" To " + inputedArr[0][1:5] + " an entry you must write " + inputedArr[0] + " and then a number. \n With a space after " + inputedArr[0] + " \n Ex: \n\t" + inputedArr[0] + " 3")
				pages.SwitchToPage("err")
			}else{
				openEditInt, intErr := strconv.Atoi(inputedArr[1])
				if intErr != nil{
					errorText.SetText(" Make sure to only use " + inputedArr[0] + " by writing a number! \n For an example do /help")
					pages.SwitchToPage("err")
				}else{
					if openEditInt >= len(entries){
						errorText.SetText(" The number you entered does not correspond to an entry. \n Do /list to see the entries (and their numbers) that exist.")
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
			returnedList = nameToString(entries, listAllIndexes, " /list \n -----")
			listTextRight.SetText(returnedList[0])
			listTextMiddle.SetText(returnedList[1])
			listTextLeft.SetText(returnedList[2])
			pages.SwitchToPage("/list")
		case "/test":
			testText.SetText(testAllFields(entries))
			pages.SwitchToPage("/test")
		case "/new":
			commandsText.SetText(newCommands)
			blankNewEntry()
			app.SetFocus(newEntryForm)
			cantTypeInputerPlaceholder()
			pages.SwitchToPage("/newEntry")
		case "/help":
			pages.SwitchToPage("/help")
		case "/open":
			if indexSelectEntry > -1 {
				openEntryText.SetText(openEntry(indexSelectEntry)) // taking input, just to be safe smile -- can change that in future
				pages.SwitchToPage("/open")
			}
		case "/edit":
			if indexSelectEntry > -1 {
				commandsText.SetText(editCommands)
				cantTypeInputerPlaceholder()
				switchToEditList()
			}
		case "/find":
			if len(inputedArr) < 2 { 
				errorText.SetText(" To find entries you must write /find and then characters. \n With a space after /find. \n Ex: \n\t /find college")
				pages.SwitchToPage("err")
			}else{
				returnedList = findEntries(entries, inputedArr[1])

				listTextRight.SetText(returnedList[0])
				listTextMiddle.SetText(returnedList[1])
				listTextLeft.SetText(returnedList[2])
				pages.SwitchToPage("/list")
			}
		default:
			errorText.SetText(" That input doesn't match a command! \n Look to the right right to see the possible commands. \n Make sure to spell it correctly!")
			pages.SwitchToPage("err")
		}
		inputer.SetText("")
	}
	// adds the function to inputer so it is run when return is pressed  // move this??
	inputer.SetDoneFunc(commandLine)

	//make func called switch to home that sets everything to rights again
	switchToHome = func(){
		pages.SwitchToPage("/home")
		app.SetFocus(inputer)
		commandsText.SetText(commands)
		lookRightInputerPlaceholder()
	}

	// ----
	// functions for making a new entry
	// ----

	blankNewEntry = func(){

		newEntryForm.Clear(true)
		newFieldsAddedList.Clear()
		
		tempEntry = entry{}
		newEntryForm.
			AddInputField("name", "", 40, nil, func(itemName string){
				tempEntry.name = itemName
			}).
			AddInputField("tags", "", 40, nil, func(tagsInput string){
				tempEntry.tags = tagsInput
			}).
			// this order of the buttons is on purpose and makes sense
			AddButton("new field", func(){
				commandsText.SetText(newFieldCommands)
				blankNewField()
				pages.ShowPage("/newField")
				app.SetFocus(newFieldForm)
			}).
			
			// !!! make it so you can't hit save if there is no tempEntry.name
			AddButton("save entry", func(){
				if tempEntry.name != ""{
					entries = append(entries, tempEntry)
					switchToHome()
				}
			}).
			AddButton("quit", func(){
				switchToHome()
			}). 
			AddButton("notes", func(){
				blankNewNote() 
				// this (blankNewNote) can be deleted and written in the 
				//command inputer cases section if one wants to be able to
				// hit quit of newNote but keep the info
				pages.ShowPage("/newNote")
				app.SetFocus(newNoteForm)
			})
	}

	blankNewField = func(){

		tempField = field{}

		fieldType = ""
		newFieldForm.Clear(true)
		newFieldForm. 
			AddDropDown("new field:", dropDownFields, -1, func(chosenDrop string, index int){
				if index > -1 {
					if newFieldForm.GetFormItemCount() < 2 { // only if there aren't the fields already there (doesn't count buttons)
						fieldType = chosenDrop
							
						switch chosenDrop {
						case dropDownFields[0]: // if username is chosen
							tempField.displayName = "email" // in case it isn't edited, sets this as the default
							newFieldForm.AddInputField("display name:", "email", 50, nil, func(display string){
								tempField.displayName = display
							})
						case dropDownFields[1]: // if password is chosen
							tempField.displayName = chosenDrop // in case it isn't edited, sets this as the default
							newFieldForm.AddInputField("display name:", chosenDrop, 20, nil, func(display string){
								tempField.displayName = display
							})
						case dropDownFields[2]:
							newFieldForm.AddInputField("question:", "", 50, nil, func(display string){
								tempField.displayName = display
							})
						}
						newFieldForm.AddInputField("value:", "", 40, nil, func(value string){
							tempField.value = value
						})
						// if it is a password or security question then it sets bool secret to true
						if (chosenDrop == dropDownFields[1])||(chosenDrop == dropDownFields[2]){
							tempField.secret = true
						} 
					}
				}
			}). 
			AddButton("save field", func(){
				if tempField.displayName != ""{ 
					switch fieldType {
					case dropDownFields[0]:
						tempEntry.usernames = append(tempEntry.usernames, tempField)
					case dropDownFields[1]:
						tempEntry.password = tempField
						dropDownFields[1] = "overide written password"
					case dropDownFields[2]:
						tempEntry.securityQ = append(tempEntry.securityQ, tempField)
					}
					if newEntryForm.GetButtonIndex("edit field") < 0{
						newEntryForm. 
							AddButton("edit field", func(){ // DON'T CHANGE LABEL NAME
								app.SetFocus(newFieldsAddedList)
							})
					}
					blankFieldsAdded()
					commandsText.SetText(newCommands)
					pages.SwitchToPage("/newEntry")
					app.SetFocus(newEntryForm)
				}
			}).
			AddButton("quit", func(){
				commandsText.SetText(newCommands)
				pages.SwitchToPage("/newEntry")
				app.SetFocus(newEntryForm)
			})
	}

	blankNewNote = func(){
		newNoteForm.Clear(true)
		
		toAddArr := [5]string{""}
		newNoteForm.
			AddInputField("notes:", "", 0, nil, func(inputed string){
				toAddArr[0] = inputed
			})

		// i := i because making a new function in a closure (for loop) it
		// has i equal to the last iteration of it (would be 4)
		for i := 1; i < 5; i++ {
			i := i
			newNoteForm.AddInputField("", "", 0, nil, func(inputed string){
				toAddArr[i] = inputed
			})
		}
		
		newNoteForm.
			AddButton("Save", func(){
				toAdd := ""

				// !! maybe have it so if at least one is not blank then all the lines get added, not caring if the others are blank? if someone is doing formatted  with lines in between? or don't and override them??
				for _, n := range toAddArr {
					if n != ""{
						toAdd += n 
						toAdd += " \n\t" // have it do \n and \t as per the formatting in /open
					}
				}
				tempEntry.notes = toAdd
				pages.SwitchToPage("/newEntry")
				app.SetFocus(newEntryForm)
			}). 
			AddButton("Quit", func(){
				tempEntry.notes = "" // must be changed to save when quit
				pages.SwitchToPage("/newEntry")
				app.SetFocus(newEntryForm)
			})
	}

	blankFieldsAdded = func(){ 
		newFieldsAddedList.Clear()

		numFields := 0

		newFieldsAddedList.
			AddItem("move back to top", "", 'a', func(){
				app.SetFocus(newEntryForm)
			})

		for i := range tempEntry.usernames {
			i := i
			u := &tempEntry.usernames[i]
			numFields++

			newFieldsAddedList.AddItem(u.displayName + ":", u.value, runeAlphabet[numFields], func(){
				blankEditFieldForm(u, &tempEntry.usernames, i, &tempEntry, false, false)
				pages.ShowPage("/editField") 
				app.SetFocus(editFieldForm)
			})
		}
		if tempEntry.password.displayName != "" {
			numFields++
			newFieldsAddedList.AddItem(tempEntry.password.displayName + ":", "SECRET!! " + tempEntry.password.value, runeAlphabet[numFields], func(){
				blankEditFieldForm(&tempEntry.password, nil, -1, &tempEntry, true, false)
				pages.ShowPage("/editField") 
				app.SetFocus(editFieldForm)
			})
		}
		for i := range tempEntry.securityQ {
			i := i
			sq := &tempEntry.securityQ[i]
			numFields++

			newFieldsAddedList.AddItem(sq.displayName + ":", "SECRET!! " + sq.value, runeAlphabet[numFields], func(){
				blankEditFieldForm(sq, &tempEntry.securityQ, i, &tempEntry, false, false)
				pages.ShowPage("/editField") 
				app.SetFocus(editFieldForm)
			})
		}
	}

	switchToNewFieldsList = func(){
		blankFieldsAdded()
		if newFieldsAddedList.GetItemCount() > 1 {
			pages.SwitchToPage("/newEntry")
			app.SetFocus(newFieldsAddedList)
		}else{ // if all the fields are deleted, then:
			newFieldsAddedList.Clear()

			editFieldIndex := newEntryForm.GetButtonIndex("edit field")
			if editFieldIndex > -1 {
				newEntryForm.RemoveButton(editFieldIndex)
				pages.SwitchToPage("/newEntry")
				app.SetFocus(newEntryForm)
			}else{
				errorText.SetText("AHHHHHHH for some reason the edit field button wasn't added despite a field later trying to be deleted!!!!")
				pages.SwitchToPage("err")
			}
		}
	}

	// ----
	// function for displaying an entry -- move outside main?
	// ----

	// precondition: i > -1
	openEntry = func(i int) string{
		e := entries[i]
		print := " "

		print += "[" + strconv.Itoa(i) + "] " + e.name + "\n " 
		print += strings.Repeat("-", len([]rune(print))-3) + " \n" // right now it matches under the letters of title, if at -2 then it goes one out
		if e.tags != ""{
			print += " tags: " + e.tags + "\n"
		}
		for _, u := range e.usernames {
			print += " " + u.displayName + ": " + u.value + "\n"
		}
		if e.password.displayName != "" {
			print += " " + e.password.displayName + ": " + strconv.FormatBool(e.password.secret) + "!! " + e.password.value + "\n"
		}
		for _, sq := range e.securityQ {
			print += " " + sq.displayName + ": " + strconv.FormatBool(sq.secret) + "!! " +sq.value + "\n"
		}
		if e.notes != ""{
			print += " notes: " + "\n\t" + e.notes
		}
		return print
	}

	// ----
	// functions for editing an entry
	// ----

	blankList = func(i int){
		editList.Clear()
		e := &entries[i]

		numEntry := 0
		editList.AddItem("leave /edit", "(takes you back to /home)", runeAlphabet[numEntry], func(){
			switchToHome()
		})
		numEntry++
		editList.AddItem("name: ", e.name, runeAlphabet[numEntry], func(){
			blankEditStringForm("name", e.name, e)
			pages.ShowPage("/editField") 
			app.SetFocus(editFieldForm)
		})
		if e.tags != "" {
			numEntry++
			editList.AddItem("tags:", e.tags, runeAlphabet[numEntry], func(){
				blankEditStringForm("tags", e.tags, e)
				pages.ShowPage("/editField") 
				app.SetFocus(editFieldForm)
			})
		}
		for i := range e.usernames {
			i := i
			u := &e.usernames[i]
			numEntry++

			editList.AddItem(u.displayName + ":", u.value, runeAlphabet[numEntry], func(){
				blankEditFieldForm(u, &e.usernames, i, e, false, true)
				pages.ShowPage("/editField") 
				app.SetFocus(editFieldForm)
			})
		}
		if e.password.displayName != "" {
			numEntry++
			editList.AddItem(e.password.displayName + ":", "SECRET!! " + e.password.value, runeAlphabet[numEntry], func(){
				blankEditFieldForm(&e.password, nil, -1, e, true, true)
				pages.ShowPage("/editField") 
				app.SetFocus(editFieldForm)
			})
		}
		for i := range e.securityQ {
			i := i
			sq := &e.securityQ[i]
			numEntry++

			editList.AddItem(sq.displayName + ":", "SECRET!! " + sq.value, runeAlphabet[numEntry], func(){
				blankEditFieldForm(sq, &e.securityQ, i, e, false, true)
				pages.ShowPage("/editField") 
				app.SetFocus(editFieldForm)
			})
		}
		if e.notes != "" {
			numEntry++
			editList.AddItem("notes:", e.notes, runeAlphabet[numEntry], func(){
				blankEditStringForm("notes", e.notes, e)
				pages.ShowPage("/editField") 
				app.SetFocus(editFieldForm)
			})
		}
		numEntry++
		editList.AddItem("delete entry", "(permanant!!)", runeAlphabet[numEntry], func(){
			blankEditDeleteEntry()
			pages.ShowPage("/editDelete")
			app.SetFocus(editDeleteForm)
		})
	}

	// includes a boolean if true it is the password field
	// can pass in nil for the slice, -1 for index if it is the password field

	// take in an extra bool for it to be edit/new field form, to check where to send back to!!!
	blankEditFieldForm = func(f *field, fieldArr *[]field, index int, e *entry, pass, edit bool){
		editFieldForm.Clear(true)

		tempField = field{} // not necessary?,, bc being set on next few lines?
		tempField.displayName = f.displayName
		tempField.value = f.value
		tempField.secret = f.secret

		editFieldForm.
			AddInputField("display name:", tempField.displayName, 40, nil, func(input string){
				tempField.displayName = input
			}).
			AddInputField("value:", tempField.value, 40, nil, func(input string){
				tempField.value = input
			}). 
			AddButton("save", func(){
				*f = tempField
				if edit {
					switchToEditList()
				}else{
					switchToNewFieldsList()
				}
			}). 
			AddButton("quit", func(){
				if edit {
					switchToEditList()
				}else{
					switchToNewFieldsList()
				}
			}).
			AddButton("delete field", func(){
				if pass {
					e.password = field{}
					if edit {
						switchToEditList()
					}else{
						dropDownFields[1] = "password"
						switchToNewFieldsList()
					}
				}else{
					if (fieldArr != nil)&&(index != -1){
						// way this is going to be coded to delete it will change the
						// order of the splice, rewrite this to be in order (SLOWER)
						// if want this to stay in order (maybe)
						(*fieldArr)[index] = (*fieldArr)[len(*fieldArr)-1]
						(*fieldArr) = (*fieldArr)[:len(*fieldArr)-1]
						if edit {
							switchToEditList()
						}else{
							switchToNewFieldsList()
						}
					}else{
						errorText.SetText("AHHHHH the array given to blankEditFieldForm is nil and it shouldnt be!!!!")
						pages.SwitchToPage("err")
					}
				}
			})
	}

	blankEditStringForm = func (display, value string, e *entry){
		if (display != "name")&&(display != "notes")&&(display != "tags"){
			errorText.SetText("AHHHH the input of display should only be notes, tags, name!!")
			pages.SwitchToPage("err")
		}else{

			editFieldForm.Clear(true)
			
			tempDisplay := display 
			tempValue := value

			editFieldForm.
				AddInputField(tempDisplay, tempValue, 40, nil, func(changed string){
					tempValue = changed
				}). 
				AddButton("save", func(){
					switch display{
					case "name":
						e.name = tempValue
					case "tags":
						e.tags = tempValue
					case "notes":
						e.notes = tempValue
					} 
					switchToEditList()
				}). 
				AddButton("quit", func(){
					switchToEditList()
				})

			// make a delete button for editing the name, as each entry must have a name
			if display != "name"{ 
				editFieldForm.AddButton("delete", func(){
					if display == "notes"{
						e.notes = ""
					}else{
						e.tags = ""
					}
					switchToEditList()
				})
			}
		}	
	}

	// maybe just have it have input as indexSelectEntry and move it outside func main? 
	switchToEditList = func(){
		blankList(indexSelectEntry)
		pages.SwitchToPage("/edit")
		app.SetFocus(editList)
	}

	blankEditDeleteEntry = func(){
		editDeleteForm.Clear(true)
		editDeleteForm.SetButtonsAlign(tview.AlignCenter)
		editDeleteForm.
			AddButton("save", func(){
				switchToEditList()
			}).
			AddButton("delete", func(){ // this deletes it, slower version, keeps everything in order
				copy(entries[indexSelectEntry:], entries[indexSelectEntry+1:])
				entries[len(entries)-1] = entry{} // ask dada why this is here?
				entries = entries[:len(entries)-1]
				testText.SetText(testAllFields(entries))
				
				switchToHome()
			})
	}


	// ------------------------------------------------ //
	// setting up the griders, pages, flexes :)
	// ------------------------------------------------ //

	
	listTextFlex. 
		AddItem(listTextRight, 0, 1, false). 
		AddItem(listTextMiddle, 0, 1, false). 
		AddItem(listTextLeft, 0, 1, false)

	newFieldFormFlex.
		AddItem(nil, 0, 1, false). 
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow). 
			AddItem(nil, 0, 2, false). 
			AddItem(newFieldFormGrider, 0, 3, false). 
			AddItem(nil, 0, 1, false), 0, 4, false)

	errorTextFlex.SetDirection(tview.FlexRow). 
		AddItem(errorTitleText, 0, 1, false).
		AddItem(errorText, 0, 8, false)

	newNoteFormFlex. 
		AddItem(nil, 0, 1, false). 
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow). 
			AddItem(nil, 0, 2, false). 
			// following two, 5 is the max for changing
			AddItem(newNoteFormGrider, 0, 6, false). // 4 fits 3 input + buttons,,5 fits 4 input + buttons
			AddItem(nil, 0, 1, false), 0, 5, false) 

	newEntryFlex.SetDirection(tview.FlexRow).
		AddItem(newEntryForm, 0, 1, false). 
		AddItem(newFieldsAddedList, 0, 2, false) // 1:2 is the maximum  

	editFieldFormFlex.
		AddItem(nil, 0, 1, false). 
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow). 
			AddItem(nil, 0, 1, false). 
			AddItem(editFieldFormGrider, 0, 3, false). 
			AddItem(nil, 0, 1, false), 0, 4, false)

	editDeleteFormFlex.
		AddItem(editDeleteText, 0, 1, false). 
		AddItem(editDeleteForm, 0, 1, false)

	editDeleteFlex. 
		AddItem(nil, 0, 1, false). 
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow). 
			AddItem(nil, 0, 2, false). // 2 
			AddItem(editDeleteFormFlexGrid, 0, 2, false). 
			AddItem(nil, 0, 2, false), 0, 1, false).  //3
		AddItem(nil, 0, 1, false)


	// uses a function to add each thing to its perspective grid
	grider(inputer, inputerGrider)
	grider(commandsText, commandsTextGrider)
	grider(listTextFlex, listTextFlexGrider)
	grider(testText, testTextGrider)
	grider(newEntryFlex, newEntryFlexGrider)
	grider(newFieldForm, newFieldFormGrider)
	grider(newNoteForm, newNoteFormGrider)
	grider(helpText, helpTextGrider)
	grider(errorTextFlex, errorTextFlexGrider)
	grider(openEntryText, openEntryTextGrider)
	grider(editList, editListGrider)
	grider(editFieldForm, editFieldFormGrider)
	grider(editList, editListGrider)
	grider(editDeleteFormFlex, editDeleteFormFlexGrid)


	// all the different pages are added here
	pages.
		AddPage("/home", emptySadBox, true, true). 
		AddPage("/list", listTextFlexGrider, true, false). 
		AddPage("/test", testTextGrider, true, false). 
		AddPage("/newEntry", newEntryFlexGrider, true, false).
		AddPage("/newField", newFieldFormFlex, true, false). 
		AddPage("/newNote", newNoteFormFlex, true, false). 
		AddPage("/help", helpTextGrider, true, false). 
		AddPage("err", errorTextFlexGrider, true, false). 
		AddPage("/open", openEntryTextGrider, true, false). 
		AddPage("/edit", editListGrider, true, false). 
		AddPage("/editField", editFieldFormFlex, true, false). 
		AddPage("/editDelete", editDeleteFlex, true, false)


	// sets up the flex row of the left side, top is the pages bottom is the inputer
	// ratio of 8:1 is the maximum that it can be (9:1 and 100:1 are the same as 8:1)
	flexRow. 
		AddItem(pages, 0, 8, false). 
		AddItem(inputerGrider, 0, 1, false)

	// the greater flex consisting of the left and right sides
	flex. 
		AddItem(flexRow, 0, 5, false). 
		AddItem(commandsTextGrider, 0, 1, false)

	switchToHome()

	if err := app.SetRoot(flex, true).SetFocus(inputer).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}


// this is the function used to put any type of primitive
// into a grid, as in using the grid to make a border
func grider(prim tview.Primitive, grid *tview.Grid){
	grid.AddItem(prim, 0, 0, 1, 1, 0, 0, false)
}


// this finds all the entries that has a string in its name or tags

// make extra check, if str is over a certain character count then don't print all the characters in a line, would look funny. also can garanetted not any entries per that amount so can skip all the cycling through
func findEntries(entries []entry, str string) []string{
	indexes := []int{}
	str = strings.ToLower(str)
	for i, e := range entries {
		if (strings.Contains(strings.ToLower(e.name), str))||(strings.Contains(strings.ToLower(e.tags), str)) {
			indexes = append(indexes, i)
		}
	}
	if len(indexes) > 0{
		return nameToString(entries, indexes, " /find " + str + " \n " + strings.Repeat("-", len([]rune(str))+6))
	}else{ 
		return []string{" /find " + str + " \n " + strings.Repeat("-", len([]rune(str))+6) + "\n no entries found", "", ""}
	}
}


// this is the function that formats each name as: " [0] twitter"
// to be printed out in /list
// remane fnction?? lol
// strconv.Itoa(i) turns int to string
// the str taken in will be: " /find str \n-----" or " /list \n -----"
func nameToString(entries []entry, indexes []int, str string) []string{
	list := []string{str, "", ""}

	third := len(indexes)/3
	if third < 19 {
		third = 19
	}
	indexesIndex := 0
	listIndex := 0

	// currently divisies it up evenly between three columns, if want to make it to not have to scroll for first two columns, replace the following line with if (entriesIndex == 18)||(entriesIndex == 38)

	for indexesIndex < len(indexes){
		if (indexesIndex == third)||(indexesIndex == third*2){
			listIndex++
			list[listIndex] = "\n"
		}
		list[listIndex] += "\n [" + strconv.Itoa(indexes[indexesIndex]) + "] " + entries[indexes[indexesIndex]].name
		indexesIndex++
	}

	return list
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
