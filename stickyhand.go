package stickyhand

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/JohannesKaufmann/html-to-markdown"
	"github.com/RealAlexandreAI/json-repair"
	"github.com/chromedp/chromedp"
	"github.com/go-shiori/go-readability"
	"github.com/samber/lo"
	"github.com/sashabaranov/go-openai"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const summarizePrompt =`
Employing the ICIO framework, the following Few-shot instruction is structured as follows:

**Instruction (I)**:
Analyze the provided text and generate a JSON output that includes a title, a detailed summary, and a list of identified keywords.

**Content (C)**:
The passage for analysis will be presented below.

**Intent (I)**:
The aim is to identify the primary themes, insights, and specific keywords within the text, and to produce a title that succinctly represents the text's main idea, along with a detailed summary that encapsulates its comprehensive essence, with emphasis on the identified keywords.

**Output (O)**:
The expected output is in JSON format, with the following structure:

{
  "title": "A concise title that reflects the text's central theme",
  "keywords": ["Keyword1", "Keyword2", "Keyword3"], // List of identified keywords
  "detailed": "An extensive summary that elaborates on the text's content in detail, incorporating the identified keywords"
}

Examples:
Below are examples illustrating the creation of a title, a detailed summary, and the identification of keywords from given text samples.

Example 1:
{
  "text": "Insert the content of example text 1 here...",
  "output": {
    "title": "Example Title 1",
    "keywords": ["Key1", "Concept1", "Theme1"],
    "detailed": "Example detailed summary for text 1, highlighting the presence and relevance of Key1, Concept1, and Theme1..."
  }
}

Example 2:
{
  "text": "Insert the content of example text 2 here...",
  "output": {
    "title": "Example Title 2",
    "keywords": ["Key2", "Idea2", "Topic2"],
    "detailed": "Example detailed summary for text 2, providing an in-depth overview and emphasizing Key2, Idea2, and Topic2..."
  }
}

The text to be processed is located below the line.

---

`

// ScrapeURL
//
//	@Description:
//	@param url
//	@param opts
//	@return string
//	@return error
func ScrapeURL(url string, opts ...Option) (string, error) {
	var rst string
	errS, ok := lo.TryWithErrorValue(func() error {
		timeout := 30 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		scraper := NewScraper(opts...)

		article, err := readability.FromURL(url, timeout)
		if err != nil {
			return fmt.Errorf("readability failed to parse %s, %v", url, err)
		}

		rst, _ = sjson.Set(rst, "metadata.title", article.Title)
		rst, _ = sjson.Set(rst, "metadata.siteName", article.SiteName)
		rst, _ = sjson.Set(rst, "metadata.length", article.Length)

		if scraper.text {
			rst, _ = sjson.Set(rst, "text", article.TextContent)
		}

		if scraper.html {
			rst, _ = sjson.Set(rst, "html", article.Content)
		}

		if scraper.markdown {

			converter := md.NewConverter("", true, nil)
			markdown, err := converter.ConvertString(article.Content)
			if err != nil {
				return fmt.Errorf("failed to html-to-markdown %s, %v", url, err)
			}

			rst, _ = sjson.Set(rst, "markdown", markdown)
		}

		if scraper.capture {

			cdCtx, cancelCD := chromedp.NewContext(ctx)
			defer cancelCD()
			_ = chromedp.Run(cdCtx)

			var buf []byte

			tCtx, cancelTF := context.WithTimeout(cdCtx, timeout)
			defer cancelTF()

			if err := chromedp.Run(tCtx,
				chromedp.Navigate(url),
				chromedp.CaptureScreenshot(&buf),
			); err != nil {
				return fmt.Errorf("failed to capture webpage: %v", err)
			}

			base64Str := base64.StdEncoding.EncodeToString(buf)
			rst, _ = sjson.Set(rst, "capture", base64Str)
		}

		if scraper.summary && scraper.llmClient != nil {

			prompt := summarizePrompt + gjson.Get(rst, "markdown").String()

			resp, err := scraper.llmClient.CreateChatCompletion(
				ctx,
				openai.ChatCompletionRequest{
					Model: openai.GPT3Dot5Turbo,
					Messages: []openai.ChatCompletionMessage{
						{
							Role:    openai.ChatMessageRoleUser,
							Content: prompt,
						},
					},
				},
			)
			if err != nil {
				return fmt.Errorf("ChatCompletion error: %v", err)
			}

			summary := resp.Choices[0].Message.Content
			summary = jsonrepair.RepairJSON(summary)

			rst, _ = sjson.SetRaw(rst, "summary", summary)
		}

		return nil
	})

	return rst, lo.Ternary(ok, nil, fmt.Errorf("failed to scrape %s, %v", url, errS))
}

// StickyHand
// @Description:
type StickyHand struct {
	scraperConfig
	outputConfig
	llmClient *openai.Client
}

// scraperConfig
// @Description:
type scraperConfig struct {
	llmEndpoint string
	llmAPIKey   string
}

// outputConfig
// @Description:
type outputConfig struct {
	text        bool
	markdown    bool
	html        bool
	capture     bool
	summary     bool
	translation bool
}

type Option func(*StickyHand)

// WithText
//
//	@Description:
//	@return Option
func WithText() Option {
	return func(scraper *StickyHand) {
		scraper.text = true
	}
}

// WithMarkdown
//
//	@Description:
//	@return Option
func WithMarkdown() Option {
	return func(scraper *StickyHand) {
		scraper.markdown = true
	}
}

// WithHTML
//
//	@Description:
//	@return Option
func WithHTML() Option {
	return func(scraper *StickyHand) {
		scraper.html = true
	}
}

// WithWebpageCapture
//
//	@Description:
//	@return Option
func WithWebpageCapture() Option {
	return func(scraper *StickyHand) {
		scraper.capture = true
	}
}

// WithSummary
//
//	@Description:
//	@return Option
func WithSummary() Option {
	return func(scraper *StickyHand) {
		scraper.summary = true
	}
}

// WithTranslation
//
//	@Description:
//	@return Option
func WithTranslation() Option {
	return func(scraper *StickyHand) {
		scraper.translation = true
	}
}

// WithLLMProvider
//
//	@Description:
//	@param endpoint
//	@param apiKey
//	@return Option
func WithLLMProvider(endpoint string, apiKey string) Option {
	return func(scraper *StickyHand) {
		scraper.llmEndpoint = endpoint
		scraper.llmAPIKey = apiKey

		config := openai.DefaultConfig(apiKey)

		if endpoint != "" {
			config.BaseURL = endpoint
		}

		c := openai.NewClientWithConfig(config)
		scraper.llmClient = c
	}
}

// NewScraper
//
//	@Description:
//	@param opts
//	@return *StickyHand
func NewScraper(opts ...Option) *StickyHand {
	scraper := &StickyHand{}

	for _, opt := range opts {
		opt(scraper)
	}

	return scraper
}
