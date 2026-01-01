package util

import "fmt"

func PrintError(err error, message ...string) {
	if len(message) == 0 {
		fmt.Println(fmt.Errorf("Encountered Error: %w", err))
	} else {
		fmt.Println(message[0], " : ", err)
	}
}
