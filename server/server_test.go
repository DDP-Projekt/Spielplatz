package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTruncateSource(t *testing.T) {
	assert := assert.New(t)

	assert.Equal("test", truncSourceString("test", 100))
	assert.Equal("tes...", truncSourceString("test", 3))
	assert.Equal("...tes...", truncSourceString("Binde\ntest", 3))
	assert.Equal("...test...", truncSourceString("Binde\ntest", 4))
	assert.Equal("...test...",
		truncSourceString("Binde\nBinde\nBinde\nBinde\nBinde\nBinde\nBinde\nBinde\nBinde\nBinde\nBinde\nBinde\nBinde\nBinde\nBinde\ntest", 4))

	// test newlineIndex == -1 break
	assert.Equal("...Bind...",
		truncSourceString("Binde \"Duden/Ausgabe\" ein.Binde \"Duden/Eingabe\" ein.test", 4))
}
