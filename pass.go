package main

import (
	"fmt"
	"github.com/atotto/clipboard" // copies the data to clipboard in /copen
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"strconv"
	"strings"
	"time"

	// encryption:
	"crypto/cipher"
	"pass/encrypt"

	// writing to a file:
	"gopkg.in/yaml.v3"
	"os"
)

// if this is changed then you should edit this also in
// in creatEncr.go and changeKey.go
type entry struct {
	Name      string
	Tags      string
	Usernames []Field
	Passwords []Field
	SecurityQ []Field
	Notes     [6]string
	Circulate bool
	Created   time.Time
	Modified  time.Time
	Opened    time.Time
}
type Field struct {
	DisplayName string
	Value       string
}

// The following stucts is to make the naming conventions
// more clear. Instead of "errorText" and "errorTextGrid"
// there is error.text and error.grid.
type textGrid struct {
	text *tview.TextView
	grid *tview.Grid
}
type formGrid struct {
	form *tview.Form
	grid *tview.Grid
}
type listGrid struct {
	list *tview.List
	grid *tview.Grid
}
type textFormFlexGrid struct {
	text *tview.TextView
	form *tview.Form
	flex *tview.Flex
	grid *tview.Grid
}
type inputGrid struct {
	input *tview.InputField
	grid  *tview.Grid
}
type twoTextFlexGrid struct {
	title *tview.TextView
	text  *tview.TextView
	flex  *tview.Flex
	grid  *tview.Grid
}

// The following struct is used in the map in /reused and /comp
type reusedPass struct {
	displayName string
	entryName   string
	entryIndex  int
}

