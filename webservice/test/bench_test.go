package main

import (
	"os"
	"testing"

	"github.com/sgsoul/pkg/database"
	sr "github.com/sgsoul/pkg/search"
	"github.com/sgsoul/pkg/words"
)

var testString = "hello from the other side"

// бенчмарк для тестирования поиска по индексу
func BenchmarkIndexSearch(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
        index, _ := os.ReadFile("../index.json")
		normalizedKeywords := words.NormalizeWords(testString)

		sr.RelevantComic(sr.IndexSearch(index, normalizedKeywords), "../database.json")
    }
}

// бенчмарк для тестирования поиска по бд
func BenchmarkDatabaseSearch(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		db, _, _ := database.LoadComicsFromFile("../database.json")
		normalizedKeywords := words.NormalizeWords(testString)

		sr.FindRelevantComics(db, normalizedKeywords)
	}
}
