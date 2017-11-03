package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/nu11p01n73R/fuz/src"
	"io"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"time"
)

var noteDir string

func getUserHome() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	return usr.HomeDir, nil
}

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

func createDir(fileName string) error {
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		os.Mkdir(fileName, 0755)
		return nil
	}

	return err
}

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

func newNoteTemplate() []byte {
	content := `[TITLE]

[TAGS]

[CONTENT]
`

	return []byte(content)
}

func createNote(newFile string) error {
	file, err := os.OpenFile(newFile, os.O_CREATE|os.O_WRONLY, 0754)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(newNoteTemplate())
	return err
}

func openEditor(file string) error {
	editor := os.Getenv("EDITOR")

	cmd := exec.Command(editor, file)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	err := cmd.Run()
	return err
}

func normalizeString(str string) string {
	normalized := strings.Trim(strings.ToLower(str), " ")
	return strings.Replace(normalized, " ", "_", -1)
}

func parseTags(tags string) []string {
	output := []string{}
	for _, tag := range strings.Split(tags, ",") {
		output = append(output, normalizeString(tag))
	}
	return output
}

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

func saveNewNote(fileName string) (string, error) {
	today := time.Now().Format("20060102")
	note := fmt.Sprintf("%s/%s.md", today, fileName)

	outputDir := fmt.Sprintf("%s/data/%s", noteDir, today)
	createDir(outputDir)

	outputFile := fmt.Sprintf("%s/%s.md", outputDir, fileName)
	err := os.Rename(noteDir+"/.new.md", outputFile)

	return note, err
}

func saveTag(tagFile, noteFile string) error {
	file, err := os.OpenFile(tagFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0754)
	if err != nil {
		return err
	}

	_, err = file.Write([]byte(noteFile + "\n"))
	return err
}

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

func newNote() error {
	newNote := noteDir + "/.new.md"
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

	err = save(title, tags)
	return err
}

func save(title string, tags []string) error {
	noteFile, err := saveNewNote(title)
	if err != nil {
		return err
	}

	err = indexTags(noteFile, tags)
	return err
}

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

func remove(note string, tags []string) error {
	err := deindexTags(note, tags)
	if err != nil {
		return err
	}

	err = os.Remove(note)
	return err
}

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

	tempFile := noteDir + "/.new.md"
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
		err = save(newTitle, newTags)
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

func listNotes() error {
	dataDir := fmt.Sprintf("%s/data", noteDir)
	cmd := exec.Command("./notes", "edit")

	fuz.Fuz(dataDir, "Notes", cmd)
	return nil
}

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

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	var err error
	noteDir, err = initNotes()
	checkError(err)

	err = parseCommands()
	checkError(err)
}
