package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"time"
)

func createDir(fileName string) error {
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		os.Mkdir(fileName, 0755)
		return nil
	}

	return err
}

func initNotes() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	noteDir := usr.HomeDir + "/.notes"
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

[DESCRIPTION]

[TAGS]
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

func saveNewNote(noteDir, fileName string) (string, error) {
	today := time.Now().Format("20060102")

	outputDir := fmt.Sprintf("%s/data/%s", noteDir, today)
	createDir(outputDir)
	outputFile := fmt.Sprintf("%s/%s", outputDir, fileName)

	err := os.Rename(noteDir+"/.new", outputFile)
	return outputFile, err
}

func saveTag(tagFile, noteFile string) error {
	file, err := os.OpenFile(tagFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0754)
	if err != nil {
		return err
	}

	_, err = file.Write([]byte(noteFile + "\n"))
	return err
}

func indexTags(noteDir, noteFile string, tags []string) error {
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

func newNote(noteDir string) error {
	newNote := noteDir + "/.new"
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

	noteFile, err := saveNewNote(noteDir, title)
	if err != nil {
		return err
	}

	err = indexTags(noteDir, noteFile, tags)
	return err
}

func parseCommands(noteDir string) error {
	if len(os.Args) == 1 {
		return errors.New("Not enough arguments")
	}

	var err error
	cmd := os.Args[1]
	switch cmd {
	case "add":
		err = newNote(noteDir)
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
	noteDir, err := initNotes()
	checkError(err)

	err = parseCommands(noteDir)
	checkError(err)
}
