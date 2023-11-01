package common

import (
	"errors"
	"log"
	"os"
	"strings"
)

/*
Python Equivalent of pattern*times: repeat("a", 3) --> "aaa"
*/
func repeat(pattern string, times int) string {

	out := ""
	for z := 0; z < times; z++ {
		out += pattern
	}
	return out
}

/*
Creates the directory if it  doesn't exist intially
*/
func createDirIfDontExist(path string) error {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(path, 0777)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

/*
It outputs the list of filepaths of all the yaml-files present in a directory, recursively
*/
func RecursiveListYamls(curFolder string) (yamlfiles []string) {
	folderContent, _ := os.ReadDir(curFolder)
	for _, files := range folderContent {
		if files.Type().IsDir() {
			if files.Name() == "tests" { //	Don't list the test yamls (if any)
				continue
			}
			returnedYamlFiles := RecursiveListYamls(curFolder + "/" + files.Name())
			yamlfiles = append(yamlfiles, returnedYamlFiles...)
		} else {
			fileName := files.Name()
			if len(fileName) > 5 {
				if fileName[len(fileName)-5:] == ".yaml" {
					yamlfiles = append(yamlfiles, curFolder+"/"+fileName)
				}
			}

		}
	}
	return
}

func handleMultiLineStrings(input string) string {
	/* There are different ways to handle Multi-Line-Strings
	Method-1: Usage of "Str1" + "Str2"
	Replacing "\n" with `\n " + \n "`
	"Str1\nStr2" :
					Str1\n" +
					"Str2
	*/
	input = strings.ReplaceAll(input, "\n", "\\n\" + \n \"")
	return "\"" + input + "\""

	// Method-2: Usage of `` for Raw-String Literal: To be Decided if Method 1 has any limitations

}
