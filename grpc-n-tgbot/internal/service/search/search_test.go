package search

import (
	"os"
	"reflect"
	"testing"

	"github.com/sgsoul/internal/core"
	mocks "github.com/sgsoul/internal/service/search/mocks"
	"github.com/stretchr/testify/assert"

	"github.com/golang/mock/gomock"
)

func TestRelevantURLS(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)

	mockStorage.EXPECT().GetAllComics().Return([]core.Comic{
		{ID: 1, URL: "comic1URL", Keywords: "apple,pie"},
		{ID: 2, URL: "comic2URL", Keywords: "apple,dock"},
	}, nil)

	mockStorage.EXPECT().GetComicByID(1).Return(core.Comic{ID: 1, URL: "comic1URL"}, nil)

	s := &search{storage: mockStorage}

	urls, comics := s.RelevantURLS("apple pie", "index.json")
	defer os.Remove("index.json")

	expectedURLs := []string{"comic1URL"}
	if !reflect.DeepEqual(urls, expectedURLs) {
		t.Errorf("got %v, expected %v", urls, expectedURLs)
	}

	expectedComics := []core.Comic{
		{ID: 1, URL: "comic1URL"},
	}
	if !reflect.DeepEqual(comics, expectedComics) {
		t.Errorf("got %v, expected %v", comics, expectedComics)
	}
}

func TestFindRelevantComics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)

	mockStorage.EXPECT().GetAllComics().Return([]core.Comic{
		{ID: 1, Keywords: "apple, doctor"},
		{ID: 2, Keywords: "apple, pie"},
		{ID: 3, Keywords: "brush"},
	}, nil)

	s := &search{storage: mockStorage}

	relevantComics, err := s.FindRelevantComics([]string{"apple,pie"})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expectedIDs := []int{2, 1} // ожидаемый порядок ID комиксов
	for i, comic := range relevantComics {
		if comic.ID != expectedIDs[i] {
			t.Errorf("got comic with ID %d at index %d, expected ID %d", comic.ID, i, expectedIDs[i])
		}
	}
}

func TestCountMatchingKeywords(t *testing.T) {
	testComic := core.Comic{Keywords: "superhero, action, adventure"}
	keywords := []string{"Superhero", "Sci-Fi"}
	result := countMatchingKeywords(testComic, keywords)

	expected := 1

	if result != expected {
		t.Errorf("Expected %d matching keyword(s), but got %d", expected, result)
	}
}

func SetupSearch(t *testing.T) *search{
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)

	s := NewSearch(mockStorage)
	return s
}

func TestNew(t *testing.T) {
	s := SetupSearch(t)
	assert.NotNil(t, s, "should've been created..")
}