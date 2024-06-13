package words

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// var StopwordsMock = map[string]bool{
// 	"the": true,
// 	"and": true,
// 	"to":  true,
// 	"of":  true,
// }

// var StyledToNormalMock = map[rune]rune{
// 	'𝓪': 'a',
// 	'𝓫': 'b',
// 	'𝓬': 'c',
// }

// func IsStopWordMock(word string) bool {
// 	switch word {
// 	case "them", "oh", "own":
// 		return true
// 	}
// 	return false
// }

func TestNormalize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"running", "run"},
		{"cleaning", "clean"},
		{"happily", "happili"},
		{"", ""},
	}

	for _, test := range tests {
		result := normalize(test.input)
		msg := fmt.Sprintf("normalize(%q) = %q; expected %q", test.input, result, test.expected)
		assert.Equal(t, test.expected, result, "they should be equal", msg)
	}
}

func TestCleanWord(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"you're", "you"},
		{"can't", "ca"},
		{"", ""},
	}

	for _, test := range tests {
		result := cleanWord(test.input)
		msg := fmt.Sprintf("cleanWord(%q) = %q; expected %q", test.input, result, test.expected)
		assert.Equal(t, test.expected, result, "they should be equal", msg)
		}
}

func TestUnstyle(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"𝓪𝓫𝓬", "abc"},
		{"ma𝓷ana", "manana"},
		{"", ""},
	}

	for _, test := range tests {
		result := unstyle(test.input)
		msg := fmt.Sprintf("unstyle(%q) = %q; expected %q", test.input, result, test.expected)
		assert.Equal(t, test.expected, result, "they should be equal", msg)
	}
}

func TestNormalizeWords(t *testing.T) {
	tests := []struct {
		input    []string
		expected []string
	}{
		{[]string{"This is a test, and only a test."}, []string{"test"}},
		{[]string{"Running, jumping, and playing!"}, []string{"run", "jump", "play"}},
		{[]string{"i'll follow you as long as you are following me"}, []string{"follow", "long"}},
		{[]string{""}, []string(nil)},

	}

	for _, test := range tests {
		result := NormalizeWords(test.input...)
		msg := fmt.Sprintf("NormalizeWords(%v) = %v; expected %v", test.input, result, test.expected)
		assert.Equal(t, test.expected, result, "they should be equal", msg)
	}
}