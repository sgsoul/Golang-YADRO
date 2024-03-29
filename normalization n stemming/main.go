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
		normalized := normalize(word)
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

func normalize(word string) string {
	// удаление лишних символов
	var cleanedWord strings.Builder
	for _, r := range word {
		if unicode.IsLetter(r) || r == '\'' { // оставляем только буквы и апострофы
			cleanedWord.WriteRune(r)
		}
	}
	cleaned := cleanedWord.String()

	// стемминг нормализованного слова
	stemmed, err := snowball.Stem(cleanWord(cleaned), "english", true)
	if err != nil {
		return cleaned
	}
	return stemmed
}

func cleanWord(word string) string {
	// регулярное выражение для удаления глагольных окончаний
	re := regexp.MustCompile(`'(ll|ve|re|s|d|m|t)\b`)
	cleanedWord := re.ReplaceAllString(word, "")
	return cleanedWord
}
