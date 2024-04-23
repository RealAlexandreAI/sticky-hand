package stickyhand

import (
	"context"
	_ "embed"
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

//go:embed prompts/summarize.prompt
var summarizePrompt []byte

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
			return fmt.Errorf("readability failed to parse %s, %v\n", url, err)
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
				return fmt.Errorf("failed to html-to-markdown %s, %v\n", url, err)
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

		if (scraper.summary || scraper.keywords) && scraper.llmClient != nil {

			prompt := string(summarizePrompt) + gjson.Get(rst, "markdown").String()

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
				return fmt.Errorf("ChatCompletion error: %v\n", err)
			}

			summary := resp.Choices[0].Message.Content
			summary = jsonrepair.RepairJSON(summary)

			rst, _ = sjson.SetRaw(rst, "summary", summary)
		}

		return nil
	})

	return rst, lo.Ternary(ok, nil, fmt.Errorf("failed to scrape %s, %v\n", url, errS))
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
	keywords    bool
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

// WithKeywords
//
//	@Description:
//	@return Option
func WithKeywords() Option {
	return func(scraper *StickyHand) {
		scraper.keywords = true
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
		config.BaseURL = endpoint

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
