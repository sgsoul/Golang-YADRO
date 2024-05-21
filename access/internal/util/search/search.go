package search

import (
	"sort"
	"strings"

	"github.com/sgsoul/internal/core/database"
	"github.com/sgsoul/internal/util/words"
)

func FindRelevantComics(db *database.DB, keywords []string) ([]database.Comic, error) {
	var relevantComics []database.Comic

	comics, err := db.GetAllComics()
	if err != nil {
		return nil, err
	}

	for _, comic := range comics {
		relevanceScore := countMatchingKeywords(comic, keywords)

		if relevanceScore > 0 {
			relevantComics = append(relevantComics, comic)
		}
	}

	sort.Slice(relevantComics, func(i, j int) bool {
		return countMatchingKeywords(relevantComics[i], keywords) > countMatchingKeywords(relevantComics[j], keywords)
	})

	return relevantComics, nil
}

func countMatchingKeywords(comic database.Comic, keywords []string) int {
	var count int
	for _, keyword := range keywords {
		normalizedKeywords := words.NormalizeWords(keyword)
		for _, normalizedKeyword := range normalizedKeywords {
			comicKeywords := strings.Split(comic.Keywords, ",")
			for _, comicKeyword := range comicKeywords {
				if strings.EqualFold(comicKeyword, normalizedKeyword) {
					count++
					break
				}
			}
		}
	}
	return count
}
