package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
)

func main() {
	filenames := []string{
		"static/index.html",
	}
	result := GenerateStatic(filenames)
	ioutil.WriteFile("static.go", result, 0644)
}

func GenerateStatic(filenames []string) []byte {
	result := "//GENERATED - DON'T EDIT\npackage main\nvar Files = map[string]string{\n"
	for _, filename := range filenames {
		file, err := ioutil.ReadFile(filename)
		if err == nil {
			quoted := Quote(file)
			result += fmt.Sprintf(`"%s": %s`, filename, quoted)
		}
	}
	return []byte(result + "}")
}

func Quote(data []byte) string {
	result := new(bytes.Buffer)
	result.WriteByte('"')
	for _, b := range data {
		fmt.Fprintf(result, "\\x%02x", b)
	}
	result.WriteByte('"')
	return result.String()
}
