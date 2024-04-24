package words

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/kljensen/snowball"
)

func NormalizeWords(texts ...string) []string {
	var words []string
	seen := make(map[string]bool)
	for _, text := range texts {
		// разбивка на слова
		splitWords := strings.FieldsFunc(text, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsNumber(r)
		})
		// нормализация и аппенд
		for _, word := range splitWords {
			normalized := unstyle(normalize(word))
			if !seen[normalized] && !Stopwords[normalized] && !IsStopWord(word) {
				seen[normalized] = true
				if normalized != "" && len(normalized) > 2 {
					words = append(words, normalized)
				}
			}
		}
	}
	return words
}

func normalize(word string) string {
	// удаление лишних символов
	f := func(r rune) bool {
		return !unicode.IsLetter(r) && r != '\''
	}
	fields := strings.FieldsFunc(word, f)
	cleaned := strings.Join(fields, "")

	// стемминг нормализованного слова
	stemmed, err := snowball.Stem(cleanWord(cleaned), "english", true)
	if err != nil {
		return cleaned
	}
	return stemmed
}

func cleanWord(word string) string {
	// регулярное выражение для удаления глагольных окончаний
	re := regexp.MustCompile(`(|n)'(ll|ve|re|s|d|m|t)\b`)
	cleanedWord := re.ReplaceAllString(word, "")
	return cleanedWord
}

func unstyle(str string) string {
	var builder strings.Builder
	for _, r := range str {
		// Проверяем, есть ли символ в маппинге
		if normal, ok := StyledToNormal[r]; ok {
			builder.WriteRune(normal)
		} else {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}
