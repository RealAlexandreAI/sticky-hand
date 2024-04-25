package stickyhand

import (
	"encoding/base64"
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
