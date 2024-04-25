package stickyhand

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestScrapeURL(t *testing.T) {
	rst, err := ScrapeURL("https://github.com/RealAlexandreAI/sticky-hand",
		WithHTML(), WithMarkdown(),
		WithSummary(), WithText())

	assert.NoError(t, err)
	assert.Contains(t, gjson.Get(rst, "metadata.title").String(), "hand")
	assert.True(t, gjson.Get(rst, "metadata.siteName").String() != "")
	assert.Greater(t, gjson.Get(rst, "metadata.length").Int(), int64(0))
	assert.True(t, gjson.Get(rst, "text").String() != "")
	assert.True(t, gjson.Get(rst, "html").String() != "")
	assert.True(t, gjson.Get(rst, "markdown").String() != "")
	_, err = base64.StdEncoding.DecodeString(gjson.Get(rst, "capture").String())
	assert.NoError(t, err)
}

func TestScrapeURL2(t *testing.T) {

	rst, err := ScrapeURL("https://en.wikipedia.org/wiki/Scrape",
		WithTranslation("Chinese"),
		WithLLMProvider("https://burn.hair/v1", "sk-LASi6o5kjiKNZnGm965748C91b384a5690Fc635a784c570e"))
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(rst)

}
