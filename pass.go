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

// The following structs are for combination primitives.
type textFormFlex struct {
	text *tview.TextView
	form *tview.Form
	flex *tview.Flex
}
type twoTextFlex struct {
	title *tview.TextView
	text  *tview.TextView
	flex  *tview.Flex
}

// This is in the map in /reused and /comp. It has the fields associated
// with a certain password.
type reusedPass struct {
	displayName string
	entryName   string
	entryIndex  int
}

// The terminal width and height represent the difference from the size that
// this was designed for (84x28). It is a global variable to avoid having to
// pass it to the many functions that make text columns and widgets.
var width = 0
var height = 0

func main() {
	// You can uncomment out the next two lines and comment out the default
	// colors in order for it to have a higher contrast that complies with WCAG
	// AAA. lavender is label and shortcut names. blue is secondary text in
	// lists, buttons in forms, and the input field color.

	// lavender := tcell.GetColor("white") // uncomment for higher contrast
	// blue := tcell.NewRGBColor(0, 0, 255) // uncomment for higher contrast
	lavender := tcell.NewRGBColor(149, 136, 204) // comment for higher contrast
	blue := tcell.NewRGBColor(106, 139, 166)     // comment for higher contrast

	white := tcell.GetColor("white")

	// This is to set the background colors of the text in the input lines.
	// There was no function for it, so it had to be done using tcell.Style.
	placeholdStyle := tcell.Style{}.Background(blue).Foreground(white)

	// This is the input line used for navigation and commands for pass.
	commandLineInput := tview.NewInputField().
		SetLabel("input: ").SetFieldWidth(59 + width).
		SetFieldBackgroundColor(blue).
		SetLabelColor(lavender).
		SetPlaceholderStyle(placeholdStyle)

	// This is called when you can type in the commandLine. It changes the
	// placeholder text to say to look at the right, where there is infoText.
	canTypeCommandLinePlaceholder := func() {
		commandLineInput.SetPlaceholder("psst look to the right")
	}
	// This function is called when you can't type in the commandLine
	cantTypeCommandLinePlaceholder := func() {
		commandLineInput.SetPlaceholder("psst you can't type here right now")
	}

	updateTerminalSize()

	app := tview.NewApplication()

	// Pages is the pages set up for the left box in pass
	pages := tview.NewPages()

	// This is the text box on the right that contains information that changes
	// depending on what the user is doing.
	infoText := tview.NewTextView().SetScrollable(true).SetWrap(false)

	// Switches to home, rights everything again.
	switchToHome := func() {
		pages.SwitchToPage("home")
		app.SetFocus(commandLineInput)
		infoText.SetText(" commands\n -------- \n /home\n /help\n /quit\n\n /open #\n /copen #\n\n /new\n /copy #\n\n /edit #\n\n /find str\n /flist str\n\n /list\n /pick\n /picc\n\n /comp # #\n /reused")
		canTypeCommandLinePlaceholder()
		app.EnableMouse(true)
	}

	// This is where the errors are written to. error.title stays the same for
	// all errors.
	error := twoTextFlex{
		title: tview.NewTextView().SetText(" Uh oh! There was an error:"),
		text:  tview.NewTextView().SetScrollable(true),
		flex:  tview.NewFlex(),
	}

	// Switches to the error page, sets error.text to the inputted err.
	switchToError := func(err string) {
		error.text.SetText(err)
		pages.SwitchToPage("err")
	}

	// This slice is all the passwords and info. The following entry names will
	// only be seen if the manager opens without loading a file.
	entries := []encrypt.Entry{
		encrypt.Entry{Name: "QUIT NOW, DANGER", Circulate: true},
		encrypt.Entry{Name: "SOMETHING'S VERY", Circulate: true},
		encrypt.Entry{Name: "BROKEN. QUIT!", Circulate: true},
		encrypt.Entry{Name: "YOUR DATA IS NOT FOUND.", Circulate: true},
	}

	// This is the cipher block generated with the key to encrypt and decrypt.
	// Normally its the key that gets passed around not the cipher block, but
	// I chose to do it this way.
	var ciphBlock cipher.Block

	// This tries to write to file, if it fails, it switches to the error page
	// and returns false. The reason for returning false is so that when used
	// else where it doesn't switch to error page and then immediately switch
	// else where so it can't be seen.
	writeFileErr := func() bool {
		writeErr := encrypt.WriteToFile(entries, ciphBlock)
		if writeErr != "" {
			switchToError(writeErr)
			return false
		}
		return true
	}

	// Has to be initialized ahead of time, comments are later
	blankEditList := func(i int) {}

	// This is the list for /edit and its function for making it
	editList := tview.NewList().
		SetSelectedFocusOnly(true).
		SetSecondaryTextColor(blue).
		SetShortcutColor(lavender).
		SetDoneFunc(switchToHome) // Needs to happen after switchToHome is filled

	// This is the variable for what entry is selected. It is set in
	// commandLineActions and used for a function below.
	indexSelected := -1

	// This is whats written in infoText during /edit
	editInfo := " /edit \n ----- \n to move: \n -tab \n -back tab \n -arrows keys\n -scroll\n\n to select: \n -return \n\n to leave: \n -esc key\n -a"

	// Switches back to the edit list after editing a specific field. It remakes
	// the list each time and uses indexSelected. It takes in a bool to know
	// whether or not to write to file the changes, as well as whether or not
	// to update the last modified time.
	switchToEditList := func(modified bool) {
		if writeFileErr() {
			if modified {
				entries[indexSelected].Modified = time.Now()
			}
			blankEditList(indexSelected)
			pages.SwitchToPage("edit")
			app.SetFocus(editList)
			infoText.SetText(editInfo)
		}
	}

	// Has to be initialized ahead of time, comments are later
	blankEditFieldForm := func(f *encrypt.Field, fieldArr *[]encrypt.Field, index int, edit bool) {}
	blankEditStringForm := func(display, value string, e *encrypt.Entry, edit bool) {}

	// This is the fields added so far list and its function, used in /new.
	newFieldsAddedList := tview.NewList().
		SetSelectedFocusOnly(true).
		SetSecondaryTextColor(blue).
		SetShortcutColor(lavender)

	// form and the flex for /new. The flex puts the list of the
	// entries added with the form of /new. The struct being used has text, but
	// this does not use a flex. Also there is the function for it.
	newEntry := textFormFlex{
		form: tview.NewForm().
			SetButtonBackgroundColor(blue).
			SetFieldBackgroundColor(blue).
			SetLabelColor(lavender),
		flex: tview.NewFlex(),
	}

	// This is the form for editing a specific field and the flexes.
	editFieldForm := tview.NewForm().
		SetButtonBackgroundColor(blue).
		SetFieldBackgroundColor(blue).
		SetLabelColor(lavender)

	// These are temporary and used when someone is making a new entry, a new
	// field, or editing an existing entry.
	tempEntry := encrypt.Entry{}
	tempField := encrypt.Field{}

	// This just uses tempEntry to get the fields, this works because
	// tempEntry is defined to be equal to entry e in blankNewEntry when called
	// after /copy.
	blankFieldsAdded := func() {
		newFieldsAddedList.Clear()
		letter := newCharIterator()

		if newEntry.form.GetButtonIndex("edit field") < 0 { // if there isn't one already
			newEntry.form.
				AddButton("edit field", func() { // Don't change the label name, brakes stuff later.
					app.SetFocus(newFieldsAddedList)
				})
		}
		newFieldsAddedList.
			AddItem("move back to top", "", rune(letter), func() {
				app.SetFocus(newEntry.form)
			})
		for _, u := range tempEntry.Urls {
			letter = increment(letter)
			newFieldsAddedList.AddItem("url:", u, rune(letter), func() {
				blankEditStringForm("url", u, &tempEntry, false)
				pages.ShowPage("editFieldStr")
				app.SetFocus(editFieldForm)
			})
		}
		for i := range tempEntry.Usernames {
			u := &tempEntry.Usernames[i]
			letter = increment(letter)

			newFieldsAddedList.AddItem(u.DisplayName+":", u.Value, rune(letter), func() {
				blankEditFieldForm(u, &tempEntry.Usernames, i, false)
				pages.ShowPage("new-editField")
				app.SetFocus(editFieldForm)
			})
		}
		for i := range tempEntry.Passwords {
			p := &tempEntry.Passwords[i]
			letter = increment(letter)

			newFieldsAddedList.AddItem(p.DisplayName+":", "[black:black]"+p.Value, rune(letter), func() {
				blankEditFieldForm(p, &tempEntry.Passwords, i, false)
				pages.ShowPage("new-editField")
				app.SetFocus(editFieldForm)
			})
		}
		for i := range tempEntry.SecurityQ {
			sq := &tempEntry.SecurityQ[i]
			letter = increment(letter)

			newFieldsAddedList.AddItem(sq.DisplayName+":", "[black:black]"+sq.Value, rune(letter), func() {
				blankEditFieldForm(sq, &tempEntry.SecurityQ, i, false)
				pages.ShowPage("new-editField")
				app.SetFocus(editFieldForm)
			})
		}
	}

	// To be used when each field is edited in /new. It creates the button
	// 'edit fields' after creation of first field in /new. It will appear
	// there already if you are in /copy # and # has fields. If doSwitch is
	// true, then you swap focus to the list of fields already added.
	switchToNewFieldsList := func(doSwitch bool) {
		blankFieldsAdded()
		if (doSwitch) && (newFieldsAddedList.GetItemCount() > 1) {
			pages.SwitchToPage("newEntry")
			app.SetFocus(newFieldsAddedList)
		}

		if newFieldsAddedList.GetItemCount() < 2 { // if all the fields are deleted, then:
			newFieldsAddedList.Clear()
			editFieldIndex := newEntry.form.GetButtonIndex("edit field")
			if editFieldIndex > -1 {
				newEntry.form.RemoveButton(editFieldIndex)
				pages.SwitchToPage("newEntry")
				app.SetFocus(newEntry.form)
			} else {
				switchToError(" For some reason the edit field button wasn't added despite a field later trying to be deleted!\n that's not supposed to happen!")
			}
		}
	}

	// Takes in an extra boolean to know if its from /edit or /new, in order to
	// know where to go back to.
	blankEditFieldForm = func(f *encrypt.Field, fieldArr *[]encrypt.Field, index int, edit bool) {
		editFieldForm.Clear(true)
		tempField.DisplayName = f.DisplayName
		tempField.Value = f.Value

		editFieldForm.
			AddInputField("display name:", tempField.DisplayName, 40+width, nil, func(input string) {
				tempField.DisplayName = input
			}).
			AddInputField("value:", tempField.Value, 40+width, nil, func(input string) {
				tempField.Value = input
			}).
			AddButton("save", func() {
				*f = tempField
				if edit {
					switchToEditList(true) // true meaning it was modified
				} else {
					switchToNewFieldsList(true) // true meaning keep in the list section
				}
			}).
			AddButton("quit", func() {
				if edit {
					switchToEditList(false) // false meaning not modified
				} else {
					switchToNewFieldsList(true) // true meaning keep in list section
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
						switchToEditList(true) // true meaning modified
					} else {
						switchToNewFieldsList(true) // true meaning keep in list section
					}
				} else {
					switchToError(" The slice given to blankEditFieldForm is nil \n and it shouldn't be! or the index is -1 which it also shouldn't be!")
				}
			})
	}

	// The form when adding a new field in /new and /edit
	newFieldForm := tview.NewForm().
		SetButtonBackgroundColor(blue).
		SetFieldBackgroundColor(blue).
		SetLabelColor(lavender)

	// This is the text to be shown on the left in infoText when creating a new
	// entry or a new field.
	newInfo := " /new \n ---- \n to move: \n -tab \n -back tab \n\n to select: \n -return \n\n must name \n entry to \n save it \n\n press quit \n to leave"
	newFieldInfo := " /new \n ---- \n to move: \n -tab \n -back tab \n\n to select: \n -return \n\n must name \n field to \n save it \n\n press quit \n to leave" // only change from this one to the newInfo is field vs. entry

	// Takes in a pointer to tempEntry if in /new. Takes in a pointer to an
	// entry if in /edit.
	blankNewField := func(e *encrypt.Entry) {
		edit := false

		dropDownFields := []string{"url", "username", "password", "security question"}

		// Only adds tags and url as an option to add on if it is in /edit
		if e != &tempEntry {
			edit = true
			if e.Tags == "" {
				dropDownFields = append(dropDownFields, "tags")
			}
		}
		tempField = encrypt.Field{}
		tempStr := ""
		fieldType := "" // To track what field is changing
		newFieldForm.Clear(true)

		fieldDropDown := tview.NewDropDown().
			SetLabel("new field: ").
			SetCurrentOption(-1).
			SetListStyles(tcell.Style{}.Background(blue).Foreground(white), tcell.Style{}.Background(white).Foreground(blue)) // changes the colors of the drop down options -- selected & unselected styles
		fieldDropDown.SetOptions(dropDownFields, func(chosenDrop string, index int) {
			for newFieldForm.GetFormItemCount() > 1 { // needed for when you change your mind
				newFieldForm.RemoveFormItem(1)
			}
			fieldType = chosenDrop
			if index > -1 { // If something is chosen
				switch fieldType {
				case "tags":
					newFieldForm.AddInputField("tags:", tempEntry.Tags, 50+width, nil, func(tags string) {
						tempStr = tags
					})
				case "url":
					newFieldForm.AddInputField("url:", "", 50+width, nil, func(url string) {
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

					tempField.DisplayName = initialValue

					newFieldForm.AddInputField(inputLabel, initialValue, 50+width, nil, func(display string) {
						tempField.DisplayName = display
					})

					newFieldForm.AddInputField("value:", "", 50+width, nil, func(value string) {
						tempField.Value = value
					})
				}
			}
		})
		newFieldForm.AddFormItem(fieldDropDown).AddButton("save field", func() {
			if (tempField.DisplayName != "") || (tempStr != "") {
				switch fieldType {
				case "username":
					e.Usernames = append(e.Usernames, tempField)
				case "password":
					e.Passwords = append(e.Passwords, tempField)
				case "security question":
					e.SecurityQ = append(e.SecurityQ, tempField)
				case "tags":
					e.Tags = tempStr
				case "url":
					e.Urls = append(e.Urls, tempStr)
				}
				if !edit { // If in /new
					blankFieldsAdded()
					infoText.SetText(newInfo)
					pages.SwitchToPage("newEntry")
					app.SetFocus(newEntry.form)
				} else { // If in /edit
					switchToEditList(true)
				}
			}
		}).
			AddButton("quit", func() {
				if !edit {
					infoText.SetText(newInfo)
					pages.SwitchToPage("newEntry")
					app.SetFocus(newEntry.form)
				} else {
					switchToEditList(false)
				}
			})
	}

	// This is the form for adding or editing notes and its function.
	newNoteForm := tview.NewForm().
		SetButtonBackgroundColor(blue).
		SetFieldBackgroundColor(blue).
		SetLabelColor(lavender)

	// Takes in a pointer to an entry if used in /edit. Takes in a pointer to
	// tempEntry if in /new.
	blankNewNote := func(e *encrypt.Entry) {
		newNoteForm.Clear(true)
		toAdd := e.Notes

		newNoteForm.
			AddInputField("notes:", toAdd[0], 0, nil, func(inputed string) {
				toAdd[0] = inputed
			})

		for i := 1; i < 6; i++ {
			newNoteForm.AddInputField("", toAdd[i], 0, nil, func(inputed string) {
				toAdd[i] = inputed
			})
		}

		newNoteForm.
			AddButton("save", func() {
				e.Notes = toAdd
				if e == &tempEntry { // if this is being done in /new
					pages.SwitchToPage("newEntry")
					app.SetFocus(newEntry.form)
				} else { // if this is being done in /edit
					switchToEditList(true)
				}
			}).
			AddButton("quit", func() {
				if e == &tempEntry { // if being done in /new
					pages.SwitchToPage("newEntry")
					app.SetFocus(newEntry.form)
				} else { // if being done in /edit
					switchToEditList(false)
				}
			}).
			AddButton("delete", func() {
				e.Notes = [6]string{} // assigns the whole array at once
				if e == &tempEntry {
					pages.SwitchToPage("newEntry")
					app.SetFocus(newEntry.form)
				} else {
					switchToEditList(true)
				}
			})
	}

	// For editing the name, tags, or url.
	blankEditStringForm = func(display, value string, e *encrypt.Entry, edit bool) {
		if (display != "name") && (display != "tags") && (display != "url") {
			switchToError(" Unexpected input!\n blankEditStringForm can only change name, tags, or url")
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

		editFieldForm.Clear(true)
		tempDisplay := display
		tempValue := value
		editFieldForm.
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
						switchToError("Tried to edit url in an entry, but could not find the url in the entry's list of urls")
						return
					}
					e.Urls[index] = tempValue
				}
				if (display == "tags") || (edit) {
					switchToEditList(true)
				} else {
					switchToNewFieldsList(true)
				}
			}).
			AddButton("quit", func() {
				if (display == "tags") || (edit) {
					switchToEditList(true)
				} else {
					switchToNewFieldsList(true)
				}
			})
		// Can only delete tags or url, not the name
		if display == "tags" || display == "urls" {
			editFieldForm.AddButton("delete", func() {

				if display == "tags" {
					e.Tags = ""
				} else { // is url

					// where index is the index of the inputted value in the
					// slice of urls of the entry
					// should not happen because value should be in the entry
					// list because it was given from the entry!
					if index < -1 {
						switchToError("Tried to delete url from an entry, but could not find the url in the entry's list of urls")
						return
					}

					// code copied from editFieldForm
					// Currently it changes the order when the element
					// is deleted from the slice. If this is wanted to
					// stay in order, then it should be rewritten.
					(e.Urls)[index] = (e.Urls)[len(e.Urls)-1]
					e.Urls = (e.Urls)[:len(e.Urls)-1]
				}

				if (display == "tags") || (edit) {
					switchToEditList(true)
				} else {
					switchToNewFieldsList(true)
				}

			})
		}
	}

	// This is the little pop up to ask if you're sure when you want to delete
	// an entry. The flex of it is to combine the text and the form.
	editDelete := textFormFlex{
		text: tview.NewTextView().
			SetTextAlign(tview.AlignCenter).
			SetText("delete entry?\nCANNOT BE UNDONE"),
		form: tview.NewForm().SetButtonBackgroundColor(blue).SetLabelColor(lavender),
		flex: tview.NewFlex().SetDirection(tview.FlexRow),
	}

	blankEditDeleteEntry := func() {
		editDelete.form.Clear(true)
		editDelete.form.SetButtonsAlign(tview.AlignCenter)
		editDelete.form.
			AddButton("cancel", func() {
				switchToEditList(false)
			}).
			AddButton("delete", func() { // deletes element from slice, slower version, keeps everything else in order, copied the code from a website lol
				copy(entries[indexSelected:], entries[indexSelected+1:])
				entries = entries[:len(entries)-1]
				if writeFileErr() {
					switchToHome()
				}
			})
	}

	editFieldInfo := " /edit \n ----- \n to move: \n -tab \n -back tab\n\n to select: \n -return \n\n must name \n field to \n save it \n\n press quit \n to leave"

	blankEditList = func(i int) {
		editList.Clear()
		e := &entries[i]
		letter := newCharIterator()

		editList.AddItem("leave /edit "+strconv.Itoa(i), "(takes you back to /home)", rune(letter), func() {
			switchToHome()
		})
		letter = increment(letter)
		editList.AddItem("name:", e.Name, rune(letter), func() {
			infoText.SetText(editFieldInfo)
			blankEditStringForm("name", e.Name, e, true)
			pages.ShowPage("editFieldStr")
			app.SetFocus(editFieldForm)
		})
		if e.Tags != "" {
			letter = increment(letter)
			editList.AddItem("tags:", e.Tags, rune(letter), func() {
				infoText.SetText(editFieldInfo)
				blankEditStringForm("tags", e.Tags, e, true)
				pages.ShowPage("editFieldStr")
				app.SetFocus(editFieldForm)
			})
		}
		for _, u := range e.Urls {
			letter = increment(letter)
			editList.AddItem("url:", u, rune(letter), func() {
				infoText.SetText(editFieldInfo)
				blankEditStringForm("url", u, e, true)
				pages.ShowPage("editFieldStr")
				app.SetFocus(editFieldForm)
			})
		}
		for i := range e.Usernames {
			u := &e.Usernames[i]
			letter = increment(letter)

			editList.AddItem(u.DisplayName+":", u.Value, rune(letter), func() {
				infoText.SetText(editFieldInfo)
				blankEditFieldForm(u, &e.Usernames, i, true)
				pages.ShowPage("edit-editField")
				app.SetFocus(editFieldForm)
			})
		}
		for i := range e.Passwords {
			p := &e.Passwords[i]
			letter = increment(letter)

			editList.AddItem(p.DisplayName+":", "[black:black]"+p.Value, rune(letter), func() {
				infoText.SetText(editFieldInfo)
				blankEditFieldForm(p, &e.Passwords, i, true)
				pages.ShowPage("edit-editField")
				app.SetFocus(editFieldForm)
			})
		}
		for i := range e.SecurityQ {
			sq := &e.SecurityQ[i]
			letter = increment(letter)

			editList.AddItem(sq.DisplayName+":", "[black:black]"+sq.Value, rune(letter), func() {
				infoText.SetText(editFieldInfo)
				blankEditFieldForm(sq, &e.SecurityQ, i, true)
				pages.ShowPage("edit-editField")
				app.SetFocus(editFieldForm)
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
			editList.AddItem("notes:", condensedNotes, rune(letter), func() {
				infoText.SetText(editFieldInfo)
				blankNewNote(e)
				pages.ShowPage("newNote")
				app.SetFocus(newNoteForm)
			})
		} else {
			letter = increment(letter)
			editList.AddItem("add notes:", "(none written so far)", rune(letter), func() {
				infoText.SetText(editFieldInfo)
				blankNewNote(e)
				pages.ShowPage("newNote")
				app.SetFocus(newNoteForm)
			})
		}
		newFieldStr := ""
		if e.Tags == "" {
			newFieldStr += "tags, "
		}
		letter = increment(letter)
		editList.AddItem("add new field", newFieldStr+"urls, usernames, passwords, security questions", rune(letter), func() {
			infoText.SetText(editFieldInfo)
			// code copied from blankNewEntry
			blankNewField(e)
			pages.ShowPage("newField")
			app.SetFocus(newFieldForm)
		})
		letter = increment(letter)
		if e.Circulate { // If it is in circulation, option to opt out
			editList.AddItem("remove from circulation", "(not permanent), check /help for info", rune(letter), func() {
				e.Circulate = false
				switchToEditList(true)
			})

		} else { // If it's not in circulation, option to opt back in
			editList.AddItem("add back to circulation", "(not permanent), check /help for info", rune(letter), func() {
				e.Circulate = true
				switchToEditList(true)
			})
		}
		letter = increment(letter)
		editList.AddItem("delete entry", "(permanent!)", rune(letter), func() {
			infoText.SetText(editFieldInfo)
			blankEditDeleteEntry()
			pages.ShowPage("editDelete")
			app.SetFocus(editDelete.form)
		})
	}

	// An entry is passed in for /copy. If making a brand new entry, then a
	// blank tempEntry is passed in.
	blankNewEntry := func(e encrypt.Entry) {
		newEntry.form.Clear(true)
		newFieldsAddedList.Clear()

		// This must be done one by one because of pointer shenanigans
		// Usernames, Passwords, SecurityQ, Urls are slices so must
		// be copied manually. Right now, notes is limited to six strings
		// [6]string so is an array, not a pointer.
		tempEntry.Name = e.Name
		tempEntry.Tags = e.Tags

		tempEntry.Urls = make([]string, len(e.Urls))
		copy(tempEntry.Urls, e.Urls)

		tempEntry.Usernames = make([]encrypt.Field, len(e.Usernames))
		copy(tempEntry.Usernames, e.Usernames)

		tempEntry.Passwords = make([]encrypt.Field, len(e.Passwords))
		copy(tempEntry.Passwords, e.Passwords)

		tempEntry.SecurityQ = make([]encrypt.Field, len(e.SecurityQ))
		copy(tempEntry.SecurityQ, e.SecurityQ)

		tempEntry.Notes = e.Notes
		tempEntry.Circulate = true

		newEntry.form.
			AddInputField("name:", tempEntry.Name, 58+width, nil, func(itemName string) {
				tempEntry.Name = itemName
			}).
			AddInputField("tags:", tempEntry.Tags, 58+width, nil, func(tagsInput string) {
				tempEntry.Tags = tagsInput
			}).
			AddCheckbox("circulate:", true, func(checked bool) {
				tempEntry.Circulate = checked
			}).
			// this order of the buttons is on purpose and makes sense
			AddButton("new field", func() {
				infoText.SetText(newFieldInfo)
				blankNewField(&tempEntry)
				pages.ShowPage("newField")
				app.SetFocus(newFieldForm)
			}).
			// You can't hit save if there's no name
			AddButton("save entry", func() {
				if tempEntry.Name != "" {
					tempEntry.Created = time.Now()
					entries = append(entries, tempEntry)
					if writeFileErr() { // if successfully wrote to file, then it switches to home, if not then it switches to error page
						switchToHome()
					}
				}
			}).
			AddButton("quit", func() {
				switchToHome()
			}).
			AddButton("notes", func() {
				blankNewNote(&tempEntry)
				pages.ShowPage("newNote")
				app.SetFocus(newNoteForm)
			})
		// Put at the end so in case there is already fields it puts the button at the end
		switchToNewFieldsList(false)
	}

	// This is the text box used for /copen.
	copenList := tview.NewList().
		SetSecondaryTextColor(blue).
		SetShortcutColor(lavender).
		SetDoneFunc(switchToHome) // Needs to happen after switchToHome is filled

	blankCopen := func(i int) {
		letter := newCharIterator()
		copenList.Clear()
		e := entries[i]

		copenList.AddItem("leave /copen "+strconv.Itoa(i), "(takes you back to /home)", rune(letter), func() {
			clipboard.WriteAll("banana")
			switchToHome()
		})
		letter = increment(letter)
		copenList.AddItem("name:", e.Name, rune(letter), func() {
			clipboard.WriteAll(e.Name)
		})
		if e.Tags != "" {
			letter = increment(letter)
			copenList.AddItem("tags:", e.Tags, rune(letter), func() {
				clipboard.WriteAll(e.Tags)
			})
		}
		for _, u := range e.Urls {
			letter = increment(letter)
			copenList.AddItem("url:", u, rune(letter), func() {
				clipboard.WriteAll(u)
			})
		}
		for _, u := range e.Usernames {
			letter = increment(letter)
			copenList.AddItem(u.DisplayName+":", u.Value, rune(letter), func() {
				clipboard.WriteAll(u.Value)
			})
		}
		for _, p := range e.Passwords {
			letter = increment(letter)
			copenList.AddItem(p.DisplayName+":", "[black:black]"+p.Value, rune(letter), func() {
				clipboard.WriteAll(p.Value)
			})
		}
		for _, sq := range e.SecurityQ {
			letter = increment(letter)
			copenList.AddItem(sq.DisplayName+":", "[black:black]"+sq.Value, rune(letter), func() {
				clipboard.WriteAll(sq.Value)
			})
		}
		for _, n := range e.Notes {
			if n != "" {
				letter = increment(letter)
				copenList.AddItem("note:", n, rune(letter), func() {
					clipboard.WriteAll(n)
				})
			}
		}
		letter = increment(letter)
		copenList.AddItem("in circulation:", strconv.FormatBool(e.Circulate), rune(letter), func() {
			clipboard.WriteAll(strconv.FormatBool(e.Circulate))
		})
		if !e.Modified.IsZero() {
			letter = increment(letter)
			copenList.AddItem("date last modified:", fmt.Sprint(e.Modified.Date()), rune(letter), func() {
				clipboard.WriteAll(fmt.Sprint(e.Modified.Date()))
			})
		}
		if !e.Opened.IsZero() {
			letter = increment(letter)
			copenList.AddItem("date last opened:", fmt.Sprint(e.Opened.Date()), rune(letter), func() {
				clipboard.WriteAll(fmt.Sprint(e.Opened.Date()))
			})
		}
		if !e.Created.IsZero() {
			letter = increment(letter)
			copenList.AddItem("date created:", fmt.Sprint(e.Created.Date()), rune(letter), func() {
				clipboard.WriteAll(fmt.Sprint(e.Created.Date()))
			})
		}
		entries[i].Opened = time.Now()
		writeFileErr()
	}

	// openText is the text box used for /open.
	openText := tview.NewTextView().SetScrollable(true).SetDynamicColors(true)
	openInfo := " /open\n -----\n to edit:\n  /edit #\n to copy:\n  /copen # \n\n /home\n /help\n /quit\n\n /new\n /copy #\n\n /find str\n /flist str\n\n /list\n /pick\n /picc\n\n /comp # #\n /reused"
	copenInfo := " /copen \n ------\n to edit: \n /edit # \n\n to move: \n -tab \n -back tab \n -arrows keys\n -scroll\n\n to select:\n -return\n\n to leave:\n -esc key\n -a"

	// This is for /list as well as /find. It has the title textBox (/find str
	// or /list) as well as the text textBox where it will list the entries.
	list := twoTextFlex{
		title: tview.NewTextView().SetWrap(false),
		text:  tview.NewTextView().SetScrollable(true).SetWrap(false),
		flex:  tview.NewFlex(),
	}

	// This list is for /pick, /picc, and /flist.
	pickList := tview.NewList().
		SetSelectedFocusOnly(true).
		SetSecondaryTextColor(blue).
		SetShortcutColor(lavender).
		SetDoneFunc(switchToHome) // Needs to happen after switchToHome is filled
	// The following will add /pick or /picc in the function itself
	pickInfo := " to move: \n -tab \n -back tab \n -arrows keys\n\n to select: \n -return\n -click\n\n to leave: \n -esc key\n -a"

	// Action is either going to be "pick", "picc", or "flist str". This is
	// done to print out the action and send the function to the correct place.
	blankPickList := func(action string, indexes []int) {
		printCommand := action

		if len([]rune(action)) > 5 { // if is /flist str or at all /flist
			printCommand = "flist\n ------ \n"
			if len([]rune(action)) > (56 + width) {
				action = action[:(53+width)] + "..."
			}
		} else {
			printCommand += "\n ----- \n"
		}

		infoText.SetText(" " + printCommand + pickInfo)
		letter := newCharIterator()
		pickList.Clear()
		pickList.AddItem("leave "+action, "(takes you back to /home)", rune(letter), func() {
			switchToHome()
		})
		for _, i := range indexes {
			// in circulation or in /flist str
			if (entries[i].Circulate) || (len([]rune(action)) > 5) {
				letter = increment(letter)

				var title string

				if !entries[i].Circulate {
					title = "(rem) "
				}
				title += "[" + strconv.Itoa(i) + "] " + entries[i].Name

				pickList.AddItem(title, "tags: "+entries[i].Tags, rune(letter), func() {
					if action == "pick" { // to transfer to /open #
						app.EnableMouse(false)
						pages.SwitchToPage("open")
						app.SetFocus(commandLineInput)
						canTypeCommandLinePlaceholder()
						openText.SetText(blankOpen(i, entries))
						infoText.SetText(openInfo)
						writeFileErr()
					} else { // to transfer to /copen # (for both /picc and /flist)
						app.SetFocus(copenList)
						app.EnableMouse(false)
						pages.SwitchToPage("copen")
						infoText.SetText(copenInfo)
						blankCopen(i)
					}
				})
			}
		}
	}

	// This is the input command line for putting in your password to the
	// password manager and also its function for what to do with the input.
	passwordInput := tview.NewInputField().
		SetLabel("password: ").
		SetFieldWidth(71 + width).
		SetMaskCharacter('*').
		SetFieldBackgroundColor(blue).
		SetLabelColor(lavender).
		SetPlaceholderStyle(placeholdStyle)

	// passBoxPages switches between passBox and passErr
	passBoxPages := tview.NewPages()
	// passPages switches between the locked screen and the unlocked normal
	// password manager.
	passPages := tview.NewPages()

	// This is the error text when logging in
	passErr := twoTextFlex{
		title: tview.NewTextView().
			SetWrap(false).
			SetText(" Uh oh! There was an error in signing in:"),
		text: tview.NewTextView().SetScrollable(true).SetWrap(false),
		flex: tview.NewFlex(),
	}

	passActions := func(key tcell.Key) {
		// guarantee only enter (13) or tab (9) can be counted
		if (key != 13) && (key != 9) {
			return
		}

		passInputed := passwordInput.GetText()
		passwordInput.SetText("")

		if (passInputed == "quit") || (passInputed == "q" ||
			(passInputed == "\\quit") || (passInputed == "\\q")) {
			app.Stop()
		}
		passBoxPages.SwitchToPage("passBox")
		var keyErr string

		ciphBlock, keyErr = encrypt.KeyGeneration(passInputed)

		if keyErr != "" {
			passBoxPages.SwitchToPage("passErr")
			passErr.text.SetText(keyErr)
			passwordInput.SetText("")
			return
		}
		readErr := encrypt.ReadFromFile(&entries, ciphBlock)

		if readErr != "" {
			passBoxPages.SwitchToPage("passErr")
			passErr.text.SetText(readErr)
			passwordInput.SetText("")
			return
		}
		passPages.SwitchToPage("passManager")
		switchToHome()
	}
	passwordInput.SetDoneFunc(passActions)

	listInfo := " /list\n -----\n to open:\n  /open #\n to copy:\n  /copen #\n to edit:\n  /edit #\n\n /home\n /help\n /quit\n\n /new\n /copy #\n\n /find str\n /flist str\n\n /pick\n /picc\n\n /comp # #\n /reused"
	findInfo := " /find\n -----\n to open:\n  /open #\n to copy:\n  /copen #\n to edit:\n  /edit #\n\n /home\n /help\n /quit\n\n /new\n /copy #\n\n /flist str\n\n /list\n /pick\n /picc\n\n /comp # #\n /reused"

	// /test or /t is a secret command. It does fmt.Sprint(entries) and prints
	// it to testText. It doesn't blot out any of the passwords.
	// /test # adds # many sample entries to the entry list
	testText := tview.NewTextView().SetScrollable(true)

	// This is the text box for /comp
	compText := tview.NewTextView().SetScrollable(true)

	// The text box for /reused
	reusedText := tview.NewTextView().SetScrollable(true).SetDynamicColors(true)

	// The text box for /help
	helpText := tview.NewTextView().SetScrollable(true).SetText(` /help
 -----
 In order to quit, press control+c or type /quit. 

 # means entry number and str means some text. 
 	example of /open # is: /open 3 
 	example of /find str is: /find library

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
 /open #, so if you have too many entries use /copen #.

 Use /new to make a new entry. You must give your entry a name to 
 save it. You must also give each field a display name in order to 
 save them. You can also write in notes and you can edit the 
 fields you've already added. You don't need to write tags, but 
 they can be helpful in searching for entries using /find str. 
 You can also do /copy # which is the same as doing /new but info
 is already filled out from entry #.
 Creating a new entry is not saved until you click the save button 
 and you are moved away from /new. 

 Use /edit # to edit an existing entry. You can edit the fields 
 already there, add new ones, remove it from circulation, or 
 delete it. While there is a delete button, it is recommended that 
 you remove it from circulation instead. When that is done, it 
 won't show up in /list or /pick. All of the other commands (such 
 as /open, /edit, etc.) will still work on it. 
 Edits are saved as soon as you click save on each specific field.

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
 variables are at the very beginning of func main() in pass.go.

 Here is a list of shortcuts for the commands which will do the 
 same thing as the normal commands:
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
  /reused → /r
The slash / is option, so "open 0" or "o 0" works like "/open 0".

 If you want to change your password or the password parameters,
 run changeKey.go. 

 More info about the project is on the README at 
 https://github.com/ksharnoff/pass.`)

	// First, 'inputted' is sanitized and checked to make sure it follows
	// conventions. Then, a page and focus is swapped and an action is called.
	commandLineActions := func(key tcell.Key) {
		// guarantee only enter (13) or tab (9) can be used
		if (key != 13) && (key != 9) {
			return
		}

		updateTerminalSize()
		app.EnableMouse(true)
		switchToHome()
		infoText.ScrollToBeginning()

		inputed := commandLineInput.GetText()
		commandLineInput.SetText("")
		inputedArr := strings.Split(inputed, " ")
		action := inputedArr[0]

		compIndSelectOne := -1
		compIndSelectTwo := -1

		// Three+ of the commands you need this, have it to be updated if you
		// add new entries
		listAllIndexes := make([]int, len(entries))
		for i := 0; i < len(entries); i++ {
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
			indexSelected = -1 // Sets it here to remove any previous doings

			if len(inputedArr) < 2 { // if there is no number written
				switchToError(" To " + action[1:] + " an entry you must write " + action + " and then a number.\n Ex: \n\t" + action + " 3")
				return
			}
			intTranslated, intErr := strconv.Atoi(inputedArr[1])

			if intErr != nil { // if what passed in is not a number
				switchToError(" Make sure to use " + action + " by writing a number!\n Ex: \n\t " + action + " 3")
				return
			}
			if (intTranslated >= len(entries)) || (intTranslated < 0) { // if the number passed in isn't an index
				switchToError(" The number you entered does not correspond to an entry.\n Do /list to see the entries (and their numbers) that exist.")
				return
			}
			indexSelected = intTranslated

		} else if action == "comp" {
			if len(inputedArr) < 3 {
				switchToError(" You must specify which two entries you would like to /comp.\n Ex: \n\t /comp 3 4")
				return
			}
			compOneInt, compOneErr := strconv.Atoi(inputedArr[1])
			compTwoInt, compTwoErr := strconv.Atoi(inputedArr[2])

			if (compOneErr != nil) || (compTwoErr != nil) {
				switchToError(" Make sure to only use /comp by writing a number! \n Ex: \n\t /comp 3 4")
				return
			}
			if compOneInt == compTwoInt {
				switchToError(" The entries you tried to /comp are the same.\n Therefore, all the passwords would be the same! \n Do /list to see the entries (and their numbers that exist)")
				return
			}
			if ((compOneInt >= len(entries)) || (compTwoInt >= len(entries))) || ((compOneInt < 0) || (compTwoInt < 0)) {
				switchToError(" The number you entered does not correspond to an entry.\n Do /list to see the entries (and their numbers) that exist.")
				return
			}
			compIndSelectOne = compOneInt
			compIndSelectTwo = compTwoInt

		} else if (action == "find") || (action == "flist") {
			// old error message: "To find entries you must write /find and then characters. \n With a space after /find. \n Ex: \n\t /find bank" <-- is specifying the space better?
			if (len(inputedArr) < 2) || (inputedArr[1] == " ") || (inputedArr[1] == "") {
				switchToError(" To find entries you must write " + action + " and then characters. \n Ex: \n\t " + action + " bank")
				return
			}
		}
		switch action {
		case "home", "h":
			pages.SwitchToPage("home")
		case "quit", "q":
			app.Stop()
		case "list", "l":
			text := listEntries(entries, listAllIndexes, false)
			list.title.SetText(" /list \n -----")
			list.text.SetText(text).ScrollToBeginning()
			infoText.SetText(listInfo)
			pages.SwitchToPage("list")
		case "find":
			title, text := blankFind(entries, inputedArr[1])
			list.title.SetText(title)
			list.text.SetText(text).ScrollToBeginning()
			pages.SwitchToPage("list")
			infoText.SetText(findInfo)
		case "test", "t":
			// if /test # then add # entries to the entries list
			if len(inputedArr) > 1 {
				intTranslated, intErr := strconv.Atoi(inputedArr[1])
				if intErr == nil {
					for i := 0; i < intTranslated; i++ {
						entries = append(entries, encrypt.Entry{
							Name:      "testing testing-123456789",
							Tags:      "This was automatically added!",
							Circulate: true})
					}
					// write to file, swap to error page if failed:
					writeFileErr()
					return
				} else {
					switchToError(" To add entries using /test you must write a number!\n Ex: \n\t \\test 3")
					return
				}
			}
			testText.SetText(testAllFields(entries))
			pages.SwitchToPage("test")
		case "new", "n":
			app.EnableMouse(false)
			infoText.SetText(newInfo)
			tempEntry = encrypt.Entry{}
			blankNewEntry(tempEntry)
			app.SetFocus(newEntry.form)
			cantTypeCommandLinePlaceholder()
			pages.SwitchToPage("newEntry")
		case "help", "he":
			helpText.ScrollToBeginning()
			pages.SwitchToPage("help")
		case "open":
			infoText.SetText(openInfo)
			app.EnableMouse(false)
			pages.SwitchToPage("open")
			openText.SetText(blankOpen(indexSelected, entries))
			openText.ScrollToBeginning()
			writeFileErr()
		case "copen":
			infoText.SetText(copenInfo)
			app.SetFocus(copenList)
			app.EnableMouse(false)
			pages.SwitchToPage("copen")
			// needs to be at the end, because writeErr is called from it
			blankCopen(indexSelected)
		case "edit":
			tempEntry = encrypt.Entry{}
			app.EnableMouse(false)
			infoText.SetText(editInfo)
			cantTypeCommandLinePlaceholder()
			switchToEditList(false)
		case "pick", "pk":
			blankPickList("pick", listAllIndexes)
			app.SetFocus(pickList)
			pages.SwitchToPage("pick")
			cantTypeCommandLinePlaceholder()
		case "copy":
			tempEntry = encrypt.Entry{}
			infoText.SetText(newInfo)
			app.EnableMouse(false)
			blankNewEntry(entries[indexSelected])
			app.SetFocus(newEntry.form)
			cantTypeCommandLinePlaceholder()
			pages.SwitchToPage("newEntry")
		case "picc", "p":
			blankPickList("picc", listAllIndexes)
			app.SetFocus(pickList)
			pages.SwitchToPage("pick")
			cantTypeCommandLinePlaceholder()
		case "flist":
			indexesFound := findIndexes(entries, inputedArr[1])
			blankPickList("flist "+inputedArr[1], indexesFound)
			app.SetFocus(pickList)
			pages.SwitchToPage("pick")
			cantTypeCommandLinePlaceholder()
		case "comp":
			app.EnableMouse(true)
			compText.SetText(blankComp(compIndSelectOne, compIndSelectTwo, entries))
			pages.SwitchToPage("comp")
		case "reused", "r":
			reusedText.SetText(" /reused\n -------\n The following are the passwords and answers reused:\n\n" + reusedAll(entries))
			pages.SwitchToPage("reused")
		default:
			switchToError(" That input doesn't match a command! \n Look to the right right to see the possible commands. \n Make sure to spell it correctly!")
		}
	}
	commandLineInput.SetDoneFunc(commandLineActions)

	passErr.flex.SetDirection(tview.FlexRow).
		AddItem(passErr.title, 2, 0, false).
		AddItem(passErr.text, 0, 1, false)

	// This is passBoxPages and passwordInput
	passFlex := tview.NewFlex().
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(passBoxPages, 0, 1, false).
			AddItem(grider(passwordInput), 3, 0, false), 0, 1, false)

	// Flex to situate newFieldForm in the middle of the page.
	newFieldFlex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 2, false).
			AddItem(grider(newFieldForm), 11, 0, false).
			AddItem(nil, 0, 1, false), 0, 4, false)

	error.flex.SetDirection(tview.FlexRow).
		AddItem(error.title, 2, 0, false).
		AddItem(error.text, 0, 1, false)

	// Flex to situate editing or making notes in the middle.
	newNoteFlex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 2, false).
			// following two, 5 is the max for changing
			AddItem(grider(newNoteForm), 17, 0, false). // 0 6
			AddItem(nil, 0, 1, false), 0, 5, false)

	newEntry.flex.SetDirection(tview.FlexRow).
		AddItem(newEntry.form, 9, 0, false).
		AddItem(newFieldsAddedList, 0, 1, false)

	// Created the grid here which is added to the following
	// three flexes.
	editFieldGrid := grider(editFieldForm)

	// To situate edit field differently if you're in /new.
	newEditFieldFlex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 3, false).
			AddItem(editFieldGrid, 9, 0, false).
			AddItem(nil, 0, 2, false), 0, 4, false)

	// To situate edit field differently if you're in /edit
	editEditFieldFlex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 2, false).
			AddItem(editFieldGrid, 9, 0, false). // 0, 3
			AddItem(nil, 0, 3, false), 0, 4, false)
	// To situate edit field if it is editing name, tags, or url in /edit

	editFieldStrFlex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 3, false).
			AddItem(editFieldGrid, 7, 0, false). // 0, 2
			AddItem(nil, 0, 2, false), 0, 4, false)

	// This contains the text and form for the pop up asking about deleting an
	// entry
	editDelete.flex.
		AddItem(editDelete.text, 0, 1, false).
		AddItem(editDelete.form, 0, 1, false)

	// To situate the editDelete.flex in the middle
	editDeleteFlex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 2, false). // 2
			AddItem(grider(editDelete.flex), 0, 2, false).
			AddItem(nil, 0, 2, false), 0, 1, false).
		AddItem(nil, 0, 1, false)

	list.flex.SetDirection(tview.FlexRow).
		AddItem(list.title, 2, 0, false).
		AddItem(list.text, 0, 1, false)

	// Added to pages for /home
	sadEmptyBox := tview.NewBox().SetBorder(true).SetTitle("sad, empty box")

	// All the different pages are added here. The order in which the pages are
	// added matters.
	pages.
		AddPage("home", sadEmptyBox, true, false).
		AddPage("list", grider(list.flex), true, false).
		AddPage("test", grider(testText), true, false).
		AddPage("edit", grider(editList), true, false).
		AddPage("help", grider(helpText), true, false).
		AddPage("err", grider(error.flex), true, false).
		AddPage("open", grider(openText), true, false).
		AddPage("pick", grider(pickList), true, false).
		AddPage("copen", grider(copenList), true, false).
		AddPage("newEntry", grider(newEntry.flex), true, false).
		AddPage("newField", newFieldFlex, true, false).
		AddPage("newNote", newNoteFlex, true, false).
		AddPage("new-editField", newEditFieldFlex, true, false).
		AddPage("editFieldStr", editFieldStrFlex, true, false).
		AddPage("editDelete", editDeleteFlex, true, false).
		AddPage("edit-editField", editEditFieldFlex, true, false).
		AddPage("comp", grider(compText), true, false).
		AddPage("reused", grider(reusedText), true, false)

	// Left side of pass manager, pages and commandLineInput. Ratio of 8:1 is
	// the max on 26x78 (9:1 is the same). Ratio of 9:1 is the max on 28x84
	// grid (10:1 is the same)
	flexRow := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(pages, 0, 1, false).                   // 0, 9
		AddItem(grider(commandLineInput), 3, 0, false) // 0, 1

	// Left and right sides of pass
	flex := tview.NewFlex().
		AddItem(flexRow, 0, 1, false).          // 0, 14
		AddItem(grider(infoText), 15, 0, false) // 0, 3

	// "passBox" just has an empty box for the password screen
	passBoxPages.
		AddPage("passBox", tview.NewBox().SetBorder(true), true, true).
		AddPage("passErr", grider(passErr.flex), true, false)

	// Contains the password screen and the password manager
	passPages.
		AddPage("passInput", passFlex, true, true).
		AddPage("passManager", flex, true, false)

	if err := app.SetRoot(passPages, true).SetFocus(passwordInput).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

