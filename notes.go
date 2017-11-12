package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/nu11p01n73R/fuz/src"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"time"
)

// Path to the parent directory
// where .notes tree is saved
var noteDir string

// Generate a random string of length n
// Gets a random number of 63 bits. Takes
// the last 6 bits of the random number.
// Checks if the 6 bits, is less than the
// length of key string, if yes add the
// character at the index to output string.
// Right shift the 6 bits.
// If the 6 bits is greater than the length,
// right shift one byte out.
func getRandomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	random := make([]byte, n)
	mask := 1<<6 - 1
	num := rand.Int63()

	for i := 0; i < n; {
		index := int(num) & mask
		if index < len(letterBytes) {
			random[i] = letterBytes[index]
			i++
			num = num >> 6
		} else {
			num = num >> 1
		}

		if num == 0 {
			num = rand.Int63()
		}
	}

	return string(random)
}

// Get the home directory of the current
// user.
// If getting fails, return error
func getUserHome() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	return usr.HomeDir, nil
}

// Find the difference between list1 and list2
// Return the differnece list.
func listDiff(list1, list2 []string) []string {
	hash := map[string]bool{}
	for _, str := range list2 {
		hash[str] = true
	}

	output := []string{}
	for _, str := range list1 {
		if _, found := hash[str]; !found {
			output = append(output, str)
		}
	}
	return output
}

// Checks if a directory exists. If not,
// create one. Returns an error object.
func createDir(fileName string) error {
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		os.Mkdir(fileName, 0755)
		return nil
	}

	return err
}

// Copy file from src to dest.
// Returns error
func copyFile(src, dest string) error {
	input, err := os.Open(src)
	if err != nil {
		return err
	}
	defer input.Close()

	output, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer output.Close()

	_, err = io.Copy(output, input)
	return err
}

// Initialize the notes app.
// Creats the ./notes directory if not
// exists with data and tags directories
// in them.
// The note directory is created at users
// home directory
// TODO Provide functionality to add a
// differnt notes directory through
// options or environment variables.
func initNotes() (string, error) {
	home, err := getUserHome()
	if err != nil {
		return "", err
	}

	noteDir := home + "/.notes"
	err = createDir(noteDir)
	if err != nil {
		return "", err
	}

	dataDir := noteDir + "/data"
	err = createDir(dataDir)
	if err != nil {
		return "", err
	}

	tagDir := noteDir + "/tags"
	err = createDir(tagDir)
	if err != nil {
		return "", err
	}

	return noteDir, nil
}

// Return a template for empty note.
func newNoteTemplate() []byte {
	content := `[TITLE]

[TAGS]

[CONTENT]
`

	return []byte(content)
}

// Create a file with newFile name,
// and write empty note template to it.
// Returns error
func createNote(newFile string) error {
	file, err := os.OpenFile(newFile, os.O_CREATE|os.O_WRONLY, 0754)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(newNoteTemplate())
	return err
}

// Open the file in the default editor.
// Return error
// TODO error on if  EDITOR is not configured
func openEditor(file string) error {
	editor := os.Getenv("EDITOR")

	cmd := exec.Command(editor, file)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	err := cmd.Run()
	return err
}

// Normaize a string,
//	Converts to lowerCase
//	Trims of spaces at start and end
//	Replace spaces with _
func normalizeString(str string) string {
	normalized := strings.Trim(strings.ToLower(str), " ")
	return strings.Replace(normalized, " ", "_", -1)
}

// Parse the [TAGS] section of the notes file.
// Split the tags line to tag elements and normalize them.
// Returns list of normalized tags.
func parseTags(tags string) []string {
	output := []string{}
	for _, tag := range strings.Split(tags, ",") {
		output = append(output, normalizeString(tag))
	}
	return output
}

// Parse the notes files.
// Extract the title and list of tags from
// the file.
func parseNote(note string) (string, []string, error) {
	var key, title string
	tags := []string{}

	file, err := os.Open(note)
	if err != nil {
		return title, tags, err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "[") &&
			strings.HasSuffix(line, "]") {
			key = line
			continue
		}

		switch key {
		case "[TITLE]":
			title = line
			break
		case "[TAGS]":
			tags = append(tags, parseTags(line)...)
			break
		}
	}

	return normalizeString(title), tags, nil
}

// Save the final note file to fileName by moving
// contents from tempFile to fileName
// The file is saved in .notes/data/date folder.
func saveNewNote(tempFile, fileName string) (string, error) {
	today := time.Now().Format("20060102")
	note := fmt.Sprintf("%s/%s.md", today, fileName)

	outputDir := fmt.Sprintf("%s/data/%s", noteDir, today)
	createDir(outputDir)

	outputFile := fmt.Sprintf("%s/%s.md", outputDir, fileName)
	err := os.Rename(tempFile, outputFile)

	return note, err
}

// Save the  file names in corresponding tagFiles
// If file exists, append to the file. Else create
// a new file and save the note file it.
func saveTag(tagFile, noteFile string) error {
	file, err := os.OpenFile(tagFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0754)
	if err != nil {
		return err
	}

	_, err = file.Write([]byte(noteFile + "\n"))
	return err
}

// Save the tags for noteFile from
// the note in corresponding tag files.
// TODO add go routines to add tags
func indexTags(noteFile string, tags []string) error {
	tagDir := noteDir + "/tags"

	for _, tag := range tags {
		tagFile := fmt.Sprintf("%s/%s", tagDir, tag)
		err := saveTag(tagFile, noteFile)
		if err != nil {
			return err
		}
	}
	return nil
}

