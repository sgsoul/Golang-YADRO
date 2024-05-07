package search

import (
	"sort"
	"strings"

	"github.com/sgsoul/internal/core/database"
	"github.com/sgsoul/internal/util/words"
)

// findRelevantComics находит наиболее релевантные комиксы к запросу
func FindRelevantComics(db map[string]database.Comic, keywords []string) []database.Comic {
	var relevantComics []database.Comic

	for _, comic := range db {
		// Подсчет количества совпадающих ключевых слов
		var relevanceScore int
		for _, comicKeyword := range comic.Keywords {
			for _, keyword := range keywords {
				normalizedKeywords := words.NormalizeWords(keyword)
				for _, normalizedKeyword := range normalizedKeywords {
					if strings.EqualFold(comicKeyword, normalizedKeyword) {
						relevanceScore++
						break
					}
				}
			}
		}
		// Добавление комикса в список, если он имеет хотя бы одно совпадение с запросом
		if relevanceScore > 0 {
			relevantComics = append(relevantComics, comic)
		}
	}

	// Сортировка комиксов по убыванию релевантности
	// (по количеству совпадающих ключевых слов)
	sort.Slice(relevantComics, func(i, j int) bool {
		return countMatchingKeywords(relevantComics[i], keywords) > countMatchingKeywords(relevantComics[j], keywords)
	})

	return relevantComics
}

// countMatchingKeywords подсчитывает количество совпадающих ключевых слов в комиксе
func countMatchingKeywords(comic database.Comic, keywords []string) int {
	var count int
	for _, keyword := range keywords {
		normalizedKeywords := words.NormalizeWords(keyword)
		for _, normalizedKeyword := range normalizedKeywords {
			for _, comicKeyword := range comic.Keywords {
				if strings.EqualFold(comicKeyword, normalizedKeyword) {
					count++
					break
				}
			}
		}
	}
	return count
}