// Input: any primitive (widget used from tview)
// Return: grid fitted around the object to add a border
func grider(prim tview.Primitive) *tview.Grid {
	grid := tview.NewGrid().SetBorders(true)
	grid.AddItem(prim, 0, 0, 1, 1, 0, 0, false)
	return grid
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
		b.WriteString("\n notes:")
		blankLines := 0 // Stops printing \n\n\n\n if only text in first line
		for _, n := range e.Notes {
			if n == "" {
				blankLines++
			} else {
				b.WriteString(strings.Repeat("\n", blankLines))

				for len([]rune(n)) > 61+width {
					b.WriteString("\n\t " + n[:61+width])
					n = n[61+width:]
				}

				b.WriteString("\n\t " + n)
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
func updateTerminalSize() {
	termWidth, termHeight, terminalSizeErr := term.GetSize(int(os.Stdin.Fd()))

	if terminalSizeErr != nil {
		width = 0
		height = 0
		return
	}

	width = termWidth - 84
	height = termHeight - 28

	// Through experimenting, if the width is below 75 then you cannot see and
	// access all the buttons in /new. The height should be at least 17 to see
	// the list of fields already added in /new.
	if termWidth < 75 || termHeight < 17 {
		fmt.Println("In order to see all elements of the password manager, please set your terminal to a minimum height of 17 characters and a minimum width of 75 characters.")
		fmt.Println("The ideal dimensions that this was designed for is a height of 28 and a width of 84.")
		os.Exit(1)
	}
}