// Create a new note.
// A temp file with random name is created
// with empty template. Open the editor for
// temp file.
// Parse and save the notes on exit.
func newNote() error {
	newNote := fmt.Sprintf("%s/.%s.md", noteDir, getRandomString(10))

	err := createNote(newNote)
	if err != nil {
		return err
	}

	err = openEditor(newNote)
	if err != nil {
		return err
	}

	title, tags, err := parseNote(newNote)
	if err != nil {
		return err
	}

	// Ignore the file with no title
	// assuming that the file is empty
	if len(title) == 0 {
		return errors.New("Empty title. Skipping the note...")
	}

	err = save(newNote, title, tags)
	return err
}

// Save the note on tempFile to final file.
func save(tempFile, title string, tags []string) error {
	noteFile, err := saveNewNote(tempFile, title)
	if err != nil {
		return err
	}

	err = indexTags(noteFile, tags)
	return err
}

// Remove noteFile from  a tagFile.
// The file will not be removed if it is emptied.
func removeTag(tagFile, noteFile string) error {
	tempFileName := tagFile + ".tmp"
	tempFile, err := os.OpenFile(tempFileName, os.O_CREATE|os.O_WRONLY, 0754)
	if err != nil {
		return err
	}
	defer tempFile.Close()

	file, err := os.Open(tagFile)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	writter := bufio.NewWriter(tempFile)
	for scanner.Scan() {
		line := scanner.Text()
		if line != noteFile {
			_, err := writter.Write([]byte(line + "\n"))
			if err != nil {
				return err
			}
		}
	}
	err = writter.Flush()
	if err != nil {
		return err
	}

	err = os.Rename(tempFileName, tagFile)
	return err
}

// Remove all tags of notFile
// TODO add go routine to remove the files.
func deindexTags(noteFile string, tags []string) error {
	tagDir := noteDir + "/tags"

	for _, tag := range tags {
		tagFile := fmt.Sprintf("%s/%s", tagDir, tag)
		err := removeTag(tagFile, noteFile)
		if err != nil {
			return err
		}
	}
	return nil
}

// Remove a note of name  note.
// The data file in data directory will
// be removed.
// The filename will be removed from all
// the index files.
func removeNote(note string) error {
	noteFile := fmt.Sprintf("%s/data/%s", noteDir, note)

	_, err := os.Stat(noteFile)
	if err != nil {
		return errors.New("Cannot find note.")
	}

	_, tags, err := parseNote(noteFile)
	if err != nil {
		return err
	}

	err = deindexTags(note, tags)
	if err != nil {
		return err
	}

	err = os.Remove(noteFile)
	return err
}

// Edit a existing note of name note.
// The note name can be
//	Starting from data/date/name
//	Fully qualified name of the file.
// The file is moved to a tempFile and
// opened in editor
// The file is moved back to data folder
// in new name or already existing name
// depending on if title was edited.
// If [TAGS] was edited, the difference
// is identied, which is either added
// or removed from the tag files.
func editNote(note string) error {
	var err error
	var noteFile string

	prefix := fmt.Sprintf("%s/data/", noteDir)
	if strings.HasPrefix(note, noteDir) {
		noteFile = note
		note = strings.TrimLeft(note, prefix)
	} else {
		noteFile = fmt.Sprintf("%s%s", prefix, note)
	}

	tempFile := fmt.Sprintf("%s/.%s.md", noteDir, getRandomString(10))
	err = copyFile(noteFile, tempFile)
	if err != nil {
		return err
	}

	openEditor(tempFile)

	newTitle, newTags, err := parseNote(tempFile)
	if err != nil {
		return err
	}

	oldTitle, oldTags, err := parseNote(noteFile)
	if err != nil {
		return err
	}

	if oldTitle != newTitle {
		err = save(tempFile, newTitle, newTags)
		if err != nil {
			return err
		}

		err = deindexTags(note, oldTags)
		if err != nil {
			return err
		}

		err = os.Remove(noteFile)
	} else {
		toAdd := listDiff(newTags, oldTags)
		toRemove := listDiff(oldTags, newTags)

		if len(toAdd) > 0 || len(toRemove) > 0 {
			indexTags(note, toAdd)
			deindexTags(note, toRemove)
		}
		os.Rename(tempFile, noteFile)

		err = nil
	}
	return err
}

// List all the current notes in the
// noteDir using Fuz
func listNotes() error {
	dataDir := fmt.Sprintf("%s/data", noteDir)
	cmd := exec.Command("./notes", "edit")

	fuz.Fuz(dataDir, "Notes", cmd)
	return nil
}

// Parse the command line argument
// to identify the command to be used.
func parseCommands() error {
	if len(os.Args) == 1 {
		return errors.New("Not enough arguments")
	}

	var err error
	cmd := os.Args[1]
	switch cmd {
	case "add":
		err = newNote()
		break
	case "remove":
		if len(os.Args) != 3 {
			err = errors.New("No note file specified")
			break
		}
		err = removeNote(os.Args[2])
		break
	case "edit":
		err = editNote(os.Args[2])
		break
	case "search":
		err = listNotes()
		break
	default:
		err = errors.New("Unknown command")
	}
	return err
}

// Handle the error.
func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	var err error
	noteDir, err = initNotes()
	checkError(err)

	err = parseCommands()
	checkError(err)
}
