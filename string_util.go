package main

import (
	"fmt"
	"path"
	"regexp"
	"strings"
)

func getFileNameWithoutExtension(fileName string) string {
	return strings.TrimSuffix(path.Base(fileName), path.Ext(fileName))
}

var link = regexp.MustCompile("(^[A-Za-z])|_([A-Za-z])")

func toCamelCase(str string) string {
	preResult := link.ReplaceAllStringFunc(str, func(s string) string {
		return strings.ToUpper(strings.Replace(s, "_", "", -1))
	})

	a := []byte(preResult)
	a[0] = a[0] | ('a' - 'A')
	return string(a)
}

func formatString(format string, a ...interface{}) string {
	return fmt.Sprintf(format, a...)
}

func appendStringSlice(slice *[]string, elems ...string) {
	*slice = append(*slice, elems...)
}
