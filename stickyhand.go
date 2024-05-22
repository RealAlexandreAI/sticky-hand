package stickyhand

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/JohannesKaufmann/html-to-markdown"
	"github.com/RealAlexandreAI/json-repair"
	"github.com/bytedance/sonic"
	"github.com/chromedp/chromedp"
	"github.com/flosch/pongo2/v6"
	"github.com/go-shiori/go-readability"
	"github.com/samber/lo"
	"github.com/sashabaranov/go-openai"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

var optsInPlace = &sjson.Options{Optimistic: true, ReplaceInPlace: true}

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
		scraper := NewScraper(opts...)

		scraper.timeout = lo.Ternary(scraper.timeout <= 0, 60, scraper.timeout)
		timeout := time.Duration(scraper.timeout) * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		article, err := readability.FromURL(url, timeout)
		if err != nil {
			return fmt.Errorf("readability failed to parse %s, %v", url, err)
		}

		rst, _ = sjson.SetOptions(rst, "metadata.title", article.Title, optsInPlace)
		rst, _ = sjson.SetOptions(rst, "metadata.siteName", article.SiteName, optsInPlace)
		rst, _ = sjson.SetOptions(rst, "metadata.length", article.Length, optsInPlace)

		if scraper.text {
			rst, _ = sjson.SetOptions(rst, "text", article.TextContent, optsInPlace)
		}

		if scraper.html {
			rst, _ = sjson.SetOptions(rst, "html", article.Content, optsInPlace)
		}

		if scraper.markdown || (scraper.summary || scraper.translation != "" || scraper.mindmap) {

			converter := md.NewConverter("", true, nil)
			markdown, err := converter.ConvertString(article.Content)
			if err != nil {
				return fmt.Errorf("failed to html-to-markdown %s, %v", url, err)
			}

			rst, _ = sjson.SetOptions(rst, "markdown", markdown, optsInPlace)
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
			rst, _ = sjson.SetOptions(rst, "capture", base64Str, optsInPlace)
		}

		if scraper.llmClient != nil {

			if scraper.summary {

				prompt := summarizePrompt_ICIO + gjson.Get(rst, "markdown").String()

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
				summary = jsonrepair.MustRepairJSON(summary)

				rst, _ = sjson.SetRawOptions(rst, "summary", summary, optsInPlace)
			}

			if scraper.translation != "" {

				tpl, err := pongo2.FromString(translatePrompt)
				if err != nil {
					return fmt.Errorf("compile translate prompt template error: %v", err)
				}

				out, err := tpl.Execute(pongo2.Context{"targetLang": scraper.translation})
				if err != nil {
					return fmt.Errorf("render translate prompt template error: %v", err)
				}

				prompt := out + gjson.Get(rst, "markdown").String()

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
					return fmt.Errorf("chat completion error: %v", err)
				}

				translation := resp.Choices[0].Message.Content

				rst, _ = sjson.SetRawOptions(rst, "translation", translation, optsInPlace)
			}

			if scraper.mindmap {

				prompt := mermaidPrompt + gjson.Get(rst, "markdown").String()

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
					return fmt.Errorf("mermaid analyse error: %v", err)
				}

				mermaid := resp.Choices[0].Message.Content

				rst, _ = sjson.SetRawOptions(rst, "mermaid", mermaid, optsInPlace)
			}
		}

		return nil
	})

	return rst, lo.Ternary(ok, nil, fmt.Errorf("failed to scrape %s, %v", url, errS))
}

// ScrapeURLToStruct
//
//	@Description:
//	@param url
//	@param opts
//	@return StructuredOutput
//	@return error
func ScrapeURLToStruct(url string, opts ...Option) (*StructuredOutput, error) {

	rstStr, err := ScrapeURL(url, opts...)
	if err != nil {
		return nil, err
	}

	var rst StructuredOutput
	if err := sonic.UnmarshalString(rstStr, &rst); err != nil {
		return nil, err
	}

	return &rst, nil
}

// StructuredOutput
// @Description:
type StructuredOutput struct {
	Metadata    Metadata `json:"metadata,omitempty"`
	Text        string   `json:"text,omitempty"`
	HTML        string   `json:"html,omitempty"`
	Markdown    string   `json:"markdown,omitempty"`
	Capture     string   `json:"capture,omitempty"`
	Summary     Summary  `json:"summary,omitempty"`
	Translation string   `json:"translation,omitempty"`
	Mindmap     string   `json:"mindmap,omitempty"`
}

// Metadata
// @Description:
type Metadata struct {
	Title    string `json:"title"`
	SiteName string `json:"siteName"`
	Length   int    `json:"length"`
}

// Summary
// @Description:
type Summary struct {
	Title    string   `json:"title"`
	Keywords []string `json:"keywords"`
	Detailed string   `json:"detailed"`
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
	timeout     int
}

// outputConfig
// @Description:
type outputConfig struct {
	text        bool
	markdown    bool
	html        bool
	capture     bool
	summary     bool
	mindmap     bool
	translation string
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
func WithTranslation(language string) Option {
	return func(scraper *StickyHand) {
		scraper.markdown = true
		scraper.translation = language
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

// WithTimeout
//
//	@Description:
//	@return Option
func WithTimeout(timeout int) Option {
	return func(scraper *StickyHand) {
		scraper.timeout = timeout
	}
}

// WithMindMap
//
//	@Description:
//	@return Option
func WithMindMap() Option {
	return func(scraper *StickyHand) {
		scraper.mindmap = true
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
