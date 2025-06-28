/*
	MIT License
	Copyright (c) 2022 Kezia Sharnoff

	pass.go
	Terminal run password manager. This file manages the widgets and
	functionality within the password manager -- encryption & file writing and
	start up are in encrypt/encrypt.go and createEncr.go respectively.
*/

package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/atotto/clipboard" // copies the data to clipboard
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	// encryption
	"crypto/cipher"
	"github.com/ksharnoff/pass/encrypt"

	// getting terminal size
	"golang.org/x/term"
	"os"
)

// This is in the map in /reused and /comp. It has the fields associated
// with a certain password.
type reusedPass struct {
	displayName string
	entryName   string
	entryIndex  int
}
type passApp struct {
	app             *tview.Application
	boxPages           *tview.Pages
	infoText        *tview.TextView
	entries         []encrypt.Entry
	ciphBlock       cipher.Block
	fieldsAddedList *tview.List
	newEntryForm    *tview.Form
	indexSelected   int
	tempEntry       encrypt.Entry
	tempField       encrypt.Field
	editDeleteText  *tview.TextView
	listText        *tview.TextView
	passwordPages   *tview.Pages
	passPages       *tview.Pages
	input           *tview.InputField
	text            *tview.TextView
	list            *tview.List
	form            *tview.Form
	leftPages       *tview.Pages
}

// You can uncomment out the next two lines and comment out the default
// colors in order for it to have a higher contrast that complies with WCAG
// AAA. lavender is label and shortcut names. blue is secondary text in
// lists, buttons in forms, and the input field color.
// var lavender = tcell.GetColor("white")       // uncomment for higher contrast
// var blue = tcell.NewRGBColor(0, 0, 255)      // uncomment for higher contrast
var lavender = tcell.NewRGBColor(149, 136, 204) // comment for higher contrast
var blue = tcell.NewRGBColor(106, 139, 166)     // comment for higher contrast

var white = tcell.GetColor("white")

// The terminal width and height represent the difference from the size that
// this was designed for (84x28). It is a global variable to avoid having to
// pass it to the many functions that make text columns and widgets.
var width = 0
var height = 0

