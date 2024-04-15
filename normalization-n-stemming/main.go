package main

import (
	"flag"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/kljensen/snowball"
)

func main() {
	var input string
	flag.StringVar(&input, "s", "", "Input string")
	flag.Parse()

	if strings.TrimSpace(input) == "" {
		fmt.Println("нужно подать непустую строку")
		return
	}

	words := strings.Fields(input)

	seen := make(map[string]bool)
	allWordsDeleted := true
	// нормализация и вывод уникальных слов
	for _, word := range words {
		normalized := Normalize(word)
		if !seen[normalized] && !Stopwords[normalized] && !IsStopWord(word) {
			fmt.Print(normalized, " ")
			seen[normalized] = true
			if normalized != "" {
				allWordsDeleted = false
			}
		}
	}

	// обработка пустого вывода
	if allWordsDeleted {
		fmt.Println("невозможно выполнить поиск по этой строке:(")
	}
}

func Normalize(word string) string {
	// удаление лишних символов
	f := func(r rune) bool {
		return !unicode.IsLetter(r) && r != '\''
	}
	fields := strings.FieldsFunc(word, f)
	cleaned := strings.Join(fields, "")

	// стемминг нормализованного слова
	stemmed, err := snowball.Stem(CleanWord(cleaned), "english", true)
	if err != nil {
		return cleaned
	}
	return stemmed
}

func CleanWord(word string) string {
	// регулярное выражение для удаления глагольных окончаний
	re := regexp.MustCompile(`(|n)'(ll|ve|re|s|d|m|t)\b`)
	cleanedWord := re.ReplaceAllString(word, "")
	return cleanedWord
}
