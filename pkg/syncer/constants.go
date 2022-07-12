package syncer

import (
	"fmt"
	"log"
	"regexp"
	"strings"
)

var (
	// TODO: expose (some of) these as flags
	foldersToIgnore = []string{
		".pytest_cache",
		".git",
		".idea",
		"node_modules",
		// e.g. these below are probably particular to my use case only
		".teamcity",
		".bash_history",
		".venv",
		".virtualenv",
		"venv",
		"coverage",
		"test_results",
	}
	folderIgnoreExp *regexp.Regexp
	// TODO: expose (some of) these as flags
	filesToIgnore = []string{
		".pyc",
		".tmp",
	}
	fileIgnoreExp *regexp.Regexp
)

func init() {
	testExp := func(exp *regexp.Regexp, testValue string) {
		if exp.MatchString(testValue) {
			return
		}

		log.Fatalf("exp=%#+v could not match testValue=%#+v", exp.String(), testValue)
	}

	rawFolderIgnoreExp := ""
	for _, folder := range foldersToIgnore {
		rawFolderIgnoreExp += fmt.Sprintf(
			"(.*(/|^)%v(/|$).*)|",
			folder,
		)
	}
	rawFolderIgnoreExp = strings.Trim(rawFolderIgnoreExp, "|")
	folderIgnoreExp = regexp.MustCompile(rawFolderIgnoreExp)

	folderIgnoreTestValues := []string{
		"/node_modules",
		"/node_modules/",
		"/node_modules/something",
		"/node_modules/something/",
		"/something/node_modules",
		"/something/node_modules/",
		"/something/node_modules/something",
		"/something/node_modules/something/",
	}

	for _, testValue := range folderIgnoreTestValues {
		testExp(folderIgnoreExp, testValue)
	}

	rawFileIgnoreExp := ""
	for _, file := range filesToIgnore {
		rawFileIgnoreExp += fmt.Sprintf(
			"(.*\\w+%v$)|",
			file,
		)
	}
	rawFileIgnoreExp = strings.Trim(rawFileIgnoreExp, "|")
	fileIgnoreExp = regexp.MustCompile(rawFileIgnoreExp)

	fileIgnoreTestValues := []string{
		"some_file.pyc",
		"/some_file.pyc",
		"/something/some_file.pyc",
	}

	for _, testValue := range fileIgnoreTestValues {
		testExp(fileIgnoreExp, testValue)
	}
}
