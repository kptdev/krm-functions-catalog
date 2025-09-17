package testutil

import (
	"io/ioutil"
)

func Example_testWithResourceList() {
	rl, err := ResourceListFromFile("testdata/resourcelist.yaml")
	if err != nil {
		// check error
	}

	// do something with the ResourceList

	actualYaml, err := rl.ToYAML()
	if err != nil {
		// check error
	}

	rlExpected, err := ResourceListFromFile("testdata/expected_resourcelist.yaml")
	if err != nil {
		// check error
	}
	expectedYaml, err := rlExpected.ToYAML()
	if err != nil {
		// check error
	}

	// Compare expectedYaml and actualYaml
	if string(expectedYaml) != string(actualYaml) {
		// do something
	}
}

func Example_testWithDirectory() {
	rl, err := ResourceListFromDirectory("testdata/resources/input", "testdata/resources/input/fn-config.yaml")
	if err != nil {
		// check error
	}

	// Do something with the ResourceList

	// Create a temporary directory and write the resources back as files.
	tempDir, err := ioutil.TempDir("", "kyaml-fn-sdk-*")
	err = ResourceListToDirectory(rl, tempDir)
	if err != nil {
		// check error
	}

	// Read the resources again from disk as a ResourceList.
	rlActual, err := ResourceListFromDirectory(tempDir, "")
	if err != nil {
		// check error
	}
	actualYaml, err := rlActual.ToYAML()
	if err != nil {
		// check error
	}

	rlExpected, err := ResourceListFromDirectory("testdata/resources/expected", "")
	if err != nil {
		// check error
	}
	expectedYaml, err := rlExpected.ToYAML()
	if err != nil {
		// check error
	}

	// Compare expectedYaml and actualYaml. The 2 directories are identical if
	// they can produce the same ResourceLists.
	if string(expectedYaml) != string(actualYaml) {
		// do something
	}
}
