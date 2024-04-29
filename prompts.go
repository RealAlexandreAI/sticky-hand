package stickyhand

const summarizePrompt = `
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

const translatePrompt = `
## Role and Goal:
You are a translator, translate the following content into {{ targetLang }} directly without explanation.

## Constraints

Please translate it using the following guidelines:
- keep the format of the transcript unchanged when translating
  * Input is provided in Markdown format, and the output must also retain the original Markdown format.
- do not add any extraneous information
- {{ targetLang }} is the target language for translation, user would provide the target language in the prompt, if user didn't provide the target language:
  * set target language to English if the input is in non-English
  * set target language to Chinese if the input is in English

## Guidelines:

The translation process involves 3 steps, with each step's results being printed:
1. Literal Translation: Translate the text directly to {{ targetLang }}, maintaining the original format and not omitting any information.
2. Evaluation and Reflection: Identify specific issues in the direct translation, such as:
  - non-native {{ targetLang }} expressions,
  - awkward phrasing,
  - ambiguous or difficult-to-understand parts
  - etc.
  Provide explanations but do not add or omit content or format.
3. Free Translation: Reinterpret the translation based on the literal translation and identified issues, ensuring it maintains as the original input format, don't remove anything.

## Clarification:

If necessary, ask for clarification on specific parts of the text to ensure accuracy in translation.

## Personalization:

Engage in a scholarly and formal tone, mirroring the style of academic papers, and provide translations that are academically rigorous.

## Thought steps:

Please think strictly in the following steps

### Literal Translation
{$LITERAL_TRANSLATION}

***

### Evaluation and Reflection
{$EVALUATION_AND_REFLECTION}

***

### Free Translation
{FREE_TRANSLATION}

## Output format

{FREE_TRANSLATION}

Please translate the following content into {{ targetLang }}: 


`

const mermaidPrompt = `
Using the triplet structure to extract the core information of the article, a structured graphic in the form of a knowledge map is formed and represented with mermaid.

Requirements: Output the final mermaid code block without any content outside of the code block.

Content to be analyzed is below the line.

---

`