func main() {
	// The colors do not comply with accessibility contrast
	// requirements. You can uncomment out the next two
	// lines and comment out the default colors in order
	// for it to have a higher contrast that complies
	// with WCAG AAA everywhere.

	// lavender := white // this is white
	// blue := tcell.NewRGBColor(0, 0, 255) // this is blue
	lavender := tcell.NewRGBColor(149, 136, 204) // label names and shortcut names
	blue := tcell.NewRGBColor(106, 139, 166) // secondary text in lists, buttons in forms, input field color

	white := tcell.GetColor("white")

	app := tview.NewApplication()

	// The following entry names should only be seen if the manager
	// opens without loading a file.
	entries := []entry{entry{Name: "QUIT NOW, DANGER", Circulate: true}, entry{Name: "SOMETHING'S VERY", Circulate: true}, entry{Name: "BROKEN. QUIT!", Circulate: true}}

	// Pages is the pages set up for the left top box
	pages := tview.NewPages()

	// This is the set up for being logged in, "flex which
	// contains flexRow and commands.text. flexRow contains
	// pages and commandLine.
	flex := tview.NewFlex()
	flexRow := tview.NewFlex().SetDirection(tview.FlexRow)

	// This is to set the background colors of the text in
	// the input lines. There was no function for it, so it
	// had to be done using tcell.Style.
	placeholdStyle := tcell.Style{}.
		Background(blue).
		Foreground(white)

	// This is the text box on the right that contains information
	// that changes depending on what the user is doing.
	info := textGrid{text: tview.NewTextView().SetScrollable(true).SetWrap(false), grid: tview.NewGrid().SetBorders(true)}
	// TO DO - this is only referenced once vvvv
	homeInfo := " commands\n -------- \n /home\n /help\n /quit\n\n /open #\n /copen #\n\n /new\n /copy #\n\n /edit #\n\n /find str\n /flist str\n\n /list\n /pick\n /picc\n\n /comp # #\n /reused"

	// TO DO - this is only referenced once vvvv
	// This is the blank box at /home.
	sadEmptyBox := tview.NewBox().SetBorder(true).SetTitle("sad, empty box")

	// String of what is inputed into the commandLind
	inputed := ""
	// This is the input line used for navigation and input of
	// commands for the entire password manager.
	commandLine := inputGrid{input: tview.NewInputField().
		SetLabel("input: ").SetFieldWidth(60).
		SetFieldBackgroundColor(blue).
		SetLabelColor(lavender).
		SetPlaceholderStyle(placeholdStyle), grid: tview.NewGrid().SetBorders(true)}

	// TO DO - this is only referenced once vvvv 
	// This is the function that handles the input from commandLine
	commandLineActions := func(key tcell.Key) {}

	// This is called when you can type in the commandLine.
	// It changes the placeholder text to say to look at the right,
	// where there is info.text.
	lookRightCommandLinePlaceholder := func() {
		commandLine.input.SetPlaceholder("psst look to the right")
	}
	// This function is called when you can't type in the commandLine
	cantTypeCommandLinePlaceholder := func() {
		commandLine.input.SetPlaceholder("psst you can't type here right now")
	}

	// This is for /list as well as /find. It has the title text
	// box (/find str) as well as the text text box where it will
	// list the entries.
	list := twoTextFlexGrid{title: tview.NewTextView().SetWrap(false), text: tview.NewTextView().SetScrollable(true).SetWrap(false), grid: tview.NewGrid().SetBorders(true), flex: tview.NewFlex()}
	listInfo := " /list\n -----\n to open:\n  /open #\n to copy:\n  /copen #\n to edit:\n  /edit #\n\n /home\n /help\n /quit\n\n /new\n /copy #\n\n /find str\n /flist str\n\n /pick\n /picc\n\n /comp # #\n /reused"
	findInfo := " /find\n -----\n to open:\n  /open #\n to copy:\n  /copen #\n to edit:\n  /edit #\n\n /home\n /help\n /quit\n\n /new\n /copy #\n\n /flist str\n\n /list\n /pick\n /picc\n\n /comp # #\n /reused"

	// This is the text box that /test prints to
	test := textGrid{text: tview.NewTextView().SetScrollable(true), grid: tview.NewGrid().SetBorders(true)}

	// This is the text box for /help and the string for it
	help := textGrid{text: tview.NewTextView().SetScrollable(true).SetText(` /help
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
 delete it. While there is a delete button, it is reccomended that 
 you remove it from circulation instead. When that is done, it 
 won't show up in /list or /pick. All of the other commands (such 
 as /open, /edit, etc.) will still work on it. 
 Edits are saved as soon as you click save on each specific field.

 Use /find str to search for entries. /find str will return all of
 the entries that contain str in the name or the tags. In both
 /find str and /list, the resulting entries may not show their
 full name for space. Use /flist str to see a list of entries with 
 that str, when clicked they are /copen.

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

 Here is a list of shorcuts for the commands which will do the same
 thing as the normal commands:
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

 If you want to change your password or the password paramenters,
 run changeKey.go. 

 More info about the project is on the README at 
 https://github.com/ksharnoff/pass.`), grid: tview.NewGrid().SetBorders(true)}

	// This is the text box used for /open.
	// Also the function for writing to the text box.
	open := textGrid{text: tview.NewTextView().SetScrollable(true).SetDynamicColors(true), grid: tview.NewGrid().SetBorders(true)}
	blankOpen := func(i int) string { return "error, blankOpen(i int) didn't run" }
	openInfo := " /open\n -----\n to edit:\n  /edit #\n to copy:\n  /copen # \n\n /home\n /help\n /quit\n\n /new\n /copy #\n\n /find str\n /flist str\n\n /list\n /pick\n /picc\n\n /comp # #\n /reused"

	// This is the text box used to /copen and its function for
	// making it
	copen := listGrid{list: tview.NewList().SetSecondaryTextColor(blue).
		SetShortcutColor(lavender), grid: tview.NewGrid().SetBorders(true)}
	blankCopen := func(i int) {}
	copenInfo := " /copen \n ------\n to edit: \n /edit # \n\n to move: \n -tab \n -back tab \n -arrows keys\n -scroll\n\n to select:\n -return\n\n to leave:\n -esc key"

	// This is where the errors are written to.
	// error.title stays the same for all errors.
	error := twoTextFlexGrid{title: tview.NewTextView().SetText(" Uh oh! There was an error:"), text: tview.NewTextView().SetScrollable(true), flex: tview.NewFlex(), grid: tview.NewGrid().SetBorders(true)}

	// Switches to the error page, sets error.text to err.
	// Also sets the commandLine to blank again -- useful for 
	// the commandLine func.
	switchToError := func(err string) {
		error.text.SetText(err)
		pages.SwitchToPage("err")
		commandLine.input.SetText("")
	}
	// Switches to home, rights everything again.
	switchToHome := func() {
		pages.SwitchToPage("/home")
		app.SetFocus(commandLine.input)
		info.text.SetText(homeInfo)
		lookRightCommandLinePlaceholder()
		app.EnableMouse(false)
	}

	writeFileErr := func() bool { return false }

	// This is the form and the flex for /new.
	// The flex puts the list of the entries added with the
	// form of /new. The struct being used has text, but this does
	// not use a flex. Also there is the function for it.
	newEntry := textFormFlexGrid{form: tview.NewForm().SetButtonBackgroundColor(blue).SetFieldBackgroundColor(blue).SetLabelColor(lavender), flex: tview.NewFlex(), grid: tview.NewGrid().SetBorders(true)}
	blankNewEntry := func(e entry) {}

	// This is the fields added so far list and its function,
	// used in /new.
	// Also the switchToNewFieldsList function to be used when
	// each field is edited in /new. It creates the button 'edit fields'
	// for after you have created your first fields in /new. It will appear
	// there already if you are in /copy # and # has fields. 
	// It takes in a bool so that it can be
	// called in /copy # but not switch to it.
	newFieldsAddedList := tview.NewList().SetSelectedFocusOnly(true).SetSecondaryTextColor(blue).SetShortcutColor(lavender)
	blankFieldsAdded := func() {}
	switchToNewFieldsList := func(doSwitch bool) {}

	// This is the form of when you're adding a new field.
	newField := formGrid{form: tview.NewForm().SetButtonBackgroundColor(blue).SetFieldBackgroundColor(blue).SetLabelColor(lavender), grid: tview.NewGrid().SetBorders(true)}
	newFieldFlex := tview.NewFlex()     // Flex to situate it in the middle of page.
	blankNewField := func(e *entry) {}  // Function to set up the form.
	newEditFieldFlex := tview.NewFlex() // Flex to situate it slightly differently you're in /edit.

	// fieldType is used to track what type of field is being added.
	fieldType := ""

	// This is the form for adding or editing notes and
	// its function.
	newNote := formGrid{form: tview.NewForm().SetButtonBackgroundColor(blue).SetFieldBackgroundColor(blue).SetLabelColor(lavender), grid: tview.NewGrid().SetBorders(true)}
	newNoteFlex := tview.NewFlex() // Flex to situate it in the middle of the page.
	blankNewNote := func(e *entry) {}

	// These are tempory and used when someone is making a new entry,
	// or new field. Also used when someone is editing an entry.
	tempEntry := entry{}
	tempField := Field{}

	// This is the drop down slice in making a new entry, "tags"
	// can get added in /edit.
	dropDownFields := []string{}

	// This is the text to be shown on the left in info.text
	// when creating a new entry or a new field.
	newInfo := " /new \n ---- \n to move: \n -tab \n -back tab \n\n to select: \n -return \n\n must name \n entry to \n save it \n\n press quit \n to leave"
	newFieldInfo := " /new \n ---- \n to move: \n -tab \n -back tab \n\n to select: \n -return \n\n must name \n field to \n save it \n\n press quit \n to leave" // only change from this one to the newInfo is field vs. entry

	// This is the list for /edit and its function for making it
	edit := listGrid{list: tview.NewList().SetSelectedFocusOnly(true).SetSecondaryTextColor(blue).SetShortcutColor(lavender), grid: tview.NewGrid().SetBorders(true)}
	blankEditList := func(i int) {}

	// runeAlphabet is used for the character shortcuts in lists.
	// The function is used to increase the index. If it at the limit
	// then it goes back to 0.
	runeAlphabet := []rune{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z'}
	runeAlphabetIterate := func(i *int) {
		if *i == len(runeAlphabet) {
			*i = 0
		} else {
			*i++
		}
	}

	// This is the form for editing a specific field and the flexes.
	// There are two functions, one to edit the name or tags.
	// The other is to edit one of the fields (password, username,
	// or security questions)
	editField := formGrid{form: tview.NewForm().SetButtonBackgroundColor(blue).SetFieldBackgroundColor(blue).SetLabelColor(lavender), grid: tview.NewGrid().SetBorders(true)}
	editEditFieldFlex := tview.NewFlex() // Flex to situate it in the middle of the page
	editFieldStrFlex := tview.NewFlex()  // Flex to situate it in the middle of the page, differenct than the other because it is smaller.
	blankEditFieldForm := func(f *Field, fieldArr *[]Field, index int, e *entry, edit bool) {}
	blankEditStringForm := func(display, value string, e *entry) {}

	// This function switches back to the edit list. It remakes
	// the list each time and uses indexSelected It takes in a
	// bool to know whether or not to write to file the changes, as
	// well as whether or not to update the last modified itme.
	switchToEditList := func(modified bool) {}
	// This is the variable for what entry is selected.
	// It is set in commandLineActions and used for the function
	// above.
	indexSelected := -1

	// This is the little pop up to ask if you're sure when you
	// want to delete an entry. The flex of it is to combine
	// the text and the form.
	editDelete := textFormFlexGrid{text: tview.NewTextView().SetTextAlign(tview.AlignCenter).SetText("delete entry?\nCANNOT BE UNDONE"), form: tview.NewForm().SetButtonBackgroundColor(blue).SetLabelColor(lavender), flex: tview.NewFlex().SetDirection(tview.FlexRow), grid: tview.NewGrid().SetBorders(true)}
	editDeleteFlex := tview.NewFlex() // Flex to situate it in the center
	blankEditDeleteEntry := func() {}

	// This is whats written in info.text during /edit
	editInfo := " /edit \n ----- \n to move: \n -tab \n -back tab \n -arrows keys\n -scroll\n\n to select: \n -return \n\n to leave: \n -esc key\n\n "
	editFieldInfo := " /edit \n ----- \n to move: \n -tab \n -back tab\n\n to select: \n -return \n\n must name \n field to \n save it \n\n press quit \n to leave"

	// This is the list and its function used for /pick, /picc, and
	// /flist. blankPickList takes in a string in order to print out
	// what it is, and to send the functions to the right place. It
	// takes in a splice of entries to see what ones get printed.
	pick := listGrid{list: tview.NewList().SetSelectedFocusOnly(true).SetSecondaryTextColor(blue).SetShortcutColor(lavender), grid: tview.NewGrid().SetBorders(true)}
	blankPickList := func(openCopen string, indexes []int) {}
	// The following will add /pick or /picc in the function itself
	pickInfo := " to move: \n -tab \n -back tab \n -arrows keys\n\n to select: \n -return\n -click\n\n to leave: \n -esc key\n\n "

	// This is the text box and the function used for /comp
	// /comp # # takes in two numbers and compares all their
	// passwords to see if there are any in common
	comp := textGrid{text: tview.NewTextView().SetScrollable(true).SetDynamicColors(true), grid: tview.NewGrid().SetBorders(true)}
	blankComp := func(i1 int, i2 int) string {
		return "failed! did not run /comp!"
	}
	compIndSelectOne := -1
	compIndSelectTwo := -1

	reused := textGrid{text: tview.NewTextView().SetScrollable(true).SetDynamicColors(true), grid: tview.NewGrid().SetBorders(true)}

	// This is the cipher block generated with the key to encrypt
	// and decrypt. Normally its the key that gets passed around
	// not the cipher block, but I chose to do it this way because
	// it's less redudant.
	var ciphBlock cipher.Block

	// This is the input command line for putting in your password
	// to the password manager and also its function for what to do
	// with the input.
	password := inputGrid{input: tview.NewInputField().SetLabel("password: ").SetFieldWidth(71).SetMaskCharacter('*').SetFieldBackgroundColor(blue).SetLabelColor(lavender).SetPlaceholderStyle(placeholdStyle), grid: tview.NewGrid().SetBorders(true)}
	passActions := func(key tcell.Key) {}
	// This is the string of what's inputed.
	passInputed := ""

	// This is passBoxPages and password.input
	passFlex := tview.NewFlex()

	// passBoxPages is switching between passBox and passErr
	passBoxPages := tview.NewPages()
	// passPages is switching between the locked screen and
	// unlocked normal passowrd manager.
	passPages := tview.NewPages()

	// This is the blank box for when its locked
	passBox := tview.NewBox().SetBorder(true)

	// This is the error text for when the password is wrong
	// or another error.
	passErr := twoTextFlexGrid{title: tview.NewTextView().SetWrap(false).SetText(" Uh oh! There was an error in signing in:"), text: tview.NewTextView().SetScrollable(true).SetWrap(false), grid: tview.NewGrid().SetBorders(true), flex: tview.NewFlex()}

	// ------------------------------------------------ //
	//    all variables initialized! functions time!    //
	// ------------------------------------------------ //

	passActions = func(key tcell.Key) {
		passInputed = password.input.GetText()

		if passInputed == "/quit" {
			app.Stop()
		}

		passBoxPages.SwitchToPage("passBox")
		var keyErr string

		ciphBlock, keyErr = encrypt.KeyGeneration(passInputed)

		if keyErr != "" {
			passBoxPages.SwitchToPage("passErr")
			passErr.text.SetText(keyErr)
			password.input.SetText("")
			return
		} 

		readErr := readFromFile(&entries, ciphBlock)

		if readErr != "" {
			passBoxPages.SwitchToPage("passErr")
			passErr.text.SetText(readErr)
			password.input.SetText("")
			return
		} 
		passPages.SwitchToPage("passManager")
		switchToHome()

		password.input.SetText("")
	}

	password.input.SetDoneFunc(passActions)

	// extra checks with: open edit copen copy comp find flist
	commandLineActions = func(key tcell.Key) {
		app.EnableMouse(true)
		switchToHome()

		inputed = commandLine.input.GetText()
		inputedArr := strings.Split(inputed, " ")
		action := inputedArr[0]

		// Three+ of the commands you need this, have it to be updated
		// if you add new entries
		listAllIndexes := make([]int, len(entries))
		for i := 0; i < len(entries); i++ {
			listAllIndexes[i] = i
		}

		// if it is one of the actions with extra checks, change it to be
		// its longer name. Less || statements needed therefore.
		if len([]rune(action)) < 5 {
			switch action {
			case "/o":
				action = "/open"
			case "/e" :
				action = "/edit"
			case "/c":
				action = "/copen"
			case "/co":
				action = "/copy"
			case "/com":
				action = "/comp"
			case "/f":
				action = "/find"
			case "/fl":
				action = "/flist"
			}
		}

		// The following is a check for the commands that take in a number.
		// they check: is there a second thing? is it a number? is it a valid number?
		if (action == "/open") || (action == "/edit") || (action == "/copen") || (action == "/copy") {

			indexSelected = -1 //  sets it here to remove any previous doings

			if len(inputedArr) < 2 { // if there is no number written
				switchToError(" To " + action[1:] + " an entry you must write " + action + " and then a number.\n Ex: \n\t" + action + " 3")
				return
			}

			intTranslated, intErr := strconv.Atoi(inputedArr[1])

			if intErr != nil { // if what passed in is not a number
				switchToError(" Make sure to use " + action + " by writing a number!\n Ex: \n\t " + action + " 3")
				return
			}

			if (intTranslated >= len(entries)) || (intTranslated < 0)  { // if the number passed in isn't an index
				switchToError(" The number you entered does not correspond to an entry.\n Do /list to see the entries (and their numbers) that exist.")
				return 
			}

			indexSelected = intTranslated
		} else if action == "/comp" {
			compIndSelectOne = -1
			compIndSelectTwo = -1

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

			if ((compOneInt >= len(entries)) || (compTwoInt >= len(entries))) || ((compOneInt < 0)||(compTwoInt < 0)) {
				switchToError(" The number you entered does not correspond to an entry.\n Do /list to see the entries (and their numbers) that exist.")
				return
			}

			compIndSelectOne = compOneInt
			compIndSelectTwo = compTwoInt
		} else if (action == "/find") || (action == "/flist") {
			// old error message: "To find entries you must write /find and then characters. \n With a space after /find. \n Ex: \n\t /find bank" <-- is specifying the space better?
			if (len(inputedArr) < 2) || (inputedArr[1] == " ") {
				switchToError(" To find entries you must write " + action + " and then characters. \n Ex: \n\t " + action + " bank")
				return
			}
		}

		switch action {
		case "/home", "/h":
			pages.SwitchToPage("/home")
		case "/quit", "/q":
			app.Stop()
		case "/list", "/l":
			title, text := listEntries(entries, listAllIndexes, " /list \n -----", false)
			list.title.SetText(title)
			list.text.SetText(text).ScrollToBeginning()
			info.text.SetText(listInfo)
			pages.SwitchToPage("/list")
		case "/find":
			title, text := findEntries(entries, inputedArr[1])
			list.title.SetText(title)
			list.text.SetText(text).ScrollToBeginning()
			pages.SwitchToPage("/list")
			info.text.SetText(findInfo)

		// /test is not listed on the left set of commands. It
		// does fmt.Sprint(entries) and prints it to /test text
		// box. It doesn't blott out any of the passwords
		case "/test", "/t":
			test.text.SetText(testAllFields(entries))
			pages.SwitchToPage("/test")
		case "/new", "/n":
			app.EnableMouse(false)
			info.text.SetText(newInfo)
			tempEntry = entry{}
			blankNewEntry(tempEntry)
			app.SetFocus(newEntry.form)
			cantTypeCommandLinePlaceholder()
			pages.SwitchToPage("/newEntry")
		case "/help", "/he": // mouse must remain on in order to scroll
			pages.SwitchToPage("/help")
		case "/open":
			info.text.SetText(openInfo)
			app.EnableMouse(false)
			pages.SwitchToPage("/open")
			open.text.SetText(blankOpen(indexSelected)) // needs to be called last because writeErr is called from it, don't want to switch to /open after errors
		case "/copen":
			info.text.SetText(copenInfo)
			app.SetFocus(copen.list)
			app.EnableMouse(false)
			pages.SwitchToPage("/copen")
			blankCopen(indexSelected) // needs to be at the end, same as above
		case "/edit":
			tempEntry = entry{}
			app.EnableMouse(false)
			info.text.SetText(editInfo)
			cantTypeCommandLinePlaceholder()
			switchToEditList(false)
		case "/pick", "/pk":
			blankPickList("/pick", listAllIndexes)
			app.SetFocus(pick.list)
			pages.SwitchToPage("/pick")
			cantTypeCommandLinePlaceholder()
		case "/copy":
			tempEntry = entry{}
			info.text.SetText(newInfo)
			app.EnableMouse(false)
			blankNewEntry(entries[indexSelected])
			app.SetFocus(newEntry.form)
			cantTypeCommandLinePlaceholder()
			pages.SwitchToPage("/newEntry")
		case "/picc", "/p":
			blankPickList("/picc", listAllIndexes)
			app.SetFocus(pick.list)
			pages.SwitchToPage("/pick")
			cantTypeCommandLinePlaceholder()
		case "/flist":
			indexesFound := findIndexes(entries, inputedArr[1])
			blankPickList("/flist " + inputedArr[1], indexesFound)
			app.SetFocus(pick.list)
			pages.SwitchToPage("/pick")
			cantTypeCommandLinePlaceholder()
		case "/comp":
			comp.text.SetText(blankComp(compIndSelectOne, compIndSelectTwo))
			pages.SwitchToPage("/comp")
		case "/reused", "/r":
			app.EnableMouse(false)
			reused.text.SetText(" /reused\n -------\n The following are the passwords and answers reused:\n\n" + reusedAll(entries)) // used to have this be its own blankReused func, not necessary.
			pages.SwitchToPage("/reused")
		default:
			switchToError(" That input doesn't match a command! \n Look to the right right to see the possible commands. \n Make sure to spell it correctly!")
		}
		commandLine.input.SetText("")
	}
	commandLine.input.SetDoneFunc(commandLineActions)

	// Needs this to be done after the function is defined.
	pick.list.SetDoneFunc(switchToHome)
	edit.list.SetDoneFunc(switchToHome)
	copen.list.SetDoneFunc(switchToHome)

	// This tries to write to file, if it fails, it switches to
	// the error page and returns false. 
	// The reason for returning false is so that when used else where 
	// it doesn't switch to error page and then immediatly switch else 
	// where so it can't be seen.
	writeFileErr = func() bool {
		writeErr := writeToFile(entries, ciphBlock)

		if writeErr != "" {
			switchToError(writeErr)
			return false
		}

		return true
	}

	// An entry is passed in for /copy. If making a brand new entry,
	// then a blank tempEntry is passed in.
	blankNewEntry = func(e entry) {
		newEntry.form.Clear(true)
		newFieldsAddedList.Clear()

		// This must be done one by one because of pointer shenanigans
		// Usernames, Passwords, SecurityQ are slices of Fields so must
		// be copied manually
		tempEntry.Name = e.Name
		tempEntry.Tags = e.Tags

		tempEntry.Usernames = make([]Field, len(e.Usernames))
		copy(tempEntry.Usernames, e.Usernames)

		tempEntry.Passwords = make([]Field, len(e.Passwords))
		copy(tempEntry.Passwords, e.Passwords)

		tempEntry.SecurityQ = make([]Field, len(e.SecurityQ))
		copy(tempEntry.SecurityQ, e.SecurityQ)

		tempEntry.Notes = e.Notes
		tempEntry.Circulate = true

		newEntry.form.
			AddInputField("name", tempEntry.Name, 50, nil, func(itemName string) {
				tempEntry.Name = itemName
			}).
			AddInputField("tags", tempEntry.Tags, 50, nil, func(tagsInput string) {
				tempEntry.Tags = tagsInput
			}).
			AddCheckbox("circulate", true, func(checked bool) {
				tempEntry.Circulate = checked
			}).
			// this order of the buttons is on purpose and makes sense
			AddButton("new field", func() {
				info.text.SetText(newFieldInfo)
				blankNewField(&tempEntry)
				pages.ShowPage("/newField")
				app.SetFocus(newField.form)
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
				pages.ShowPage("/newNote")
				app.SetFocus(newNote.form)
			})

		// put at the end so in case there is already fields it puts the button at the end
		switchToNewFieldsList(false)
	}

	// Takes in a pointer to tempEntry if in /new
	// Takes in a pointer to an entry if in /edit
	blankNewField = func(e *entry) {
		edit := false

		dropDownFields = []string{"username", "password", "security question"}

		// Only adds tags as an option to add on if it is in /edit,
		// if there is no tags written already, and if tags isn't
		// already added.
		if e != &tempEntry {
			if (e.Tags == "") && (len(dropDownFields) == 3) {
				dropDownFields = append(dropDownFields, "tags") // don't change the text of "tags", its used elsewhere
			}
			edit = true
		}

		tempField = Field{}
		tempTags := ""
		fieldType = ""
		newField.form.Clear(true)

		fieldDropDown := tview.NewDropDown().SetLabel("new field: ").SetCurrentOption(-1).SetListStyles(tcell.Style{}.Background(blue).Foreground(white), tcell.Style{}.Background(white).Foreground(blue)) // changes the colors of the drop down options -- selected & unselected styles 
		fieldDropDown.SetOptions(dropDownFields, func(chosenDrop string, index int) {
			for newField.form.GetFormItemCount() > 1 { // needed for when you change your mind
				newField.form.RemoveFormItem(1)
			}

			fieldType = chosenDrop
			if index > -1 { // If something is chosen
				if fieldType != "tags" { // If not tags, add displayName and value
					
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

					newField.form.AddInputField(inputLabel, initialValue, 50, nil, func(display string){
						tempField.DisplayName = display
					})

					newField.form.AddInputField("value:", "", 50, nil, func(value string) {
						tempField.Value = value
					})
				} else { // Only has one input line for adding new tags
					newField.form.AddInputField("tags:", tempEntry.Tags, 50, nil, func(tags string) {
						tempTags = tags
					})
				}
			}
		})

		newField.form.AddFormItem(fieldDropDown).AddButton("save field", func() {
			if (tempField.DisplayName != "") || (tempTags != "") {
				switch fieldType {
				case "username":
					e.Usernames = append(e.Usernames, tempField)
				case "password":
					e.Passwords = append(e.Passwords, tempField)
				case "security question":
					e.SecurityQ = append(e.SecurityQ, tempField)
				case "tags":
					e.Tags = tempTags
				}
				if !edit { // If in /new
					blankFieldsAdded()
					info.text.SetText(newInfo)
					pages.SwitchToPage("/newEntry")
					app.SetFocus(newEntry.form)
				} else { // If in /edit
					switchToEditList(true)
				}
			}
		}).
			AddButton("quit", func() {
				if !edit {
					info.text.SetText(newInfo)
					pages.SwitchToPage("/newEntry")
					app.SetFocus(newEntry.form)
				} else {
					switchToEditList(false)
				}
			})
	}

	// Takes in a pointer to an entry if used in /edit.
	// akes in a pointer to tempEntry if in /new.
	blankNewNote = func(e *entry) {
		newNote.form.Clear(true)
		toAdd := e.Notes

		newNote.form.
			AddInputField("notes:", toAdd[0], 0, nil, func(inputed string) {
				toAdd[0] = inputed
			})

		// i := i because making a new function in a closure (for loop) it
		// has i equal to the last iteration of it (would be 6)
		for i := 1; i < 6; i++ {
			i := i
			newNote.form.AddInputField("", toAdd[i], 0, nil, func(inputed string) {
				toAdd[i] = inputed
			})
		}

		// TO DO? idea: write a func to go between edit or home instead of the if/else in each button!
		// ask em what they think :D
		newNote.form.
			AddButton("save", func() {
				e.Notes = toAdd
				if e == &tempEntry { // if this is being done in /new
					pages.SwitchToPage("/newEntry")
					app.SetFocus(newEntry.form)
				} else { // if this is being done in /edit
					switchToEditList(true)
				}
			}).
			AddButton("quit", func() {
				if e == &tempEntry { // if being done in /new
					pages.SwitchToPage("/newEntry")
					app.SetFocus(newEntry.form)
				} else { // if being done in /edit
					switchToEditList(false)
				}
			}).
			AddButton("delete", func() {
				e.Notes = [6]string{} // assigns the whole array at once
				if e == &tempEntry {
					pages.SwitchToPage("/newEntry")
					app.SetFocus(newEntry.form)
				} else {
					switchToEditList(true)
				}
			})
	}

	// This just calls tempEntry to get the fields, this works
	// because tempEntry is defined to be equal to entry e in 
	// blankNewEntry when called after /copy.


	// TO DO -- idea -- stop giving the funcs to do with /new a pointer to tempEntry or to 
	// entry e, instead just defined tempEntry to be e and carry a bool of its in /edit or
	// /new. would be easier? maybe!
	blankFieldsAdded = func() {
		newFieldsAddedList.Clear()
		num := 0 // This iterates through all of runeAlphabet for the shortcuts.

		if newEntry.form.GetButtonIndex("edit field") < 0 { // if there isn't one already
			newEntry.form.
				AddButton("edit field", func() { // Don't change the label name, brakes stuff later.
					app.SetFocus(newFieldsAddedList)
				})
		}
		newFieldsAddedList.
			AddItem("move back to top", "", runeAlphabet[num], func() {
				app.SetFocus(newEntry.form)
			})
		for i := range tempEntry.Usernames {
			i := i
			u := &tempEntry.Usernames[i]
			runeAlphabetIterate(&num)
			newFieldsAddedList.AddItem(u.DisplayName + ":", u.Value, runeAlphabet[num], func() {
				blankEditFieldForm(u, &tempEntry.Usernames, i, &tempEntry, false)
				pages.ShowPage("/new-editField")
				app.SetFocus(editField.form)
			})
		}
		for i := range tempEntry.Passwords {
			i := i
			p := &tempEntry.Passwords[i]
			runeAlphabetIterate(&num)
			newFieldsAddedList.AddItem(p.DisplayName + ":", "[black]" + p.Value, runeAlphabet[num], func() {
				blankEditFieldForm(p, &tempEntry.Passwords, i, &tempEntry, false)
				pages.ShowPage("/new-editField")
				app.SetFocus(editField.form)
			})
		}
		for i := range tempEntry.SecurityQ {
			i := i
			sq := &tempEntry.SecurityQ[i]
			runeAlphabetIterate(&num)
			newFieldsAddedList.AddItem(sq.DisplayName + ":", "[black]" + sq.Value, runeAlphabet[num], func() {
				blankEditFieldForm(sq, &tempEntry.SecurityQ, i, &tempEntry, false)
				pages.ShowPage("/new-editField")
				app.SetFocus(editField.form)
			})
		}
	}

	// If doSwitch is true, then you swap focus to the list of fields
	// already added. If it is false, then you go to /new. 
	switchToNewFieldsList = func(doSwitch bool) {
		blankFieldsAdded()
		if (doSwitch) && (newFieldsAddedList.GetItemCount() > 1) {
			pages.SwitchToPage("/newEntry")
			app.SetFocus(newFieldsAddedList)
		}

		if newFieldsAddedList.GetItemCount() < 2 { // if all the fields are deleted, then:
			newFieldsAddedList.Clear()
			editFieldIndex := newEntry.form.GetButtonIndex("edit field")
			if editFieldIndex > -1 {
				newEntry.form.RemoveButton(editFieldIndex)
				pages.SwitchToPage("/newEntry")
				app.SetFocus(newEntry.form)
			} else {
				switchToError(" AHHHHHHH for some reason the edit field button wasn't added despite a field later trying to be deleted!!!!")
			}
		}
	}

	blankOpen = func(i int) string {
		e := entries[i]
		print := " "

		print += "[" + strconv.Itoa(i) + "] " + e.Name + "\n "
		print += strings.Repeat("-", len([]rune(print))-3) + " \n" // Right now it matches under the letters of title, if at -2 then it goes one out
		if e.Tags != "" {
			print += " tags: " + e.Tags + "\n"
		}
		for _, u := range e.Usernames {
			print += " " + u.DisplayName + ": " + u.Value + "[white]\n"
		}
		for _, p := range e.Passwords {
			print += " " + p.DisplayName + ": [black]" + p.Value + "[white]\n"
		}
		for _, sq := range e.SecurityQ {
			print += " " + sq.DisplayName + ": [black]" + sq.Value + "[white]\n"
		}
		emptyNotes := true
		for _, n := range e.Notes {
			if n != "" {
				emptyNotes = false
				break
			}
		}
		if !emptyNotes {
			blankLines := 0
			print += " notes: "
			for _, n := range e.Notes {
				if n == "" {
					blankLines++
				} else {
					print += strings.Repeat("\n", blankLines)
					print += "\n\t " + n
					blankLines = 0
				}
			}
		}
		print += "\n\n[white]"
		// Following is info about the entry
		print += " in circulation: " + strconv.FormatBool(e.Circulate) + "\n"
		if !e.Modified.IsZero() { // if it's not jan 1, year 1
			print += " date last modified: " + fmt.Sprint(e.Modified.Date()) + "\n"
		}
		if !e.Opened.IsZero() { // if it's not jan 1, year 1
			print += " date last opened: " + fmt.Sprint(e.Opened.Date()) + "\n"
		}
		if !e.Created.IsZero() { // if it's not jan 1, year 1
			print += " date created: " + fmt.Sprint(e.Created.Date())
		}
		entries[i].Opened = time.Now()
		writeFileErr()
		return print
	}

	blankCopen = func(i int) {
		num := 0
		copen.list.Clear()
		e := entries[i]

		copen.list.AddItem("leave /copen " + strconv.Itoa(i), "(takes you back to /home)", runeAlphabet[num], func() {
			clipboard.WriteAll("banana")
			switchToHome()
		})
		runeAlphabetIterate(&num)
		copen.list.AddItem("name:", e.Name, runeAlphabet[num], func() {
			clipboard.WriteAll(e.Name)
		})
		if e.Tags != "" {
			runeAlphabetIterate(&num)
			copen.list.AddItem("tags:", e.Tags, runeAlphabet[num], func() {
				clipboard.WriteAll(e.Tags)
			})
		}
		for _, u := range e.Usernames {
			u := u
			runeAlphabetIterate(&num)
			copen.list.AddItem(u.DisplayName + ":", u.Value, runeAlphabet[num], func() {
				clipboard.WriteAll(u.Value)
			})
		}
		for _, p := range e.Passwords {
			p := p
			runeAlphabetIterate(&num)
			copen.list.AddItem(p.DisplayName + ":", "[black]" + p.Value, runeAlphabet[num], func() {
				clipboard.WriteAll(p.Value)
			})
		}
		for _, sq := range e.SecurityQ {
			sq := sq
			runeAlphabetIterate(&num)
			copen.list.AddItem(sq.DisplayName + ":", "[black]" + sq.Value, runeAlphabet[num], func() {
				clipboard.WriteAll(sq.Value)
			})
		}
		for _, n := range e.Notes {
			n := n
			if n != "" {
				runeAlphabetIterate(&num)
				copen.list.AddItem("note:", n, runeAlphabet[num], func() {
					clipboard.WriteAll(n)
				})
			}
		}
		runeAlphabetIterate(&num)
		copen.list.AddItem("in circulation:", strconv.FormatBool(e.Circulate), runeAlphabet[num], func() {
			clipboard.WriteAll(strconv.FormatBool(e.Circulate))
		})
		if !e.Modified.IsZero() {
			runeAlphabetIterate(&num)
			copen.list.AddItem("date last modifed:", fmt.Sprint(e.Modified.Date()), runeAlphabet[num], func() {
				clipboard.WriteAll(fmt.Sprint(e.Modified.Date()))
			})
		}
		if !e.Opened.IsZero() {
			runeAlphabetIterate(&num)
			copen.list.AddItem("date last opened:", fmt.Sprint(e.Opened.Date()), runeAlphabet[num], func() {
				clipboard.WriteAll(fmt.Sprint(e.Opened.Date()))
			})
		}
		if !e.Created.IsZero() {
			runeAlphabetIterate(&num)
			copen.list.AddItem("date created:", fmt.Sprint(e.Created.Date()), runeAlphabet[num], func() {
				clipboard.WriteAll(fmt.Sprint(e.Created.Date()))
			})
		}
		entries[i].Opened = time.Now()
		writeFileErr()
	}

	// TO DO -- maybe make this tempentry as well?! ask em about it
	blankEditList = func(i int) {
		edit.list.Clear()
		e := &entries[i]
		num := 0

		edit.list.AddItem("leave /edit " + strconv.Itoa(i), "(takes you back to /home)", runeAlphabet[num], func() {
			switchToHome()
		})
		runeAlphabetIterate(&num)
		edit.list.AddItem("name: ", e.Name, runeAlphabet[num], func() {
			info.text.SetText(editFieldInfo)
			blankEditStringForm("name", e.Name, e)
			pages.ShowPage("/editFieldStr")
			app.SetFocus(editField.form)
		})
		if e.Tags != "" {
			runeAlphabetIterate(&num)
			edit.list.AddItem("tags:", e.Tags, runeAlphabet[num], func() {
				info.text.SetText(editFieldInfo)
				blankEditStringForm("tags", e.Tags, e)
				pages.ShowPage("/editFieldStr")
				app.SetFocus(editField.form)
			})
		}
		for i := range e.Usernames {
			i := i
			u := &e.Usernames[i]
			runeAlphabetIterate(&num)
			edit.list.AddItem(u.DisplayName + ":", u.Value, runeAlphabet[num], func() {
				info.text.SetText(editFieldInfo)
				blankEditFieldForm(u, &e.Usernames, i, e, true)
				pages.ShowPage("/edit-editField")
				app.SetFocus(editField.form)
			})
		}
		for i := range e.Passwords {
			i := i
			p := &e.Passwords[i]
			runeAlphabetIterate(&num)

			edit.list.AddItem(p.DisplayName + ":", "[black]" + p.Value, runeAlphabet[num], func() {
				info.text.SetText(editFieldInfo)
				blankEditFieldForm(p, &e.Passwords, i, e, true)
				pages.ShowPage("/edit-editField")
				app.SetFocus(editField.form)
			})
		}
		for i := range e.SecurityQ {
			i := i
			sq := &e.SecurityQ[i]
			runeAlphabetIterate(&num)

			edit.list.AddItem(sq.DisplayName + ":", "[black]" + sq.Value, runeAlphabet[num], func() {
				info.text.SetText(editFieldInfo)
				blankEditFieldForm(sq, &e.SecurityQ, i, e, true)
				pages.ShowPage("/edit-editField")
				app.SetFocus(editField.form)
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
			runeAlphabetIterate(&num)
			edit.list.AddItem("notes:", condensedNotes, runeAlphabet[num], func() {
				info.text.SetText(editFieldInfo)
				blankNewNote(e)
				pages.ShowPage("/newNote")
				app.SetFocus(newNote.form)
			})
		} else {
			runeAlphabetIterate(&num)
			edit.list.AddItem("add notes:", "(none written so far)", runeAlphabet[num], func() {
				info.text.SetText(editFieldInfo)
				blankNewNote(e)
				pages.ShowPage("/newNote")
				app.SetFocus(newNote.form)
			})
		}
		newFieldStr := ""
		if e.Tags == "" {
			newFieldStr += "tags, "
		}
		runeAlphabetIterate(&num)
		edit.list.AddItem("add new field", newFieldStr + "usernames, passwords, security questions", runeAlphabet[num], func() {
			info.text.SetText(editFieldInfo)
			// code copied from blankNewEntry
			blankNewField(e)
			pages.ShowPage("/newField")
			app.SetFocus(newField.form)
		})
		runeAlphabetIterate(&num)
		if e.Circulate { // If it is in circulation, option to opt out
			edit.list.AddItem("remove from circulation", "(not permanant), check /help for info", runeAlphabet[num], func() {
				e.Circulate = false
				switchToEditList(true)
			})

		} else { // If it's not in circulation, option to opt back in
			edit.list.AddItem("add back to circulation", "(not permanant), check /help for info", runeAlphabet[num], func() {
				e.Circulate = true
				switchToEditList(true)
			})
		}
		runeAlphabetIterate(&num)
		edit.list.AddItem("delete entry", "(permanant!!)", runeAlphabet[num], func() {
			info.text.SetText(editFieldInfo)
			blankEditDeleteEntry()
			pages.ShowPage("/editDelete")
			app.SetFocus(editDelete.form)
		})
	}

	// Takes in an extra boolean to know if its from /edit or /new,
	// in order to know where to go back to.
	blankEditFieldForm = func(f *Field, fieldArr *[]Field, index int, e *entry, edit bool) {
		editField.form.Clear(true)
		tempField.DisplayName = f.DisplayName
		tempField.Value = f.Value

		editField.form.
			AddInputField("display name:", tempField.DisplayName, 40, nil, func(input string) {
				tempField.DisplayName = input
			}).
			AddInputField("value:", tempField.Value, 40, nil, func(input string) {
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
				} else { // TO DO - re-write the following error? it shouldn't ever get to it though!
					switchToError(" AHHHHH! the array given to blankEditFieldForm is nil \n and it shouldnt be!! or the index is -1 which it also shouldn't be")
				}
			})
	}

	// For editing the name or tags, nothing else
	blankEditStringForm = func(display, value string, e *entry) {
		if (display != "name") && (display != "tags") {
			switchToError(" Error, unexpected input\n blackEditStringForm can only change name or tags")
			return
		}
		editField.form.Clear(true)
		tempDisplay := display
		tempValue := value
		editField.form.
			AddInputField(tempDisplay + ":", tempValue, 50, nil, func(changed string) {
				tempValue = changed
			}).
			AddButton("save", func() {
				if display == "name" {
					e.Name = tempValue
				} else {
					e.Tags = tempValue
				}
				switchToEditList(true)
			}).
			AddButton("quit", func() {
				switchToEditList(false)
			})
		// Can only delete tags, not the name
		if display == "tags" {
			editField.form.AddButton("delete", func() {
				e.Tags = ""
				switchToEditList(true)
			})
		}
	}

	switchToEditList = func(modified bool) {
		if writeFileErr() {
			if modified {
				entries[indexSelected].Modified = time.Now()
			}
			blankEditList(indexSelected)
			pages.SwitchToPage("/edit")
			app.SetFocus(edit.list)
			info.text.SetText(editInfo)
		}
	}

	blankEditDeleteEntry = func() {
		editDelete.form.Clear(true)
		editDelete.form.SetButtonsAlign(tview.AlignCenter)
		editDelete.form.
			AddButton("save", func() {
				switchToEditList(false)
			}).
			AddButton("delete", func() { // deletes element from slice, slower version, keeps everything else in order, copied the code from a website lol
				copy(entries[indexSelected:], entries[indexSelected+1:])
				entries[len(entries)-1] = entry{} // TO DO -- ask why this is here?
				entries = entries[:len(entries)-1]
				if writeFileErr() {
					switchToHome()
				}
			})
	}

	// openCopen is either going to be "/pick", "/picc", or "/flist str"
	blankPickList = func(openCopen string, indexes []int) {
		printCommand := openCopen

		if len([]rune(openCopen)) > 5 { // if is /flist str or at all /flist
			printCommand = "/flist\n ------ \n"
		} else {
			printCommand += "\n ----- \n"
		}

		info.text.SetText(" " + printCommand + pickInfo)
		num := 0
		pick.list.Clear()
		pick.list.AddItem("leave " + openCopen, "(takes you back to /home)", runeAlphabet[num], func() {
			switchToHome()
		})
		for _, i := range indexes {
			num++
			if num == len(runeAlphabet) {
				num = 0
			}
			i := i
			if entries[i].Circulate {
				pick.list.AddItem("[" + strconv.Itoa(i) + "] " + entries[i].Name, "tags: " + entries[i].Tags, runeAlphabet[num], func() {
					if openCopen == "/pick" { // to transfer to /open #
						// following code copied from commandLineActions function
						pages.SwitchToPage("/open")
						app.SetFocus(commandLine.input)
						lookRightCommandLinePlaceholder()
						open.text.SetText(blankOpen(i)) // taking input, just to be safe smile -- can change that in future
					} else { // to transfer to /copen # (for both /picc and /flist)
						// following code copied from commandLineActions function
						app.SetFocus(copen.list)
						app.EnableMouse(false)
						pages.SwitchToPage("/copen")
						blankCopen(i)
					}
				})
			}
		}
	}

	blankComp = func(i1 int, i2 int) string {
		e1 := entries[i1]
		e2 := entries[i2]
		print := " "

		print += "/comp: " + "[" + strconv.Itoa(i1) + "] " + e1.Name + " and " + "[" + strconv.Itoa(i2) + "] " + e2.Name + "\n "
		print += strings.Repeat("-", len([]rune(print))-3) + "\n\n"

		print += compPass(e1, e2)

		return print
	}

	// ------------------------------------------------ //
	//     setting up the flexes, grids, pages :)       //
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
			AddItem(newNote.grid, 0, 6, false). // 4 fits 3 input + buttons, 5 fits 4 input + buttons
			AddItem(nil, 0, 1, false), 0, 5, false)

	newEntry.flex.SetDirection(tview.FlexRow).
		AddItem(newEntry.form, 0, 2, false).
		AddItem(newFieldsAddedList, 0, 3, false) // 1:2 is the maximum

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
			AddItem(nil, 0, 2, false), 0, 1, false).
		AddItem(nil, 0, 1, false)

	list.flex.SetDirection(tview.FlexRow).
		AddItem(list.title, 0, 1, false).
		AddItem(list.text, 0, 8, false)

	// Add each primitive to its perspective grid
	grider(commandLine.input, commandLine.grid)
	grider(info.text, info.grid)
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
	grider(comp.text, comp.grid)
	grider(reused.text, reused.grid)

	// All the different pages are added here.
	// The order in which the pages are added matters.
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
		AddPage("/edit-editField", editEditFieldFlex, true, false).
		AddPage("/comp", comp.grid, true, false).
		AddPage("/reused", reused.grid, true, false)

	// Sets up the flex row of the left side, top is the pages
	// bottom is the commandLine.input
	// Ratio of 8:1 is the maximum that it can be (9:1 is the same
	// as 8:1) for 26x78.
	// ratio of 9:1 is the maximum on 28x84 grid (10:1 is the same)
	flexRow.
		AddItem(pages, 0, 9, false).
		AddItem(commandLine.grid, 0, 1, false)

	// Left and right sides of pass
	flex.
		AddItem(flexRow, 0, 14, false).
		AddItem(info.grid, 0, 3, false)

	passBoxPages.
		AddPage("passBox", passBox, true, true).
		AddPage("passErr", passErr.grid, true, false)

	passPages.
		AddPage("passInput", passFlex, true, true).
		AddPage("passManager", flex, true, false) // the page passManager is the flex of pass

	if err := app.SetRoot(passPages, true).SetFocus(password.input).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

// This is the function used to put any type of primitive
// into a grid, to make a border.
func grider(prim tview.Primitive, grid *tview.Grid) {
	grid.AddItem(prim, 0, 0, 1, 1, 0, 0, false)
}

// This finds all of the entries with str in tags or names
func findEntries(entries []entry, str string) (string, string) {
	indexes := []int{}
	str = strings.ToLower(str)
	for i, e := range entries {
		if (strings.Contains(strings.ToLower(e.Name), str)) || (strings.Contains(strings.ToLower(e.Tags), str)) {
			indexes = append(indexes, i)
		}
	}
	// Trims out the /find str that prints so it won't go all the
	// way to the other side of the text box. It still used the full
	// str in the search. When printed, it will add a ... if it was trimmed
	trimmedStr := str

	if len([]byte(trimmedStr)) > 59 {
		trimmedStr = trimmedStr[:56]
		trimmedStr += "..."
	}

	if len(indexes) > 0 {
		return listEntries(entries, indexes, " /find " + trimmedStr + " \n " + strings.Repeat("-", len([]rune(trimmedStr)) + 6), true)
	} else {
		return " /find " + trimmedStr + " \n " + strings.Repeat("-", len([]rune(trimmedStr)) + 6), " no entries found"
	}
}

// This function formats some/all of the entries into
// thirds in a string. The str taken in is put at the top,
// and is for example: " /find str \n-----..." or " /list \n -----"
// The bool taken in differentiates it from /list or /find, to show
// or not to show the ones that are not in ciculation.
// If it is in /find but the entry isn't in circulation, it will
// type (rem) right before the entry name. Ex: [1] (rem) Twitter
func listEntries(entries []entry, indexes []int, str string, showOld bool) (string, string) {
	printStr := ""
	printEntries := []entry{}

	if showOld {
		for _, i := range indexes {
			printEntries = append(printEntries, entries[i]) // equivalent to entries[i] is entries[indexes[i]]
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
	floatThird := float64(len(indexes)) / 3.0

	if floatThird < 21.0 {
		floatThird = 21.0
	} else if floatThird > float64(int(floatThird)) {
		floatThird ++
	}

	third := int(floatThird)

	for i := 0; i < third; i++ {
		if i >= len(indexes) {
			break
		}
		printStr += " " + indexName(indexes[i], entries)
		if len(indexes) > i+third {
			printStr += indexName(indexes[i+third], entries)
		}
		if len(indexes) > i+third+third {
			printStr += indexName(indexes[i+third+third], entries)
		}
		if i != third-1 { // so it doesn't do it on the last one
			printStr += "\n"
		}
	}
	return str, printStr // first string is the title, second is the body of the text
}

// This returns from a single index from entries to
// " [0] twitterDEMO       ", with those exact spaces/number of
// characters in order to make a good column shape. Used in /list
func indexName(index int, entries []entry) string {
	str := "[" + strconv.Itoa(index) + "] "

	if !entries[index].Circulate { // if out of circulation
		str += "(rem) "
	}

	str += entries[index].Name
	len := len([]rune(str))

	if len > 21 { // Trims it if it's over the character limit
		str = str[0:21]
		str += " "
	} else {
		str += strings.Repeat(" ", 22-len)
	}
	return str
}

// Looks for duplicates between the passwords and security
// question answers of two entries, e1 and e2
func compPass(e1 entry, e2 entry) string {
	compared := ""

	compMap := make(map[string][]reusedPass) // reusedPass is a struct

	// adding all passwords and securityQs to the map
	for _, p := range e1.Passwords {
		compMap[p.Value] = append(compMap[p.Value], reusedPass{displayName: p.DisplayName, entryName: e1.Name})
	}
	for i, s := range e1.SecurityQ {
		compMap[s.Value] = append(compMap[s.Value], reusedPass{displayName: "security question " + strconv.Itoa(i), entryName: e1.Name})
	}
	for _, p := range e2.Passwords {
		compMap[p.Value] = append(compMap[p.Value], reusedPass{displayName: p.DisplayName, entryName: e2.Name})
	}
	for i, s := range e2.SecurityQ {
		compMap[s.Value] = append(compMap[s.Value], reusedPass{displayName: "security question " + strconv.Itoa(i), entryName: e2.Name})
	}

	// going through the map and looking at duplicates
	for _, reusedStruct := range compMap {
		if len(reusedStruct) == 2 { // if same pass twice, most common
			compared += " " + reusedStruct[0].entryName + "'s " + reusedStruct[0].displayName + " = " + reusedStruct[1].entryName + "'s " + reusedStruct[1].displayName + "\n"
		} else if len(reusedStruct) > 2 { // less common
			for i, r := range reusedStruct {
				compared += " " + r.entryName + "'s " + r.displayName
				if (i + 1) < len(reusedStruct) { // on the not last time through
					compared += " =\n"
				} else { // happens on the last time through
					compared += "\n"
				}
			}
		}
	}

	if compared == "" {
		if (len(e1.Passwords) < 1) && (len(e1.SecurityQ) < 1) {
			compared = " " + e1.Name + " has no passwords or security questions" + "\n"
		}
		if len(e2.Passwords) < 1 && (len(e2.SecurityQ) < 1){
			compared += " " + e2.Name + " has no passwords or security questions" + "\n"
		}
		compared += "\n Therefore, there are no passwords in common!"
	}

	return compared
}

// Looks for password duplicates between all of the passwords,
// using a map of slices of structs
func reusedAll(entries []entry) string {
	print := ""

	reused := make(map[string][]reusedPass) // reusedPass is a struct

	for i, e := range entries {
		for _, p := range e.Passwords {
			reused[p.Value] = append(reused[p.Value], reusedPass{displayName: p.DisplayName, entryName: e.Name, entryIndex: i})
		}
		for iSq, s := range e.SecurityQ {
			reused[s.Value] = append(reused[s.Value], reusedPass{displayName: "security question " + strconv.Itoa(iSq), entryName: e.Name, entryIndex: i})
		}
	}

	for pass, reusedStruct := range reused {
		if len(reusedStruct) > 1 { // if there's more than one entry in the list of entries for password
			print += " [darkslategray]" + pass + "[white]:\n"
			for _, r := range reusedStruct {
				print += " [" + strconv.Itoa(r.entryIndex) + "] " + r.entryName + "'s " + r.displayName + "\n"
			}
			print += "\n"
		}
	}

	if print == "" {
		return " There are no reused passwords anywhere!?\n Good job!"
	}

	print = print[:len([]rune(print))-2] // gets rid of the last \n\n

	return print
}

// Used in /find. Returns a slice of ints of all indexes
// with 'inputStr' in the tags or name.
func findIndexes(entries []entry, inputStr string) []int {
	indexes := []int{}
	str := strings.ToLower(inputStr)
	for i, e := range entries {
		if (strings.Contains(strings.ToLower(e.Name), str)) || (strings.Contains(strings.ToLower(e.Tags), str)) {
			indexes = append(indexes, i)
		}
	}
	return indexes
}

// This is called in /test in order to add all of the entries
// to a single string.
func testAllFields(entries []entry) string {
	allValues := " Test of all fields that are known:"
	for _, e := range entries {
		allValues += "\n\n " + fmt.Sprint(e)
	}
	return allValues
}

// If this is changed, also change createEncr.go and changeKey.go
// If it fails to write to the file then it returns
// a string with the errors, else it returns ""
func writeToFile(entries []entry, ciphBlock cipher.Block) string {
	output, marshErr := yaml.Marshal(entries)

	if marshErr != nil {
		return " Error in yaml.Marshal\n\n " + marshErr.Error()
	}

	encryptedOutput := encrypt.Encrypt(output, ciphBlock)
	// conventions of writing to a temp file is write to .tmp
	writeErr := os.WriteFile(encrypt.FileName + ".tmp", encryptedOutput, 0600) // 0600 is the permissions that only this user can read/write/excet to this file
	
	if writeErr != nil {
		return " Error in os.WriteFile\n\n " + writeErr.Error()
	}

	os.Rename(encrypt.FileName + ".tmp", encrypt.FileName) // Only will do this if the previous writing to a file worked, keeps it safe.

	return ""
}

// If this is changed, also change changeKey.go
// If it fails to read from the file then it returns
// a string with the errors, else it returns ""
func readFromFile(entries *[]entry, ciphBlock cipher.Block) string {
	input, inputErr := os.ReadFile(encrypt.FileName)

	if inputErr != nil {
		return " Error in os.ReadFile\n Make sure that a file named " + encrypt.FileName + " exists.\n There isn't one, run createEncr.go\n\n " + inputErr.Error()
	}

	decryptedInput := encrypt.Decrypt(input, ciphBlock)
	unmarshErr := yaml.Unmarshal(decryptedInput, &entries)

	if unmarshErr != nil {
		return " Error in yaml.Unmarshal\n Make sure you write the correct password.\n\n " + unmarshErr.Error()
	}
	return ""
}
