package stickyhand

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestScrapeURL(t *testing.T) {
	rst, err := ScrapeURL("https://en.wikipedia.org/wiki/Scrape",
		WithHTML(), WithMarkdown(),
		WithSummary(), WithText(), WithTranslation())

	assert.NoError(t, err)
	assert.Equal(t, gjson.Get(rst, "metadata.title").String(), "Scrape")
	assert.Contains(t, gjson.Get(rst, "metadata.siteName").String(), "Wikimedia")

	fmt.Println(gjson.Get(rst, "metadata.length").Int())
	assert.Greater(t, gjson.Get(rst, "metadata.length").Int(), int64(0))
	assert.True(t, gjson.Get(rst, "text").String() != "")
	assert.True(t, gjson.Get(rst, "html").String() != "")
	assert.True(t, gjson.Get(rst, "markdown").String() != "")
	assert.True(t, gjson.Get(rst, "capture").String() != "")
	_, err = base64.StdEncoding.DecodeString(gjson.Get(rst, "capture").String())
	assert.NoError(t, err)
}
