package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: ./link _file_")
	}

	filename := os.Args[1]

	if !strings.Contains(filename, "/") {
		filename = "./" + filename
	}

	content, err := os.ReadFile(filename)

	if err != nil {
		log.Fatal(err)
	}

	result := iterateText(string(content))

	if len(result) != 0 {
		file, err := os.OpenFile(filename, os.O_RDWR|os.O_TRUNC, 0600)
		defer file.Close()

		if err != nil {
			log.Fatal(err)
		}

		_, err = file.Write([]byte(result))

		if err != nil {
			log.Fatal(err)
		}
	}

}

func createLink(videoID string) string {
	return "[![](https://img.youtube.com/vi/" + videoID + "/maxresdefault.jpg)](https://www.youtube.com/watch?v=" + videoID + ")"
}

func iterateText(text string) string {
	result := ""
	lines := strings.Fields(text)

	for _, line := range lines {
		if strings.Contains(line, "[!") {
			result += line + "\n"
			continue
		}

		if strings.Contains(line, "!") {
			fmt.Println("TEST_DOBERMAN")
			lineParts := strings.Split(line, "v=")
			videoID, _ := strings.CutSuffix(lineParts[1], ")")
			result += createLink(videoID) + "\n"
		}
	}

	return result
}
