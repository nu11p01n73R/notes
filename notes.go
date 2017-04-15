package main

import (
	"fmt"
	"os"
	"os/user"
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

	return noteDir, nil
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	fileName, err := initNotes()
	checkError(err)

	fmt.Println(fileName)
}
