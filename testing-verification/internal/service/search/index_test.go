package search

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"testing"

	"github.com/sgsoul/internal/core"
	"github.com/stretchr/testify/assert"
)

func TestIndexSearch(t *testing.T) {
	indexData := []byte(`{"keyword1":[1,2,3],"keyword2":[2,3,4]}`)

	testCases := []struct {
		name               string
		file               []byte
		normalizedKeywords []string
		expected           map[int]int
	}{
		{
			name:               "Single keyword",
			file:               indexData,
			normalizedKeywords: []string{"keyword1"},
			expected:           map[int]int{1: 1, 2: 1, 3: 1},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IndexSearch(tc.file, tc.normalizedKeywords)
			assert.Equal(t, tc.expected, result)
		})
	}
}

type MockStorage struct{}

func (m *MockStorage) GetAllComics() ([]core.Comic, error) {
	return []core.Comic{}, nil
}

func (m *MockStorage) GetComicByID(id int) (core.Comic, error) {
	return core.Comic{}, nil
}

func TestBuildIndex(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	s := &search{storage: &MockStorage{}}

	err = s.buildIndex(tmpDir + "/test_index.json")
	assert.NoError(t, err)

	indexFile, err := os.ReadFile(tmpDir + "/test_index.json")
	assert.NoError(t, err)

	var index map[string][]int
	err = json.Unmarshal(indexFile, &index)
	assert.NoError(t, err)
}

type mockStorage struct{}

func (m *mockStorage) GetAllComics() ([]core.Comic, error) {
	return []core.Comic{
		{ID: 1, Keywords: "apple,pie"},
		{ID: 2, Keywords: "pie"},
		{ID: 3, Keywords: "apple"},
		{ID: 4, Keywords: "root"},
		{ID: 5, Keywords: "pie"},
	}, nil
}

func (m *mockStorage) GetComicByID(id int) (core.Comic, error) {
	comics, err := m.GetAllComics()
	if err != nil {
		return core.Comic{}, err
	}
	for _, comic := range comics {
		if comic.ID == id {
			return comic, nil
		}
	}
	return core.Comic{}, fmt.Errorf("comic with ID %d not found", id)
}

var yourMockStorageImplementation = &mockStorage{}

func TestNewIndex(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "index")
	if err != nil {
		t.Fatalf("error creating temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name()) 

	s := &search{storage: yourMockStorageImplementation}

	testIndex := []byte(`{
        "apple": [
            1,
            3
        ],
        "pie": [
            1,
            2,
            5
        ],
        "root": [
            4
        ]
    }`)
	if _, err := tmpfile.Write(testIndex); err != nil {
		t.Fatalf("error writing test index to file: %v", err)
	}

	indexBytes, err := s.newIndex(tmpfile.Name())
	assert.NoError(t, err)

	var expectedIndex, actualIndex map[string][]int
	if err := json.Unmarshal(testIndex, &expectedIndex); err != nil {
		t.Fatalf("error decoding expected index: %v", err)
	}
	if err := json.Unmarshal(indexBytes, &actualIndex); err != nil {
		t.Fatalf("error decoding actual index: %v", err)
	}

	expectedKeys := make([]string, 0, len(expectedIndex))
	for key := range expectedIndex {
		expectedKeys = append(expectedKeys, key)
	}
	sort.Strings(expectedKeys)

	actualKeys := make([]string, 0, len(actualIndex))
	for key := range actualIndex {
		actualKeys = append(actualKeys, key)
	}
	sort.Strings(actualKeys)

	for _, key := range expectedKeys {
		if !reflect.DeepEqual(expectedIndex[key], actualIndex[key]) {
			t.Errorf("loaded index does not match expected index, expected: %v, got: %v", expectedIndex, actualIndex)
			break
		}
	}
}

func TestRelevantComic(t *testing.T) {
	s := &search{storage: yourMockStorageImplementation}

	relevantComics := map[int]int{
		1: 10,
		2: 5,
		3: 3,
	}

	expectedSortedComics := []core.Comic{
		{ID: 1, Keywords: "apple,pie"},
		{ID: 2, Keywords: "pie"},
		{ID: 3, Keywords: "apple"},
	}

	sortedComics, err := s.RelevantComic(relevantComics)
	if err != nil {
		t.Fatalf("error getting relevant comics: %v", err)
	}

	sort.Slice(expectedSortedComics, func(i, j int) bool {
		return expectedSortedComics[i].ID < expectedSortedComics[j].ID
	})
	sort.Slice(sortedComics, func(i, j int) bool {
		return sortedComics[i].ID < sortedComics[j].ID
	})

	if !reflect.DeepEqual(sortedComics, expectedSortedComics) {
		t.Errorf("sorted comics do not match expected sorted comics, expected: %v, got: %v", expectedSortedComics, sortedComics)
	}
}