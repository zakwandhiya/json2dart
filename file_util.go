package main

import (
	"fmt"
	"os"
)

func writeDartFile(fileContent string, fileName string) {
	fileOutputName := formatString("./%s_model.dart", fileName)

	f, err := os.Create(fileOutputName)
	if err != nil {
		panic(err)
	}

	_, err = f.WriteString(fileContent)
	if err != nil {
		panic(err)
	}

	err = f.Close()
	if err != nil {
		panic(err)
	} else {
		fmt.Printf("file %s written successfully\n", fileOutputName)
	}
}

func readJsonFile(jsonFileUri string) *os.File {

	jsonFile, err := os.Open(jsonFileUri)
	if err != nil {
		fmt.Println("file open error: ", err)
		os.Exit(1)
	}

	fmt.Println(formatString("Successfully Opened %s", jsonFile.Name()))

	return jsonFile
}
