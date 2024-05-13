package search

import (
	"sort"
	"strings"

	"github.com/sgsoul/internal/core/database"
	"github.com/sgsoul/internal/util/words"
)

// FindRelevantComics finds the most relevant comics for a query
func FindRelevantComics(db *database.DB, keywords []string) ([]database.Comic, error) {
	var relevantComics []database.Comic

	// Retrieve all comics from the database
	comics, err := db.GetAllComics()
	if err != nil {
		return nil, err
	}

	for _, comic := range comics {
		// Count the number of matching keywords
		relevanceScore := countMatchingKeywords(comic, keywords)

		// Add the comic to the list if it has at least one match with the query
		if relevanceScore > 0 {
			relevantComics = append(relevantComics, comic)
		}
	}

	// Sort comics in descending order of relevance
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