func main() {
	passApp{}.updateTerminalSize()

	a := passApp{

		app: tview.NewApplication(),

		// Pages is the pages set up for the left box in pass
		pages: tview.NewPages(),

		// This is the text box on the right that contains information that changes
		// depending on what the user is doing.
		infoText: newScrollableTextView().SetWrap(false),

		// Entries is the persistent slice of all the entries used throughout.
		// The following entry names will only be seen if the manager opens
		// without loading a file.
		entries: []encrypt.Entry{
			encrypt.Entry{Name: "QUIT NOW, DANGER", Circulate: true},
			encrypt.Entry{Name: "SOMETHING'S VERY", Circulate: true},
			encrypt.Entry{Name: "BROKEN. QUIT!", Circulate: true},
			encrypt.Entry{Name: "DATA NOT LOADED", Circulate: true},
		},

		// Normally its the key that gets passed around not the cipher block,
		// but I chose to do it this way as the functions in encrypt.go made
		// more sense.

		// This is the variable for what entry is selected. It is set in
		// commandLineActions and used for a function below.
		indexSelected: -1,

		// This is the fields added so far list and its function, used in /new.
		fieldsAddedList: newList(),

		newEntryForm: newForm(),

		// These are temporary and used when someone is making a new entry, a new
		// field, or editing an existing entry.
		tempEntry: encrypt.Entry{},
		tempField: encrypt.Field{},

		// This is the little pop up to ask if you're sure when you want to delete
		// an entry.
		editDeleteText: tview.NewTextView().
			SetTextAlign(tview.AlignCenter).
			SetText("delete entry?\nCANNOT BE UNDONE"),

		// This is for /list as well as /find, the text view to show the
		// entries. It shouldn't wrap because that would look bad. 
		listText: newScrollableTextView().SetWrap(false),

		// passwordPages switches between passBox and passErr
		passwordPages: tview.NewPages(),

		// passPages switches between the locked screen and the unlocked normal
		// password manager.
		passPages: tview.NewPages(),

		// For putting in initial master password and later for navigating and
		// commands in the manager -- the settings are redone
		input: newInputField().
			SetLabel("password: ").
			SetFieldWidth(71 + width).
			SetMaskCharacter('*'),

		// Text box for errors on the password screen, errors in the manager,
		// /help, /open, /reused, /comp, /test, title box for /list and /find
		text: newScrollableTextView().SetDynamicColors(true).SetWrap(true),

		// This list is used for copen, pick, picc, flist, and edit
		list: newList(),

		// editFieldForm, newFieldForm, editDeleteForm, newNoteForm
		form: newForm(),

		// contains a.boxPages and the command line and also the boxes for when
		leftPages: tview.NewPages(),
	}
	// Define the SetDoneFuncs after because they use methods
	a.list.SetDoneFunc(a.switchToHome)
	a.input.SetDoneFunc(a.passActions)
	a.fieldsAddedList.SetDoneFunc(func() {
		a.infoText.SetText(getInfoText("newEntry"))
		a.app.SetFocus(a.newEntryForm)
	})

	a.setUpPages()

	if err := a.app.SetRoot(a.passPages, true).SetFocus(a.input).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

func (a *passApp) setUpPages() {
	// Top box of left side, the pages flipped between when you can see the
	// command input line
	a.boxPages.
		AddPage("home", tview.NewBox().SetBorder(true).SetTitle("sad, empty box"), true, false).
		AddPage("list", grider(newFlex("list", a.text, a.listText)), true, false).
		AddPage("test", grider(a.text), true, false).
		AddPage("help", grider(a.text), true, false).
		AddPage("err", grider(newFlex("error", tview.NewTextView().SetText(" Uh oh! There was an error:"), a.text)), true, false).
		AddPage("open", grider(a.text), true, false).
		AddPage("comp", grider(a.text), true, false).
		AddPage("reused", grider(a.text), true, false)

	// Left side pages where "commandLine" is when you should see the input field
	// at the bottom and the rest cover it up
	a.leftPages.
		AddPage("commandLine", newFlex("leftPages", a.boxPages, a.input), true, false).
		AddPage("copen", grider(a.list), true, false).
		AddPage("pick", grider(a.list), true, false).
		AddPage("edit", grider(a.list), true, false).
		AddPage("newEntry", grider(newFlex("newEntry", a.newEntryForm, a.fieldsAddedList)), true, false).
		AddPage("editFieldStr", newFlex("editFieldStr", a.form), true, false).
		AddPage("newNote", newFlex("newNote", a.form), true, false).
		AddPage("editDelete", newFlex("editDelete", a.editDeleteText, a.form), true, false).
		AddPage("edit-editField", newFlex("editEditField", a.form), true, false).
		AddPage("new-editField", newFlex("newEditField", a.form), true, false).
		AddPage("newField", newFlex("newField", a.form), true, false)

	// Password manager boxes, error box and blank box. 
	a.passwordPages.
		AddPage("passBox", tview.NewBox().SetBorder(true), true, true).
		AddPage("passErr", grider(newFlex("passErr", tview.NewTextView().SetText(" Uh oh! There was an error in signing in:"), a.text)), true, false)

	// Contains the password screen and the password manager
	a.passPages.
		AddPage("passInput", newFlex("password", a.passwordPages, a.input), true, true).
		AddPage("passManager", newFlex("main", a.leftPages, a.infoText), true, false)
}

// Switches to home, rights everything again.
func (a *passApp) switchToHome() {
	a.leftPages.SwitchToPage("commandLine")
	a.boxPages.SwitchToPage("home")
	a.app.SetFocus(a.input)
	a.infoText.SetText(getInfoText("home"))
	a.app.EnableMouse(true)
}

// Switches to the error page, sets text to the inputted err.
func (a *passApp) switchToError(err string) {
	a.text.SetText(err)
	a.boxPages.SwitchToPage("err")
}

// This tries to write to file, if it fails, it switches to the error page
// and returns false. The reason for returning false is so that when used
// else where it doesn't switch to error page and then immediately switch
// else where so it can't be seen.
func (a *passApp) writeFileErrNone() bool {
	writeErr := encrypt.WriteToFile(a.entries, a.ciphBlock)
	if writeErr != "" {
		a.switchToError(writeErr)
		return false
	}
	return true
}

// Switches back to the edit list after editing a specific field. It remakes
// the list each time and uses indexSelected. It takes in a bool to know
// whether or not to write to file the changes, as well as whether or not
// to update the last modified time.
func (a *passApp) switchToEditList(modified bool) {
	succeeded := true
	if modified {
		a.entries[a.indexSelected].Modified = time.Now()
		succeeded = a.writeFileErrNone()
	}

	if succeeded {
		a.app.EnableMouse(false)
		a.blankEditlist(a.indexSelected)
		a.leftPages.SwitchToPage("edit")
		a.app.SetFocus(a.list)
		a.infoText.SetText(getInfoText("editEntry"))
	}
}

// This just uses tempEntry to get the fields, this works because
// tempEntry is defined to be equal to entry e in blankNewEntry when called
// after /copy.
func (a *passApp) blankFieldsAdded() {
	a.fieldsAddedList.Clear()
	letter := newCharIterator()

	if a.newEntryForm.GetButtonIndex("edit field") < 0 { // if there isn't one already
		a.newEntryForm.
			// Don't change this label name, breaks stuff later.
			AddButton("edit field", func() {
				a.infoText.SetText(getInfoText("newFieldsAdded"))
				a.app.SetFocus(a.fieldsAddedList)
			})
	}
	a.fieldsAddedList.
		AddItem("move back to top", "", rune(letter), func() {
			a.infoText.SetText(getInfoText("newEntry"))
			a.app.SetFocus(a.newEntryForm)
		})
	for _, u := range a.tempEntry.Urls {
		letter = increment(letter)
		a.fieldsAddedList.AddItem("url:", u, rune(letter), func() {
			a.infoText.SetText(getInfoText("newField"))
			a.blankEditStringForm("url", u, &a.tempEntry, false)
			a.leftPages.ShowPage("editFieldStr")
			a.app.SetFocus(a.form)
		})
	}
	for i := range a.tempEntry.Usernames {
		u := &a.tempEntry.Usernames[i]
		letter = increment(letter)

		a.fieldsAddedList.AddItem(u.DisplayName+":", u.Value, rune(letter), func() {
			a.infoText.SetText(getInfoText("newField"))
			a.blankEditFieldForm(u, &a.tempEntry.Usernames, i, false)
			a.leftPages.ShowPage("new-editField")
			a.app.SetFocus(a.form)
		})
	}
	for i := range a.tempEntry.Passwords {
		p := &a.tempEntry.Passwords[i]
		letter = increment(letter)

		a.fieldsAddedList.AddItem(p.DisplayName+":", "[black:black]"+p.Value, rune(letter), func() {
			a.infoText.SetText(getInfoText("newField"))
			a.blankEditFieldForm(p, &a.tempEntry.Passwords, i, false)
			a.leftPages.ShowPage("new-editField")
			a.app.SetFocus(a.form)
		})
	}
	for i := range a.tempEntry.SecurityQ {
		sq := &a.tempEntry.SecurityQ[i]
		letter = increment(letter)

		a.fieldsAddedList.AddItem(sq.DisplayName+":", "[black:black]"+sq.Value, rune(letter), func() {
			a.infoText.SetText(getInfoText("newField"))
			a.blankEditFieldForm(sq, &a.tempEntry.SecurityQ, i, false)
			a.leftPages.ShowPage("new-editField")
			a.app.SetFocus(a.form)
		})
	}
}

// To be used when each field is edited in /new. It creates the button
// 'edit fields' after creation of first field in /new. It will appear
// there already if you are in /copy # and # has fields. If doSwitch is
// true, then you swap focus to the list of fields already added.
func (a *passApp) switchToNewFieldsList(doSwitch bool) {
	a.blankFieldsAdded()
	if (doSwitch) && (a.fieldsAddedList.GetItemCount() > 1) {
		a.leftPages.SwitchToPage("newEntry")
		a.infoText.SetText(getInfoText("newFieldsAdded"))
		a.app.SetFocus(a.fieldsAddedList)
	}

	if a.fieldsAddedList.GetItemCount() < 2 { // if all the fields are deleted, then:
		a.fieldsAddedList.Clear()
		editFieldIndex := a.newEntryForm.GetButtonIndex("edit field")
		if editFieldIndex > -1 {
			a.newEntryForm.RemoveButton(editFieldIndex)
			a.leftPages.SwitchToPage("newEntry")
			a.infoText.SetText(getInfoText("newEntry"))
			a.app.SetFocus(a.newEntryForm)
		} else {
			a.switchToError(" For some reason the edit field button wasn't added despite a field later trying to be deleted!\n that's not supposed to happen!")
		}
	}
}

// Takes in an extra boolean to know if its from /edit or /new, in order to
// know where to go back to.
func (a *passApp) blankEditFieldForm(f *encrypt.Field, fieldArr *[]encrypt.Field, index int, edit bool) {
	a.form.Clear(true)
	a.tempField.DisplayName = f.DisplayName
	a.tempField.Value = f.Value

	a.form.
		AddInputField("display name:", a.tempField.DisplayName, 40+width, nil, func(input string) {
			a.tempField.DisplayName = input
		}).
		AddInputField("value:", a.tempField.Value, 40+width, nil, func(input string) {
			a.tempField.Value = input
		}).
		AddButton("save", func() {
			if len([]rune(a.tempField.DisplayName)) < 1 {
				return
			}

			*f = a.tempField
			if edit {
				a.switchToEditList(true) // true meaning it was modified
			} else {
				a.switchToNewFieldsList(true) // true meaning keep in the list section
			}
		}).
		AddButton("quit", func() {
			if edit {
				a.switchToEditList(false) // false meaning not modified
			} else {
				a.switchToNewFieldsList(true) // true meaning keep in list section
			}
		}).
		AddButton("delete field", func() {
			if (fieldArr != nil) && (index != -1) {
				// Currently it changes the order when the element
				// is deleted from the slice. If this is wanted to
				// stay in order, then it should be rewritten.
				(*fieldArr)[index] = (*fieldArr)[len(*fieldArr)-1]
				(*fieldArr) = (*fieldArr)[:len(*fieldArr)-1]
				if edit {
					a.switchToEditList(true) // true meaning modified
				} else {
					a.switchToNewFieldsList(true) // true meaning keep in list section
				}
			} else {
				a.switchToError(" The slice given to blankEditFieldForm is nil \n and it shouldn't be! or the index is -1 which it also shouldn't be!")
			}
		})
}

// Takes in a pointer to tempEntry if in /new. Takes in a pointer to an
// entry if in /edit.
func (a *passApp) blankNewField(e *encrypt.Entry) {
	edit := false

	dropDownFields := []string{"url", "username", "password", "security question"}

	// Only adds tags and url as an option to add on if it is in /edit
	if e != &a.tempEntry {
		edit = true
		if e.Tags == "" {
			dropDownFields = append(dropDownFields, "tags")
		}
	}
	a.tempField = encrypt.Field{}
	tempStr := ""
	fieldType := "" // To track what field is changing
	a.form.Clear(true)

	fieldDropDown := tview.NewDropDown().
		SetLabel("new field: ").
		SetCurrentOption(-1).
		// changes the colors of the drop down options -- selected & unselected styles
		SetListStyles(tcell.Style{}.Background(blue).Foreground(white), tcell.Style{}.Background(white).Foreground(blue))

	fieldDropDown.SetOptions(dropDownFields, func(chosenDrop string, index int) {
		for a.form.GetFormItemCount() > 1 { // needed for when you change your mind
			a.form.RemoveFormItem(1)
		}
		fieldType = chosenDrop
		if index > -1 { // If something is chosen
			switch fieldType {
			case "tags":
				a.form.AddInputField("tags:", a.tempEntry.Tags, 50+width, nil, func(tags string) {
					tempStr = tags
				})
			case "url":
				a.form.AddInputField("url:", "", 50+width, nil, func(url string) {
					tempStr = url
				})
			default: // username, password, security question
				inputLabel := "display name:"
				initialValue := ""

				switch fieldType {
				case "username":
					initialValue = "email"
				case "password":
					initialValue = "password"
				case "security question":
					inputLabel = "question:"
				}

				a.tempField.DisplayName = initialValue

				a.form.AddInputField(inputLabel, initialValue, 50+width, nil, func(display string) {
					a.tempField.DisplayName = display
				})

				a.form.AddInputField("value:", "", 50+width, nil, func(value string) {
					a.tempField.Value = value
				})
			}
		}
	})
	a.form.AddFormItem(fieldDropDown).AddButton("save field", func() {
		if (a.tempField.DisplayName != "") || (tempStr != "") {
			switch fieldType {
			case "username":
				e.Usernames = append(e.Usernames, a.tempField)
			case "password":
				e.Passwords = append(e.Passwords, a.tempField)
			case "security question":
				e.SecurityQ = append(e.SecurityQ, a.tempField)
			case "tags":
				e.Tags = tempStr
			case "url":
				e.Urls = append(e.Urls, tempStr)
			}
			if !edit { // If in /new
				a.blankFieldsAdded()
				a.infoText.SetText(getInfoText("newEntry"))
				a.leftPages.SwitchToPage("newEntry")
				a.app.SetFocus(a.newEntryForm)
			} else { // If in /edit
				a.switchToEditList(true)
			}
		}
	}).
		AddButton("quit", func() {
			if !edit {
				a.infoText.SetText(getInfoText("newEntry"))
				a.leftPages.SwitchToPage("newEntry")
				a.app.SetFocus(a.newEntryForm)
			} else {
				a.switchToEditList(false)
			}
		})
}

func (a *passApp) blankEditDeleteEntry() {
	a.form.Clear(true)
	a.form.SetButtonsAlign(tview.AlignCenter)
	a.form.
		AddButton("cancel", func() {
			a.switchToEditList(false)
		}).
		AddButton("delete", func() { // deletes element from slice, slower version, keeps everything else in order, copied the code from a website lol
			copy(a.entries[a.indexSelected:], a.entries[a.indexSelected+1:])
			a.entries = a.entries[:len(a.entries)-1]
			if a.writeFileErrNone() {
				a.switchToHome()
			}
		})
}

func (a *passApp) blankEditlist(i int) {
	a.list.Clear()
	e := &a.entries[i]
	letter := newCharIterator()

	a.list.AddItem("leave /edit "+strconv.Itoa(i), "(takes you back to /home)", rune(letter), func() {
		a.switchToHome()
	})
	letter = increment(letter)
	a.list.AddItem("name:", e.Name, rune(letter), func() {
		a.infoText.SetText(getInfoText("editField"))
		a.blankEditStringForm("name", e.Name, e, true)
		a.leftPages.ShowPage("editFieldStr")
		a.app.SetFocus(a.form)
	})
	if e.Tags != "" {
		letter = increment(letter)
		a.list.AddItem("tags:", e.Tags, rune(letter), func() {
			a.infoText.SetText(getInfoText("editField"))
			a.blankEditStringForm("tags", e.Tags, e, true)
			a.leftPages.ShowPage("editFieldStr")
			a.app.SetFocus(a.form)
		})
	}
	for _, u := range e.Urls {
		letter = increment(letter)
		a.list.AddItem("url:", u, rune(letter), func() {
			a.infoText.SetText(getInfoText("editField"))
			a.blankEditStringForm("url", u, e, true)
			a.leftPages.ShowPage("editFieldStr")
			a.app.SetFocus(a.form)
		})
	}
	for i := range e.Usernames {
		u := &e.Usernames[i]
		letter = increment(letter)

		a.list.AddItem(u.DisplayName+":", u.Value, rune(letter), func() {
			a.infoText.SetText(getInfoText("editField"))
			a.blankEditFieldForm(u, &e.Usernames, i, true)
			a.leftPages.ShowPage("edit-editField")
			a.app.SetFocus(a.form)
		})
	}
	for i := range e.Passwords {
		p := &e.Passwords[i]
		letter = increment(letter)

		a.list.AddItem(p.DisplayName+":", "[black:black]"+p.Value, rune(letter), func() {
			a.infoText.SetText(getInfoText("editField"))
			a.blankEditFieldForm(p, &e.Passwords, i, true)
			a.leftPages.ShowPage("edit-editField")
			a.app.SetFocus(a.form)
		})
	}
	for i := range e.SecurityQ {
		sq := &e.SecurityQ[i]
		letter = increment(letter)

		a.list.AddItem(sq.DisplayName+":", "[black:black]"+sq.Value, rune(letter), func() {
			a.infoText.SetText(getInfoText("editField"))
			a.blankEditFieldForm(sq, &e.SecurityQ, i, true)
			a.leftPages.ShowPage("edit-editField")
			a.app.SetFocus(a.form)
		})
	}
	condensedNotes := ""
	emptyNotes := true
	for _, n := range e.Notes {
		if n != "" {
			condensedNotes += n + ", "
			emptyNotes = false
		}
	}
	if !emptyNotes {
		letter = increment(letter)
		a.list.AddItem("notes:", condensedNotes, rune(letter), func() {
			a.infoText.SetText(getInfoText("editField"))
			a.blankNewNote(e)
			a.leftPages.ShowPage("newNote")
			a.app.SetFocus(a.form)
		})
	} else {
		letter = increment(letter)
		a.list.AddItem("add notes:", "(none written so far)", rune(letter), func() {
			a.infoText.SetText(getInfoText("editField"))
			a.blankNewNote(e)
			a.leftPages.ShowPage("newNote")
			a.app.SetFocus(a.form)
		})
	}
	newFieldStr := ""
	if e.Tags == "" {
		newFieldStr += "tags, "
	}
	letter = increment(letter)
	a.list.AddItem("add new field", newFieldStr+"urls, usernames, passwords, security questions", rune(letter), func() {
		a.infoText.SetText(getInfoText("editField"))
		// code copied from blankNewEntry
		a.blankNewField(e)
		a.leftPages.ShowPage("newField")
		a.app.SetFocus(a.form)
	})
	letter = increment(letter)
	if e.Circulate { // If it is in circulation, option to opt out
		a.list.AddItem("remove from circulation", "(not permanent), check /help for info", rune(letter), func() {
			e.Circulate = false
			a.switchToEditList(true)
		})

	} else { // If it's not in circulation, option to opt back in
		a.list.AddItem("add back to circulation", "(not permanent), check /help for info", rune(letter), func() {
			e.Circulate = true
			a.switchToEditList(true)
		})
	}
	letter = increment(letter)
	a.list.AddItem("delete entry", "(permanent!)", rune(letter), func() {
		a.infoText.SetText(getInfoText("editField"))
		a.blankEditDeleteEntry()
		a.leftPages.ShowPage("editDelete")
		a.app.SetFocus(a.form)
	})
}

// An entry is passed in for /copy. If making a brand new entry, then a
// blank tempEntry is passed in.
func (a *passApp) blankNewEntry(e encrypt.Entry) {
	a.newEntryForm.Clear(true)
	a.fieldsAddedList.Clear()

	// This must be done one by one because of pointer shenanigans
	// Usernames, Passwords, SecurityQ, Urls are slices so must
	// be copied manually. Right now, notes is limited to six strings
	// [6]string so is an array, not a pointer.
	a.tempEntry.Name = e.Name
	a.tempEntry.Tags = e.Tags

	a.tempEntry.Urls = make([]string, len(e.Urls))
	copy(a.tempEntry.Urls, e.Urls)

	a.tempEntry.Usernames = make([]encrypt.Field, len(e.Usernames))
	copy(a.tempEntry.Usernames, e.Usernames)

	a.tempEntry.Passwords = make([]encrypt.Field, len(e.Passwords))
	copy(a.tempEntry.Passwords, e.Passwords)

	a.tempEntry.SecurityQ = make([]encrypt.Field, len(e.SecurityQ))
	copy(a.tempEntry.SecurityQ, e.SecurityQ)

	a.tempEntry.Notes = e.Notes
	a.tempEntry.Circulate = true

	a.newEntryForm.
		AddInputField("name:", a.tempEntry.Name, 58+width, nil, func(itemName string) {
			a.tempEntry.Name = itemName
		}).
		AddInputField("tags:", a.tempEntry.Tags, 58+width, nil, func(tagsInput string) {
			a.tempEntry.Tags = tagsInput
		}).
		AddCheckbox("circulate:", true, func(checked bool) {
			a.tempEntry.Circulate = checked
		}).
		// this order of the buttons is on purpose and makes sense
		AddButton("new field", func() {
			a.infoText.SetText(getInfoText("newField"))
			a.blankNewField(&a.tempEntry)
			a.leftPages.ShowPage("newField")
			a.app.SetFocus(a.form)
		}).
		// You can't hit save if there's no name
		AddButton("save entry", func() {
			if a.tempEntry.Name != "" {
				a.tempEntry.Created = time.Now()
				a.entries = append(a.entries, a.tempEntry)
				if a.writeFileErrNone() { // if successfully wrote to file, then it switches to home, if not then it switches to error page
					a.switchToHome()
				}
			}
		}).
		AddButton("quit", func() {
			a.switchToHome()
		}).
		AddButton("notes", func() {
			a.blankNewNote(&a.tempEntry)
			a.leftPages.ShowPage("newNote")
			a.app.SetFocus(a.form)
		})
	// Put at the end so in case there is already fields it puts the button at the end
	a.switchToNewFieldsList(false)
}

// Takes in a pointer to an entry if used in /edit. Takes in a pointer to
// tempEntry if in /new.
func (a *passApp) blankNewNote(e *encrypt.Entry) {
	a.form.Clear(true)
	toAdd := e.Notes

	a.form.
		AddInputField("notes:", toAdd[0], 0, nil, func(inputed string) {
			toAdd[0] = inputed
		})

	for i := 1; i < 6; i++ {
		a.form.AddInputField("", toAdd[i], 0, nil, func(inputed string) {
			toAdd[i] = inputed
		})
	}

	a.form.
		AddButton("save", func() {
			e.Notes = toAdd
			if e == &a.tempEntry { // if this is being done in /new
				a.leftPages.SwitchToPage("newEntry")
				a.app.SetFocus(a.newEntryForm)
			} else { // if this is being done in /edit
				a.switchToEditList(true)
			}
		}).
		AddButton("quit", func() {
			if e == &a.tempEntry { // if being done in /new
				a.leftPages.SwitchToPage("newEntry")
				a.app.SetFocus(a.newEntryForm)
			} else { // if being done in /edit
				a.switchToEditList(false)
			}
		}).
		AddButton("delete", func() {
			e.Notes = [6]string{} // assigns the whole array at once
			if e == &a.tempEntry {
				a.leftPages.SwitchToPage("newEntry")
				a.app.SetFocus(a.newEntryForm)
			} else {
				a.switchToEditList(true)
			}
		})
}

// For editing the name, tags, or url.
func (a *passApp) blankEditStringForm(display, value string, e *encrypt.Entry, edit bool) {
	if (display != "name") && (display != "tags") && (display != "url") {
		a.switchToError(" Unexpected input!\n blankEditStringForm can only change name, tags, or url")
		return
	}

	// if url was inputted, will need to find the index from its list in
	// order to change or delete it later
	index := -1
	if display == "url" {
		for i, u := range e.Urls {
			if strings.Contains(u, value) {
				index = i
				break
			}
		}
	}

	a.form.Clear(true)
	tempDisplay := display
	tempValue := value
	a.form.
		AddInputField(tempDisplay+":", tempValue, 50+width, nil, func(changed string) {
			tempValue = changed
		}).
		AddButton("save", func() {
			switch display {
			case "name":
				e.Name = tempValue
			case "tags":
				e.Tags = tempValue
			case "url":
				if index < -1 {
					a.switchToError("Tried to edit url in an entry, but could not find the url in the entry's list of urls")
					return
				}
				e.Urls[index] = tempValue
			}
			if (display == "tags") || (edit) {
				a.switchToEditList(true)
			} else {
				a.switchToNewFieldsList(true)
			}
		}).
		AddButton("quit", func() {
			if (display == "tags") || (edit) {
				a.switchToEditList(true)
			} else {
				a.switchToNewFieldsList(true)
			}
		})
	// Can only delete tags or url, not the name
	if display == "tags" || display == "urls" {
		a.form.AddButton("delete", func() {

			if display == "tags" {
				e.Tags = ""
			} else { // is url

				// where index is the index of the inputted value in the
				// slice of urls of the entry
				// should not happen because value should be in the entry
				// list because it was given from the entry!
				if index < -1 {
					a.switchToError("Tried to delete url from an entry, but could not find the url in the entry's list of urls")
					return
				}

				// code copied from form
				// Currently it changes the order when the element
				// is deleted from the slice. If this is wanted to
				// stay in order, then it should be rewritten.
				(e.Urls)[index] = (e.Urls)[len(e.Urls)-1]
				e.Urls = (e.Urls)[:len(e.Urls)-1]
			}

			if (display == "tags") || (edit) {
				a.switchToEditList(true)
			} else {
				a.switchToNewFieldsList(true)
			}

		})
	}
}

func (a *passApp) blankCopen(i int) {
	letter := newCharIterator()
	a.list.Clear()
	e := a.entries[i]

	a.list.AddItem("leave /copen "+strconv.Itoa(i), "(takes you back to /home)", rune(letter), func() {
		clipboard.WriteAll("banana")
		a.switchToHome()
	})
	letter = increment(letter)
	a.list.AddItem("name:", e.Name, rune(letter), func() {
		clipboard.WriteAll(e.Name)
	})
	if e.Tags != "" {
		letter = increment(letter)
		a.list.AddItem("tags:", e.Tags, rune(letter), func() {
			clipboard.WriteAll(e.Tags)
		})
	}
	for _, u := range e.Urls {
		letter = increment(letter)
		a.list.AddItem("url:", u, rune(letter), func() {
			clipboard.WriteAll(u)
		})
	}
	for _, u := range e.Usernames {
		letter = increment(letter)
		a.list.AddItem(u.DisplayName+":", u.Value, rune(letter), func() {
			clipboard.WriteAll(u.Value)
		})
	}
	for _, p := range e.Passwords {
		letter = increment(letter)
		a.list.AddItem(p.DisplayName+":", "[black:black]"+p.Value, rune(letter), func() {
			clipboard.WriteAll(p.Value)
		})
	}
	for _, sq := range e.SecurityQ {
		letter = increment(letter)
		a.list.AddItem(sq.DisplayName+":", "[black:black]"+sq.Value, rune(letter), func() {
			clipboard.WriteAll(sq.Value)
		})
	}
	for _, n := range e.Notes {
		if n != "" {
			letter = increment(letter)
			a.list.AddItem("note:", n, rune(letter), func() {
				clipboard.WriteAll(n)
			})
		}
	}
	letter = increment(letter)
	a.list.AddItem("in circulation:", strconv.FormatBool(e.Circulate), rune(letter), func() {
		clipboard.WriteAll(strconv.FormatBool(e.Circulate))
	})
	if !e.Modified.IsZero() {
		letter = increment(letter)
		a.list.AddItem("date last modified:", fmt.Sprint(e.Modified.Date()), rune(letter), func() {
			clipboard.WriteAll(fmt.Sprint(e.Modified.Date()))
		})
	}
	if !e.Opened.IsZero() {
		letter = increment(letter)
		a.list.AddItem("date last opened:", fmt.Sprint(e.Opened.Date()), rune(letter), func() {
			clipboard.WriteAll(fmt.Sprint(e.Opened.Date()))
		})
	}
	if !e.Created.IsZero() {
		letter = increment(letter)
		a.list.AddItem("date created:", fmt.Sprint(e.Created.Date()), rune(letter), func() {
			clipboard.WriteAll(fmt.Sprint(e.Created.Date()))
		})
	}
	a.entries[i].Opened = time.Now()
	a.writeFileErrNone()
}

// Action is either going to be "pick", "picc", or "flist str". This is
// done to print out the action and send the function to the correct place.
func (a *passApp) blankPicklist(action string, indexes []int) {
	// if /flist str and str is really long:
	if len([]rune(action)) > (56 + width) {
		action = action[:(53+width)] + "..."
	}

	a.infoText.SetText(getInfoText(action))
	letter := newCharIterator()
	a.list.Clear()
	a.list.AddItem("leave "+action, "(takes you back to /home)", rune(letter), func() {
		a.switchToHome()
	})
	for _, i := range indexes {
		// in circulation or in /flist str
		if (a.entries[i].Circulate) || (len([]rune(action)) > 5) {
			letter = increment(letter)

			var title string

			if !a.entries[i].Circulate {
				title = "(rem) "
			}
			title += "[" + strconv.Itoa(i) + "] " + a.entries[i].Name

			a.list.AddItem(title, "tags: "+a.entries[i].Tags, rune(letter), func() {
				if action == "pick" { // to transfer to /open #
					a.app.EnableMouse(false)
					a.boxPages.SwitchToPage("open")
					a.app.SetFocus(a.input)
					a.text.SetText(blankOpen(i, a.entries))
					a.infoText.SetText(getInfoText("open"))
					a.writeFileErrNone()
				} else { // to transfer to /copen # (for both /picc and /flist)
					a.app.SetFocus(a.list)
					a.app.EnableMouse(false)
					a.leftPages.SwitchToPage("copen")
					a.infoText.SetText(getInfoText("copen"))
					a.blankCopen(i)
				}
			})
		}
	}
}

func (a *passApp) passActions(key tcell.Key) {
	// guarantee only enter (13) or tab (9) can be counted
	if (key != 13) && (key != 9) {
		return
	}

	passInputed := a.input.GetText()
	a.input.SetText("")

	if (passInputed == "quit") || (passInputed == "q" ||
		(passInputed == "\\quit") || (passInputed == "\\q")) {
		a.app.Stop()
	}
	a.passwordPages.SwitchToPage("passBox")
	var keyErr string

	a.ciphBlock, keyErr = encrypt.KeyGeneration(passInputed)

	if keyErr != "" {
		a.passwordPages.SwitchToPage("passErr")
		a.text.SetText(keyErr)
		a.input.SetText("")
		return
	}
	readErr := encrypt.ReadFromFile(&a.entries, a.ciphBlock)

	if readErr != "" {
		a.passwordPages.SwitchToPage("passErr")
		a.text.SetText(readErr)
		a.input.SetText("")
		return
	}

	a.updateTerminalSize()
	// set the command line input for the password manager
	a.input.
		SetLabel("input: ").
		SetFieldWidth(59 + width).
		SetPlaceholder("psst look to the right for actions").
		SetDoneFunc(a.commandLineActions).
		SetMaskCharacter(0)
	a.leftPages.SwitchToPage("commandLine")
	a.passPages.SwitchToPage("passManager")
	a.switchToHome()
}

// First, 'inputted' is sanitized and checked to make sure it follows
// conventions. Then, a page and focus is swapped and an action is called.
func (a *passApp) commandLineActions(key tcell.Key) {
	// guarantee only enter (13) or tab (9) can be used
	if (key != 13) && (key != 9) {
		return
	}

	if a.updateTerminalSize() {
		a.input.SetFieldWidth(59 + width)
	}
	a.app.EnableMouse(true)
	a.infoText.ScrollToBeginning()
	a.switchToHome()
	inputed := a.input.GetText()
	a.input.SetText("")
	inputedArr := strings.Split(inputed, " ")
	action := inputedArr[0]

	compIndSelectOne := -1
	compIndSelectTwo := -1

	// Three+ of the commands you need this, have it to be updated if you
	// add new entries
	listAllIndexes := make([]int, len(a.entries))
	for i := 0; i < len(a.entries); i++ {
		listAllIndexes[i] = i
	}

	if len([]rune(action)) > 1 {
		if rune(action[0]) == '/' {
			action = action[1:]
		}
	}

	// if it is one of the actions with extra checks, change it to be its
	// longer name. Therefore, less if statements are needed.
	if len([]rune(action)) < 5 {
		switch action {
		case "o":
			action = "open"
		case "e":
			action = "edit"
		case "c":
			action = "copen"
		case "co":
			action = "copy"
		case "com":
			action = "comp"
		case "f":
			action = "find"
		case "fl":
			action = "flist"
		}
	}

	// The following is a check for the commands that take in a number. Is
	// there a second thing? is it a number? is it a valid number?
	if (action == "open") || (action == "edit") || (action == "copen") || (action == "copy") {
		a.indexSelected = -1 // Sets it here to remove any previous doings

		if len(inputedArr) < 2 { // if there is no number written
			a.switchToError(" To " + action[1:] + " an entry you must write " + action + " and then a number.\n Ex: \n\t/" + action + " 3")
			return
		}
		intTranslated, intErr := strconv.Atoi(inputedArr[1])

		if intErr != nil { // if what passed in is not a number
			a.switchToError(" Make sure to use " + action + " by writing a number!\n Ex: \n\t/" + action + " 3")
			return
		}
		if (intTranslated >= len(a.entries)) || (intTranslated < 0) { // if the number passed in isn't an index
			a.switchToError(" The number you entered does not correspond to an entry.\n Do /list to see the entries (and their numbers) that exist.")
			return
		}
		a.indexSelected = intTranslated

	} else if action == "comp" {
		if len(inputedArr) < 3 {
			a.switchToError(" You must specify which two entries you would like to /comp.\n Ex: \n\t /comp 3 4")
			return
		}
		compOneInt, compOneErr := strconv.Atoi(inputedArr[1])
		compTwoInt, compTwoErr := strconv.Atoi(inputedArr[2])

		if (compOneErr != nil) || (compTwoErr != nil) {
			a.switchToError(" Make sure to only use /comp by writing a number! \n Ex: \n\t /comp 3 4")
			return
		}
		if compOneInt == compTwoInt {
			a.switchToError(" The entries you tried to /comp are the same.\n Therefore, all the passwords would be the same! \n Do /list to see the entries (and their numbers that exist)")
			return
		}
		if ((compOneInt >= len(a.entries)) || (compTwoInt >= len(a.entries))) || ((compOneInt < 0) || (compTwoInt < 0)) {
			a.switchToError(" The number you entered does not correspond to an entry.\n Do /list to see the entries (and their numbers) that exist.")
			return
		}
		compIndSelectOne = compOneInt
		compIndSelectTwo = compTwoInt

	} else if (action == "find") || (action == "flist") {
		// old error message: "To find entries you must write /find and then characters. \n With a space after /find. \n Ex: \n\t /find bank" <-- is specifying the space better?
		if (len(inputedArr) < 2) || (inputedArr[1] == " ") || (inputedArr[1] == "") {
			a.switchToError(" To find entries you must write " + action + " and then characters. \n Ex: \n\t/" + action + " bank")
			return
		}
	}
	switch action {
	case "home", "h":
		a.switchToHome()
	case "quit", "q":
		a.app.Stop()
	case "list", "l":
		text := listEntries(a.entries, listAllIndexes, false)
		a.text.SetText(" /list \n -----")
		a.listText.SetText(text).ScrollToBeginning()
		a.infoText.SetText(getInfoText("list"))
		a.boxPages.SwitchToPage("list")
	case "find":
		title, text := blankFind(a.entries, inputedArr[1])
		a.text.SetText(title)
		a.listText.SetText(text).ScrollToBeginning()
		a.boxPages.SwitchToPage("list")
		a.infoText.SetText(getInfoText("find"))
	case "test", "t":
		// if /test # then add # entries to the entries list
		if len(inputedArr) > 1 {
			intTranslated, intErr := strconv.Atoi(inputedArr[1])
			if intErr == nil {
				for i := 0; i < intTranslated; i++ {
					a.entries = append(a.entries, encrypt.Entry{
						Name:      "testing testing-123456789",
						Tags:      "This was automatically added!",
						Circulate: true})
				}
				// write to file, swap to error page if failed:
				a.writeFileErrNone()
				return
			} else {
				a.switchToError(" To add entries using /test you must write a number!\n Ex: \n\t \\test 3")
				return
			}
		}
		a.text.SetText(testAllFields(a.entries))
		a.boxPages.SwitchToPage("test")
	case "new", "n":
		a.app.EnableMouse(false)
		a.infoText.SetText(getInfoText("newEntry"))
		a.tempEntry = encrypt.Entry{}
		a.blankNewEntry(a.tempEntry)
		a.app.SetFocus(a.newEntryForm)
		a.leftPages.SwitchToPage("newEntry")
	case "help", "he":
		a.text.SetText(helpText())
		a.text.ScrollToBeginning()
		a.boxPages.SwitchToPage("help")
	case "open":
		a.infoText.SetText(getInfoText("open"))
		a.app.EnableMouse(false)
		a.boxPages.SwitchToPage("open")
		a.text.SetText(blankOpen(a.indexSelected, a.entries))
		a.text.ScrollToBeginning()
		a.writeFileErrNone()
	case "copen":
		a.infoText.SetText(getInfoText("copen"))
		a.app.SetFocus(a.list)
		a.app.EnableMouse(false)
		a.leftPages.SwitchToPage("copen")
		// needs to be at the end, because writeErr is called from it
		a.blankCopen(a.indexSelected)
	case "edit":
		a.tempEntry = encrypt.Entry{}
		a.switchToEditList(false)
	case "pick", "pk":
		a.blankPicklist("pick", listAllIndexes)
		a.app.SetFocus(a.list)
		a.leftPages.SwitchToPage("pick")
	case "copy":
		a.tempEntry = encrypt.Entry{}
		a.infoText.SetText(getInfoText("newEntry"))
		a.app.EnableMouse(false)
		a.blankNewEntry(a.entries[a.indexSelected])
		a.app.SetFocus(a.newEntryForm)
		a.leftPages.SwitchToPage("newEntry")
	case "picc", "p":
		a.blankPicklist("picc", listAllIndexes)
		a.app.SetFocus(a.list)
		a.leftPages.SwitchToPage("pick")
	case "flist":
		indexesFound := findIndexes(a.entries, inputedArr[1])
		a.blankPicklist("flist "+inputedArr[1], indexesFound)
		a.app.SetFocus(a.list)
		a.leftPages.SwitchToPage("pick")
	case "comp":
		a.text.SetText(blankComp(compIndSelectOne, compIndSelectTwo, a.entries))
		a.boxPages.SwitchToPage("comp")
	case "reused", "r":
		a.text.SetText(" /reused\n -------\n The following are the passwords and answers reused:\n\n" + reusedAll(a.entries))
		a.boxPages.SwitchToPage("reused")
	default:
		a.switchToError(" That input doesn't match a command! \n Look to the right right to see the possible commands. \n Make sure to spell it correctly!")
	}
}

// Input: any primitive (widget used from tview)
// Return: grid fitted around the primitive to add a border
func grider(prim tview.Primitive) *tview.Grid {
	grid := tview.NewGrid().SetBorders(true)
	grid.AddItem(prim, 0, 0, 1, 1, 0, 0, false)
	return grid
}

// Input: string detailing what the primitives are, like "newEntry" or "open"
// and any number of primitives that belong in the same flex
// Return: new flex situated according to the action string with the inputted
// primitives.
func newFlex(action string, prims ...tview.Primitive) *tview.Flex {
	// to map the action type to the necessary prims length
	actionLen := map[string]int{
		"list":          2,
		"error":         2,
		"newEntry":      2,
		"newField":      1,
		"newNote":       1,
		"newEditField":  1,
		"editEditField": 1,
		"editFieldStr":  1,
		"editDelete":    2,
		"passErr":       2,
		"password":      2,
		"main":          2,
		"leftPages":     2,
	}

	if actionLen[action] < len(prims) {
		return tview.NewFlex().AddItem(tview.NewBox().SetTitle("error!"), 0, 1, false)
	}

	switch action {
	case "list", "error", "passErr":
		return newFlexRow().
			AddItem(prims[0], 2, 0, false).
			AddItem(prims[1], 0, 1, false)
	case "newEntry":
		// Positioning the form for the name with the buttons at the top with
		// the list of already added items at the bottom
		return newFlexRow().
			AddItem(prims[0], 9, 0, false).
			AddItem(prims[1], 0, 1, false)
	case "newField":
		// creating a new field in /edit or /new
		return tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(newFlexRow().
				AddItem(nil, 0, 2, false).
				AddItem(grider(prims[0]), 11, 0, false).
				AddItem(nil, 0, 1, false), 0, 4, false)
	case "newNote":
		// Editing new or existing notes in /new or /edit
		return tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(newFlexRow().
				AddItem(nil, 0, 2, false).
				AddItem(grider(prims[0]), 17, 0, false).
				AddItem(nil, 0, 1, false), 0, 5, false)
	case "newEditField":
		// Editing or creating a field in /new, positioned further down
		return tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(newFlexRow().
				AddItem(nil, 0, 3, false).
				AddItem(grider(prims[0]), 9, 0, false).
				AddItem(nil, 0, 2, false), 0, 4, false)
	case "editEditField":
		// Editing or creating a complex field (username, password, security
		// question) in /edit, positioned further up
		return tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(newFlexRow().
				AddItem(nil, 0, 1, false).
				AddItem(grider(prims[0]), 9, 0, false).
				AddItem(nil, 0, 1, false), 0, 4, false)
	case "editFieldStr":
		// Editing a single string filed (URL or tags) in /edit
		return tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(newFlexRow().
				AddItem(nil, 0, 3, false).
				AddItem(grider(prims[0]), 7, 0, false).
				AddItem(nil, 0, 2, false), 0, 4, false)
	case "editDelete":
		// Situate the text and the form together for deleting an entry
		editDeleteFlex := newFlexRow().
			AddItem(prims[0], 0, 1, false).
			AddItem(prims[1], 0, 1, false)

		// Put them in the middle
		return tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(newFlexRow().
				AddItem(nil, 0, 2, false).
				AddItem(grider(editDeleteFlex), 0, 2, false).
				AddItem(nil, 0, 2, false), 0, 1, false).
			AddItem(nil, 0, 1, false)
	case "password":
		// password page with the input and empty box
		return tview.NewFlex().
			AddItem(newFlexRow().
				AddItem(prims[0], 0, 1, false).
				AddItem(grider(prims[1]), 3, 0, false), 0, 1, false)
	case "leftPages":
		// Left side of pass manager, pages and input
		return newFlexRow().
			AddItem(prims[0], 0, 1, false).
			AddItem(grider(prims[1]), 3, 0, false)
	case "main":
		// Left and right sides of the main password manager
		return tview.NewFlex().
			AddItem(prims[0], 0, 1, false).
			AddItem(grider(prims[1]), 15, 0, false)
	}

	return tview.NewFlex().AddItem(tview.NewBox().SetTitle("error!"), 0, 1, false)
}

// Return: new flex with its direction set to rows (the default is
// columns)
func newFlexRow() *tview.Flex {
	return tview.NewFlex().SetDirection(tview.FlexRow)
}

// Return: new list with colors set and the setting of it only being
// highlighted when the list is in focus turned on.
func newList() *tview.List {
	return tview.NewList().
		SetSelectedFocusOnly(true).
		SetSecondaryTextColor(blue).
		SetShortcutColor(lavender)
}

// Return: new input field with colors set
func newInputField() *tview.InputField {
	return tview.NewInputField().
		SetFieldBackgroundColor(blue).
		SetLabelColor(lavender).
		SetPlaceholderStyle(tcell.Style{}.Background(blue).Foreground(white))
}

// Retrurn: new TextView that has scrolling on
func newScrollableTextView() *tview.TextView {
	return tview.NewTextView().SetScrollable(true)
}

// Return: new form that has the colors defined
func newForm() *tview.Form {
	return tview.NewForm().
		SetButtonBackgroundColor(blue).
		SetFieldBackgroundColor(blue).
		SetLabelColor(lavender)
}

// Input: string detailing what page going to, like "newEntry" or "open"
// Return: string to put in the right side info area
func getInfoText(action string) string {

	// dealing with the case of /flist str
	spaceSplit := strings.Split(action, " ")
	action = spaceSplit[0]

	switch action {
	case "editEntry":
		return " /edit \n ----- \n to move: \n -tab \n -back tab \n -arrows keys\n -scroll\n\n to select: \n -return \n\n to leave: \n -esc key\n -a"
	case "editField":
		return " /edit \n ----- \n to move: \n -tab \n -back tab\n\n to select: \n -return \n\n must name \n field to \n save it \n\n press quit \n to leave"
	case "newEntry", "newField", "newFieldsAdded":
		fieldEntry := "entry"
		backToTop := ""
		if action == "newField" {
			fieldEntry = "field"
		} else if action == "newFieldsAdded" {
			backToTop = "\n\n press esc\n or a to go\n back to top"
		}

		return " /new \n ---- \n to move: \n -tab \n -back tab \n\n to select: \n -return \n\n must name \n " + fieldEntry + " to \n save it \n\n press quit \n to leave" + backToTop
	case "open":
		return " /open\n -----\n to edit:\n  /edit #\n to copy:\n  /copen # \n\n /home\n /help\n /quit\n\n /new\n /copy #\n\n /find str\n /flist str\n\n /list\n /pick\n /picc\n\n /comp # #\n /reused"
	case "copen":
		return " /copen \n ------\n to edit: \n /edit # \n\n to move: \n -tab \n -back tab \n -arrows keys\n -scroll\n\n to select:\n -return\n\n to leave:\n -esc key\n -a"
	case "pick", "picc", "flist":
		actionNewLines := " /" + action + "\n " +
			strings.Repeat("-", len([]rune(action)))

		return actionNewLines + "\n to move: \n -tab \n -back tab \n -arrows keys\n\n to select: \n -return\n -click\n\n to leave: \n -esc key\n -a"
	default:
		return " commands\n -------- \n /home\n /help\n /quit\n\n /open #\n /copen #\n\n /new\n /copy #\n\n /edit #\n\n /find str\n /flist str\n\n /list\n /pick\n /picc\n\n /comp # #\n /reused"
	}
}

// Return: text box with the help text inside. The help text is resized based
// on the width of the terminal.
func helpText() string {
	text0 := ` /help
 -----
 In order to quit, press control+c or type /quit.

 # means entry number and str means some text.
 	example of /open # is: /open 3
 	example of /find str is: /find library`

	text1 := `

 Sometimes you can use the mouse to click, but sometimes you
 can't. This is because when you can use the mouse to click, you
 can't use it to select and copy text. You can only scroll your
 mouse when you can click, not when you can select text. You can
 scroll with your mouse on this page.

 Use /open # to view an entry. Passwords and security question
 values will be blotted out but they can be highlighted and then
 copied. You can also use /copen # to view it in a list form, and
 select a field which will be copied to your clipboard. If you
 want to edit an entry you must do /edit #. You cannot scroll in
 /open #, so if you have too many fields use /copen #.

 Use /new to make a new entry. You must give your entry a name to
 save it. You must also give each field a display name in order to
 save them. You can also write in notes and you can edit the
 fields you've already added. You don't need to write tags, but
 they can be helpful in searching for entries using /find str.
 You can also do /copy # which is the same as doing /new but info
 is already filled out from entry #. A new entry is not saved 
 until you click the save button and you are moved away from /new.

 Use /edit # to edit an existing entry. You can edit the fields
 already there, add new ones, remove it from circulation, or
 delete it. While there is a delete button, it is recommended that
 you remove it from circulation instead. When that is done, it
 won't show up in /list or /pick. All of the other commands (such
 as /open, /edit, etc.) will still work on it. Edits are saved as
 soon as you click save on each specific field.

 Use /find str to search for entries. /find str will return all of
 the entries that contain str in the name, tags, or url. For
 example, if my tag says "gmail" and I search /find mail, then
 that entry will show up as the string "mail" is within "gmail".
 In both /find str and /list, the resulting entries may not show
 their full name for space. Use /flist str to see a list of
 entries with that str, when clicked they are /copen.

 Use /list or /pick to view the list of entries. /list will
 display them all with their numbers and you can then type /open #
 or /copen #. /pick will display a list of the entries and you can
 click on one of them to open it. You can do /picc to copen it.

 Use /reused to see a list of passwords or answers to
 security questions that are reused in any entries.

 Use /comp # # to compare the passwords and question answers
 between two entries to see if there are any duplicates.

 When in /edit, /flist, /pick, or /picc you can press esc to
 go back to home. This can be more efficient than scrolling to
 select the item to leave.

 When making or editing notes, you can write [black] to have it
 blotted out. Make sure at the end of the line you write [white]
 at the end in order for the other lines in notes to show up!

 The colors of this project, lavender and light blue, do not
 comply with WCAG AAA standards. To have a higher contrast,
 uncomment the lines before the variables lavender and blue are
 defined, comment out the current color definitions. These
 variables are near the beginning of the pass.go file.

 Here is a list of shortcuts for the commands which will do the
 same thing as the normal commands:`

	text2 := `
  /home → /h
  /help → /he
  /quit → /q
  /open # → /o #
  /copen # → /c #
  /new → /n
  /copy # → /co #
  /edit # → /e #
  /find str → /f str
  /flist str → /fl str
  /list → /l
  /pick → /pk
  /picc → /p
  /comp # # → /com # #
  /reused → /r`

	text3 := `

 The slash / is optional, so "open 0" or "o 0" works like
 "/open 0".

 If you want to change your password or the password parameters,
 run changeKey.go. 

 More info about the project is on the README at
 https://github.com/ksharnoff/pass.`

	var b strings.Builder

	// text0 has some indentation that would mess up the splitting and is less
	// than 50 characters across so should show up on all screens.
	b.WriteString(text0)

	// If not at default size, move the newlines to fit better.
	// This is more important when at a smaller width.
	if width != 0 {
		col := 66 + width

		split1 := strings.Split(text1, "\n\n")
		b.WriteString(resizeHelpTextChunk(split1, col))

		b.WriteString(text2)

		split3 := strings.Split(text3, "\n\n")
		b.WriteString(resizeHelpTextChunk(split3, col))

	} else {
		b.WriteString(text1)
		b.WriteString(text2)
		b.WriteString(text3)
	}

	return b.String()
}

// Input: slice of strings where each string represents a distinct paragraph
// whose lines are broken at every 66 characters with a new line and an int
// named col which is the difference from 66 characters that the paragraphs
// should be broken into
// Return: string of all the paragraphs formatted for the different sized
// terminal and combined all together
func resizeHelpTextChunk(text []string, col int) string {
	length := len(text)

	var b strings.Builder

	for i := 0; i < length; i++ {
		paragraph := strings.Join(strings.Split(text[i], "\n"), "")

		for len([]rune(paragraph)) > col {
			ind := strings.LastIndex(paragraph[:col], " ")

			// if no space found
			if ind < 0 {
				ind = col
			}

			b.WriteString(paragraph[:ind])
			b.WriteByte(byte('\n'))
			paragraph = paragraph[ind:]
		}
		b.WriteString(paragraph)

		// if not at second to last
		if length-i != 1 {
			b.WriteByte(byte('\n'))
			b.WriteByte(byte('\n'))
		}
	}
	return b.String()
}

// Used for letters in the lists in /pic(k/c), /copen, /flist.
// Returns an int the value of 'a'
func newCharIterator() int {
	return int('a')
}

// Returns an int the value of the next character in alphabet, or if the
// input was 'z' then it will return 'a'.
func increment(count int) int {
	count++

	if (count - int('a')) > 25 { // if past z, reset to a
		count = int('a')
	}
	return count
}

// Input: index, slice of entries
// Return: string formatted for /open of the entry at the index
// Entries needs to be inputted an not just the single one so that we can modify
// the last opened time and then write it to file.
func blankOpen(i int, entries []encrypt.Entry) string {
	e := entries[i]

	var b strings.Builder

	name := e.Name
	if len([]rune(name)) > 57+width {
		name = name[:57+width] + "..."
	}

	b.WriteString(" [" + strconv.Itoa(i) + "] " + name)
	// print hyphens under the title:
	b.WriteString("\n " + strings.Repeat("-", len([]rune(b.String()))-1))

	if e.Tags != "" {
		b.WriteString("\n tags: " + e.Tags)
	}
	for _, u := range e.Urls {
		b.WriteString("\n url: " + u + "[white]")
	}
	for _, u := range e.Usernames {
		b.WriteString("\n " + u.DisplayName + ": " + u.Value + "[white]")
	}
	for _, p := range e.Passwords {
		b.WriteString("\n " + p.DisplayName + ": [black:black]" + p.Value + "[white]")
	}
	for _, sq := range e.SecurityQ {
		b.WriteString("\n " + sq.DisplayName + ": [black:black]" + sq.Value + "[white]")
	}
	emptyNotes := true
	for _, n := range e.Notes {
		if n != "" {
			emptyNotes = false
			break
		}
	}
	if !emptyNotes {
		col := 62 + width

		b.WriteString("\n notes:")
		blankLines := 0 // Stops printing \n\n\n\n if only text in first line
		for _, n := range e.Notes {
			if n == "" {
				blankLines++
			} else {
				b.WriteString(strings.Repeat("\n", blankLines))

				for len([]rune(n)) > col {
					indexes := regexp.MustCompile(`[^a-zA-Z0-9]`).FindAllStringIndex(n[:col], -1)

					ind := -1

					length := len(indexes)
					if length < 1 {
						ind = col - 1
					} else {
						if len(indexes[length-1]) < 1 {
							ind = col - 1
						} else {
							ind = indexes[length-1][0]
						}
					}

					if ind <= 0 {
						ind = col - 1
					}
					ind++

					b.WriteString("\n\t" + n[:ind])
					n = n[ind:]
				}

				b.WriteString("\n\t" + n)
				blankLines = 0
			}
		}
	}
	b.WriteString("\n\n[white]")
	// Following is info about the entry
	b.WriteString(" in circulation: " + strconv.FormatBool(e.Circulate) + "\n")
	if !e.Modified.IsZero() { // if it's not jan 1, year 1
		b.WriteString(" date last modified: " + fmt.Sprint(e.Modified.Date()) + "\n")
	}
	if !e.Opened.IsZero() { // if it's not jan 1, year 1
		b.WriteString(" date last opened: " + fmt.Sprint(e.Opened.Date()) + "\n")
	}
	if !e.Created.IsZero() { // if it's not jan 1, year 1
		b.WriteString(" date created: " + fmt.Sprint(e.Created.Date()))
	}
	entries[i].Opened = time.Now()
	return b.String()
}

// Inputs: entries slice and search string
// Returns: the action name as a string, "/find str" and the found entries in
// a second string, formatted into columns
func blankFind(entries []encrypt.Entry, str string) (string, string) {
	indexes := findIndexes(entries, str)

	// Trims the /find str for printing, adding a ... if trimmed. It still
	// used the full str in the search.
	if len([]byte(str)) > 59+width {
		str = str[:56+width]
		str += "..."
	}

	titleUnderline := " /find " + str + " \n " + strings.Repeat("-", len([]rune(str))+6)

	if len(indexes) > 0 {
		return titleUnderline, listEntries(entries, indexes, true)
	} else {
		return titleUnderline, " no entries found"
	}
}

// Input: slice of entries and search string
// Returns: slice of ints of all indexes with the search string in the
// name, tags, or URLs.
// This is used in /find and /flist
func findIndexes(entries []encrypt.Entry, str string) []int {
	indexes := []int{}
	str = strings.ToLower(str)
	for i, e := range entries {
		if (strings.Contains(strings.ToLower(e.Name), str)) ||
			(strings.Contains(strings.ToLower(e.Tags), str)) {
			indexes = append(indexes, i)
		} else {
			for _, u := range e.Urls {
				if strings.Contains(strings.ToLower(u), str) {
					indexes = append(indexes, i)
					break
				}
			}
		}
	}
	return indexes
}

// Input: slice of entries, slice of indices (ints) to display, showOld bool
// which is true if entries not in circulation should be shown (in /find), false
// otherwise (like in /list)
// Returns: a string of the entries formatted into three columns.
func listEntries(entries []encrypt.Entry, indexes []int, showOld bool) string {
	printEntries := []encrypt.Entry{}

	if showOld {
		for _, i := range indexes {
			// entries[i] is equivalent to entries[indexes[i]]
			printEntries = append(printEntries, entries[i])
		}
	} else {
		indexesCirculated := []int{}

		for i, _ := range indexes {
			if entries[i].Circulate {
				printEntries = append(printEntries, entries[i])
				indexesCirculated = append(indexesCirculated, i)
			}
		}
		indexes = indexesCirculated
	}

	// floatThird was made to deal with the case of if there are more
	// than 63 entries but the number of entries is not a multiple of
	// 3. It looks weird when it shifts, losing 2 rows from the furthest
	// right column, but that is how the math works out given it's in
	// thirds.
	// 63 entries is the number to go off the screen in a 84x28
	// terminal
	floatThird := float64(len(indexes)) / 3.0

	if floatThird < 21.0+float64(height) {
		floatThird = 21.0 + float64(height)
	} else if floatThird > float64(int(floatThird)) {
		floatThird++
	}

	third := int(floatThird)
	var b strings.Builder

	for i := 0; i < third; i++ {
		if i >= len(indexes) {
			break
		}
		b.WriteString(" " + indexName(indexes[i], entries))

		if len(indexes) > i+third {
			b.WriteString(indexName(indexes[i+third], entries))
		}

		if len(indexes) > i+third+third {
			b.WriteString(indexName(indexes[i+third+third], entries))
		}

		// so it doesn't do it on the last one
		if i != third-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

// Inputs: int index of entry and slice of entries. The reason that the slice
// of entries is inputted and not a single entry is because the logic for the
// indexes is non trivial and would be not great to repeat in order to input the
// exact entry.
// Return: string of a single name formatted for a third of a column in /list or
// /find, where if the user is on the default terminal size it'll be exactly
// " [0] twitterDEMO       "
func indexName(index int, entries []encrypt.Entry) string {
	var b strings.Builder

	b.WriteString("[" + strconv.Itoa(index) + "] ")

	if !entries[index].Circulate {
		b.WriteString("(rem) ")
	}

	b.WriteString(entries[index].Name)
	len := len([]rune(b.String()))

	// Trims it if it's over the character limit
	if len > 21+(width/3) {
		return b.String()[0:21+(width/3)] + " "
	}

	b.WriteString(strings.Repeat(" ", 22+(width/3)-len))
	return b.String()
}

// Input: two indexes, i1 and i2 that are the two entries whose passwords are
// being compared and the slice of all entries
// Return: string to write to the text box for /comp i1 i2 -- with title and
// comparison information.
func blankComp(i1, i2 int, entries []encrypt.Entry) string {
	e1 := entries[i1]
	e2 := entries[i2]

	var b strings.Builder

	b.WriteString(" /comp: " + "[" + strconv.Itoa(i1) + "] " +
		shortenedName(e1.Name) + " and " + "[" + strconv.Itoa(i2) + "] " +
		shortenedName(e2.Name) + "\n ")
	b.WriteString(strings.Repeat("-", len([]rune(b.String()))-3) + "\n\n")

	b.WriteString(compPass(e1, e2))

	return b.String()
}

// Input: two entries, e1 and e2
// Return: string formatted to be sent to the /comp text box about any
// passwords or security questions in common between the two entries.
func compPass(e1, e2 encrypt.Entry) string {
	// reusedPass is a struct
	compMap := make(map[string][]reusedPass)
	name1 := shortenedName(e1.Name)
	name2 := shortenedName(e2.Name)

	// Adding all passwords and securityQs to the map
	for _, p := range e1.Passwords {
		compMap[p.Value] = append(compMap[p.Value],
			reusedPass{displayName: p.DisplayName, entryName: name1})
	}
	for i, s := range e1.SecurityQ {
		compMap[s.Value] = append(compMap[s.Value],
			reusedPass{displayName: "security question " + strconv.Itoa(i), entryName: name1})
	}
	for _, p := range e2.Passwords {
		compMap[p.Value] = append(compMap[p.Value],
			reusedPass{displayName: p.DisplayName, entryName: name2})
	}
	for i, s := range e2.SecurityQ {
		compMap[s.Value] = append(compMap[s.Value],
			reusedPass{displayName: "security question " + strconv.Itoa(i), entryName: name2})
	}
	var b strings.Builder

	// Going through the map and looking at duplicates
	for _, reusedStruct := range compMap {
		// if same pass twice, most common
		if len(reusedStruct) == 2 {
			b.WriteString(" " + reusedStruct[0].entryName + "'s " +
				reusedStruct[0].displayName + " = " + reusedStruct[1].entryName +
				"'s " + reusedStruct[1].displayName + "\n")

			// if more than twice
		} else if len(reusedStruct) > 2 {
			for i, r := range reusedStruct {
				b.WriteString(" " + r.entryName + "'s " + r.displayName)

				// add extra space when not in last time through
				if (i + 1) < len(reusedStruct) {
					b.WriteString(" \n")
				} else {
					b.WriteString("\n")
				}
			}
		}
	}
	if b.Len() < 1 {
		if (len(e1.Passwords) < 1) && (len(e1.SecurityQ) < 1) {
			b.WriteString(" " + name1 + " has no passwords or security questions" + "\n")
		}
		if len(e2.Passwords) < 1 && (len(e2.SecurityQ) < 1) {
			b.WriteString(" " + name2 + " has no passwords or security questions" + "\n")
		}
		b.WriteString("\n Therefore, there are no passwords in common!")
	}
	return b.String()
}

// Input: slice of entries
// Output: formatted string containing any reused passwords or security
// question answers between all entries in the slice
func reusedAll(entries []encrypt.Entry) string {
	var b strings.Builder

	reused := make(map[string][]reusedPass) // reusedPass is a struct

	for i, e := range entries {
		name := shortenedName(e.Name)

		for _, p := range e.Passwords {
			reused[p.Value] = append(reused[p.Value],
				reusedPass{
					displayName: p.DisplayName,
					entryName:   name,
					entryIndex:  i,
				})
		}
		for iSq, s := range e.SecurityQ {
			reused[s.Value] = append(reused[s.Value],
				reusedPass{
					displayName: "security question " + strconv.Itoa(iSq),
					entryName:   name,
					entryIndex:  i,
				})
		}
	}
	for pass, reusedStruct := range reused {

		// if there's more than one entry in the slice of entries for password
		if len(reusedStruct) > 1 {
			b.WriteString(" [darkslategray]" + pass + "[white]:\n")

			for _, r := range reusedStruct {
				b.WriteString(" [" + strconv.Itoa(r.entryIndex) + "] " +
					r.entryName + "'s " + r.displayName + "\n")
			}

			b.WriteString("\n")
		}
	}

	if b.Len() < 1 {
		return " There are no reused passwords anywhere!?\n Good job!"
	}

	// gets rid of the last \n\n and return
	printStr := b.String()
	printStr = printStr[:len([]rune(printStr))-2]
	return printStr
}

// Input: full name of entry
// Return: shortened name of an entry, for use in /reused and /comp
// so that they never go off the side of the screen
func shortenedName(name string) string {
	if len([]rune(name)) > 22+(width/3) {
		return name[:22+(width/3)]
	}
	return name
}

// Input: slice of entries
// Return: all of the entries formatted into one string, used for testing
func testAllFields(entries []encrypt.Entry) string {
	var b strings.Builder
	b.WriteString(" Test of all fields that are known:")

	for _, e := range entries {
		b.WriteString("\n\n" + fmt.Sprint(e))
	}
	return b.String()
}

// Updates the global variables width, height that store the difference from
// the terminal size of the user and the terminal size designed for.
// If there is an error in getting the terminal size, then it is assumed that
// the user has the same size as original terminal (84x28).
func (a passApp) updateTerminalSize() bool {
	oldWidth := width
	oldHeight := height

	termWidth, termHeight, terminalSizeErr := term.GetSize(int(os.Stdin.Fd()))

	if terminalSizeErr != nil {
		width = 0
		height = 0
		return false
	}

	width = termWidth - 84
	height = termHeight - 28

	// Through experimenting, if the width is below 75 then you cannot see and
	// access all the buttons in /new. The height should be at least 17 to see
	// the list of fields already added in /new.
	if termWidth < 75 || termHeight < 14 {
		if a.app != nil {
			a.app.Stop()
		}
		fmt.Println("In order to see all necessary elements of the password manager, please set your terminal to a minimum height of 14 characters and a minimum width of 75 characters.")
		fmt.Println("The ideal dimensions that this was designed for is a height of 28 and a width of 84.")
		os.Exit(1)
	}

	if (oldWidth != width) || (oldHeight != height) {
		return true
	}
	return false
}
