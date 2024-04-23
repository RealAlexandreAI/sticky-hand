package main

import (
	"flag"
	"fmt"

	stickyhand "github.com/RealAlexandreAI/sticky-hand"
)

const AppVersion = "0.0.2"

var (
	versionFlag bool
	helpFlag    bool
	text        bool
	markdown    bool
	html        bool
	capture     bool
	summary     bool
	translation bool

	llmEndpoint string
	llmAPIKey   string
)

// init
//
//	@Description:
func init() {
	flag.BoolVar(&versionFlag, "v", false, "Print version details")
	flag.BoolVar(&helpFlag, "h", false, "Print help")
	flag.BoolVar(&text, "text", false, "With text output")
	flag.BoolVar(&markdown, "markdown", false, "With markdown output")
	flag.BoolVar(&html, "html", false, "With html output")
	flag.BoolVar(&capture, "capture", false, "With capture output")
	flag.BoolVar(&summary, "summary", false, "With summary output")
	// flag.BoolVar(&translation, "translation", false, "With text output")

	flag.StringVar(&llmEndpoint, "with-llm-endpoint", "", "With llm endpoint eg: https://<llm-provider>/v1")
	flag.StringVar(&llmAPIKey, "with-llm-apikey", "", "With llm apikey eg: sk-xxxxx")
}

// printDefaults
//
//	@Description:
func printDefaults() {
	fmt.Println("Usage: stickyhand <options> <url> ")
	fmt.Println("Options:")
	flag.VisitAll(func(flag *flag.Flag) {
		fmt.Println("\t-"+flag.Name, "\t", flag.Usage, "(Default "+flag.DefValue+")")
	})
}

// main
//
//	@Description:
func main() {
	flag.Parse()

	if versionFlag {
		fmt.Println("Version:", AppVersion)
		return
	} else if helpFlag {
		printDefaults()
		return
	}

	targetUrl := flag.Arg(0)
	if targetUrl == "" {
		targetUrl = "https://github.com/RealAlexandreAI/sticky-hand"
	}

	var options []stickyhand.Option

	if text {
		options = append(options, stickyhand.WithText())
	}

	if html {
		options = append(options, stickyhand.WithHTML())
	}

	if markdown {
		options = append(options, stickyhand.WithMarkdown())
	}

	if capture {
		options = append(options, stickyhand.WithWebpageCapture())
	}

	if summary {
		options = append(options, stickyhand.WithSummary())
	}

	if llmAPIKey != "" {
		options = append(options, stickyhand.WithLLMProvider(llmEndpoint, llmAPIKey))
	}

	rst, err := stickyhand.ScrapeURL(targetUrl, options...)
	if err != nil {
		fmt.Println("stickyhand scrape error:", err)
		return
	}

	fmt.Println(rst)
}
