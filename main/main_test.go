package main

import (
	"bytes"
	"testing"
)

type testCase struct {
	name   string
	input  string
	output string
}

func TestConvertMarkdownFileToBlogHTML(t *testing.T) {
	testCases := []testCase{
		{
			name:   "an empty string should be returned as-is",
			input:  "",
			output: "",
		},
		{
			name:   "a single line with no markdown character should be returned as-is",
			input:  "This is a plain text file.",
			output: "This is a plain text file.",
		},
		{
			name:   "a line starting with # should be in <h1> tags",
			input:  "# This is an h1 header",
			output: "<h1> This is an h1 header</h1>",
		},
		{
			name:   "a line starting with ## should be in <h2> tags",
			input:  "## This is an h2 header",
			output: "<h2> This is an h2 header</h2>",
		},
		{
			name:   "a line starting with ### should be in <h3> tags",
			input:  "### This is an h3 header",
			output: "<h3> This is an h3 header</h3>",
		},
		{
			name:   "'#' tags in a header should be returned as-is",
			input:  "### This is an ### h3 ### header",
			output: "<h3> This is an ### h3 ### header</h3>",
		},
		{
			name:   "text across multiple lines with no markdown should be returned as-is",
			input:  "This\nis\nsome simple text\nwhich has been spread out\nacross multiple lines.",
			output: "This\nis\nsome simple text\nwhich has been spread out\nacross multiple lines.",
		},
		{
			name:   "paragraph tags should be added if there is a blank line before a line of plain text",
			input:  "This is a first line.\n\nParagraph one.",
			output: "This is a first line.\n<p>\nParagraph one.\n</p>",
		},
		{
			name:   "paragraph tags should be added if there is a blank line before a line of plain text for multiple paragraphs",
			input:  "This is a first line.\n\nParagraph one.\n\nParagraph two.",
			output: "This is a first line.\n<p>\nParagraph one.\n</p>\n<p>\nParagraph two.\n</p>",
		},
		{
			name:   "italics tags should be added if a string is surrounded by '*'",
			input:  "*italic text*",
			output: "<i>italic text</i>",
		},
		{
			name:   "bold tags should be added if a string is surrounded by '**'",
			input:  "**bold text**",
			output: "<b>bold text</b>",
		},
		{
			name:   "italics and bold tags should be added if a string is surrounded by '***'",
			input:  "***bold text***",
			output: "<i><b>bold text</b></i>",
		},
		{
			name:   "reserved characters should be substituted with html entities",
			input:  "This is a file ' which is <filled> with - HTML \"entities\" of interest.",
			output: "This is a file &apos; which is &lt;filled&gt; with &ndash; HTML &quot;entities&quot; of interest.",
		},
		{
			name:   "an empty inline code block should be skipped",
			input:  "``",
			output: "",
		},
		{
			name:   "plain text between a code block should be kept as-is",
			input:  "`This is a simple inline code block`",
			output: "<code>This is a simple inline code block</code>",
		},
		{
			name:   "code tags should be positioned correctly around an inline code block within another sentence",
			input:  "This is some text surrounding `and inline code block`.",
			output: "This is some text surrounding <code>and inline code block</code>.",
		},
		{
			name:   "reserved characters within a code block should be replaced with HTML entities",
			input:  "This file contains `a code block` with `a number of ' <> - \" ` html entities in it.",
			output: "This file contains <code>a code block</code> with <code>a number of &apos; &lt;&gt; &ndash; &quot; </code> html entities in it.",
		},
		{
			name:   "multi-line plain text within a code block should be kept as-is",
			input:  "```programming_language\nThis is a multiline code block.\nLine one,\nLine two,\nLine three.\n```",
			output: "<pre><code>\nThis is a multiline code block.\nLine one,\nLine two,\nLine three.\n</code></pre>",
		},
		{
			name:   "paragraphs of plain text within a code block should be kept as-is (without paragraph tags)",
			input:  "```some programming language\nThis is a line.\n\nHere is another line. It should not be in paragraph tags.\n\nA final line.\n```",
			output: "<pre><code>\nThis is a line.\n\nHere is another line. It should not be in paragraph tags.\n\nA final line.\n</code></pre>",
		},
		{
			name:   "a paragraph of plain text with an inline code block in it should wrap the <code> tags around it properly",
			input:  "This is a line.\n\nParagraph `with a code block` in it.",
			output: "This is a line.\n<p>\nParagraph <code>with a code block</code> in it.\n</p>",
		},
		{
			name:   "a paragraph of plain text with an inline code block in it should wrap the <code> tags around it properly",
			input:  "This is a line.\n\nHere is a multi-line code block:\n\n```code\nLine one,\n\nLine two,\n\nline three.\n```\n\nThat's the end of the code block.",
			output: "This is a line.\n<p>\nHere is a multi&ndash;line code block:\n</p>\n<p>\n<pre><code>\nLine one,\n\nLine two,\n\nline three.\n</code></pre>\n</p>\n<p>\nThat&apos;s the end of the code block.\n</p>",
		},
		{
			name:   "a multi-line code block with a directory structure within it should be rendered correctly",
			input:  "```\n- dashboard\n| - frontend\n| - backend\n```",
			output: "<pre><code>\n&ndash; dashboard\n| &ndash; frontend\n| &ndash; backend\n</code></pre>",
		},
		{
			name:   "a multi-line code block with a directory structure within it should be rendered correctly",
			input:  "```\n- dashboard\n| - frontend\n| - backend\n```",
			output: "<pre><code>\n&ndash; dashboard\n| &ndash; frontend\n| &ndash; backend\n</code></pre>",
		},
		{
			name:   "asterisks within a code block should be left as-is",
			input:  "```js\nimport * as echarts from 'echarts';\n```",
			output: "<pre><code>\nimport * as echarts from &apos;echarts&apos;;\n</code></pre>",
		},
		{
			name:   "inline footnotes should be replaced with <a id=\"footnote-anchor-n\" href=\"#footnote-n\">[n]</a>",
			input:  "Here is a footnote.[^1]",
			output: "Here is a footnote.<a id=\"footnote-anchor-1\" href=\"#footnote-1\">[1]</a>",
		},
		{
			name:   "successive inline footnotes should be replaced with  <a id=\"footnote-anchor-n\" href=\"#footnote-n\">[n]</a> and be numbered correctly",
			input:  "Here is a footnote[^1] and another footnote.[^2]",
			output: "Here is a footnote<a id=\"footnote-anchor-1\" href=\"#footnote-1\">[1]</a> and another footnote.<a id=\"footnote-anchor-2\" href=\"#footnote-2\">[2]</a>",
		},
		{
			name:   "out-of-order footnote numbers should be updated to be in increasing order",
			input:  "Here is a footnote[^2] and another footnote.[^1]",
			output: "Here is a footnote<a id=\"footnote-anchor-1\" href=\"#footnote-1\">[1]</a> and another footnote.<a id=\"footnote-anchor-2\" href=\"#footnote-2\">[2]</a>",
		},
		{
			name:   "a footnote in a paragraph should have paragraph and anchor tags added correctly",
			input:  "# This is a heading\n\nHere is a footnote.[^1]",
			output: "<h1> This is a heading</h1>\n<p>\nHere is a footnote.<a id=\"footnote-anchor-1\" href=\"#footnote-1\">[1]</a>\n</p>",
		},
		{
			name:   "a footnote in a paragraph should have paragraph and anchor tags added correctly, and successive footnotes should be numbered in increasing order",
			input:  "# This is a heading\n\nHere is a footnote.[^2] Here's another.[^1]",
			output: "<h1> This is a heading</h1>\n<p>\nHere is a footnote.<a id=\"footnote-anchor-1\" href=\"#footnote-1\">[1]</a> Here&apos;s another.<a id=\"footnote-anchor-2\" href=\"#footnote-2\">[2]</a>\n</p>",
		},
		{
			name:   "a footnote in a paragraph and a footnote at the end of the post should have anchor tags added correctly",
			input:  "Throwaway line\n\nThis paragraph references a footnote.[^1]\n\n[^1]: This is the reference.",
			output: "Throwaway line\n<p>\nThis paragraph references a footnote.<a id=\"footnote-anchor-1\" href=\"#footnote-1\">[1]</a>\n</p>\n\n<p id=\"footnote-1\">\n<a href=\"#footnote-anchor-1\">[1]</a>\n This is the reference.\n</p>",
		},
		{
			name:   "successive footnotes in text and at the end should be numbered correctly",
			input:  "Throwaway line\n\nThis paragraph references a footnote.[^1]\n\nThis paragraph[^2] also has a footnote.\n\n[^1]: This is the reference.\n[^2]: This is a footnote.",
			output: "Throwaway line\n<p>\nThis paragraph references a footnote.<a id=\"footnote-anchor-1\" href=\"#footnote-1\">[1]</a>\n</p>\n<p>\nThis paragraph<a id=\"footnote-anchor-2\" href=\"#footnote-2\">[2]</a> also has a footnote.\n</p>\n\n<p id=\"footnote-1\">\n<a href=\"#footnote-anchor-1\">[1]</a>\n This is the reference.\n</p>\n<p id=\"footnote-2\">\n<a href=\"#footnote-anchor-2\">[2]</a>\n This is a footnote.\n</p>",
		},
		{
			name:   "'#' in footnotes should not cause header tags to be added",
			input:  "Throwaway line\n\nThis paragraph references a footnote.[^1]\n\n[^1]: This is the reference, it has a url: https://this-is-not-a-real-url.blue/database?query=#a-query.",
			output: "Throwaway line\n<p>\nThis paragraph references a footnote.<a id=\"footnote-anchor-1\" href=\"#footnote-1\">[1]</a>\n</p>\n\n<p id=\"footnote-1\">\n<a href=\"#footnote-anchor-1\">[1]</a>\n This is the reference, it has a url: https://this&ndash;is&ndash;not&ndash;a&ndash;real&ndash;url.blue/database?query=#a&ndash;query.\n</p>",
		},
		{
			name:   "double-digit footnotes should be numbered correctly",
			input:  "Throwaway line\n\n[^1]\n[^2]\n[^3]\n[^4]\n[^5]\n[^6]\n[^7]\n[^8]\n[^9]\n[^10]\n[^11]\n[^12]\n\n[^1]: 1\n[^2]: 2\n[^3]: 3\n[^4]: 4\n[^5]: 5\n[^6]: 6\n[^7]: 7\n[^8]: 8\n[^9]: 9\n[^10]: 10\n[^11]: 11\n[^12]: 12",
			output: "Throwaway line\n\n<a id=\"footnote-anchor-1\" href=\"#footnote-1\">[1]</a>\n<a id=\"footnote-anchor-2\" href=\"#footnote-2\">[2]</a>\n<a id=\"footnote-anchor-3\" href=\"#footnote-3\">[3]</a>\n<a id=\"footnote-anchor-4\" href=\"#footnote-4\">[4]</a>\n<a id=\"footnote-anchor-5\" href=\"#footnote-5\">[5]</a>\n<a id=\"footnote-anchor-6\" href=\"#footnote-6\">[6]</a>\n<a id=\"footnote-anchor-7\" href=\"#footnote-7\">[7]</a>\n<a id=\"footnote-anchor-8\" href=\"#footnote-8\">[8]</a>\n<a id=\"footnote-anchor-9\" href=\"#footnote-9\">[9]</a>\n<a id=\"footnote-anchor-10\" href=\"#footnote-10\">[10]</a>\n<a id=\"footnote-anchor-11\" href=\"#footnote-11\">[11]</a>\n<a id=\"footnote-anchor-12\" href=\"#footnote-12\">[12]</a>\n\n<p id=\"footnote-1\">\n<a href=\"#footnote-anchor-1\">[1]</a>\n 1\n</p>\n<p id=\"footnote-2\">\n<a href=\"#footnote-anchor-2\">[2]</a>\n 2\n</p>\n<p id=\"footnote-3\">\n<a href=\"#footnote-anchor-3\">[3]</a>\n 3\n</p>\n<p id=\"footnote-4\">\n<a href=\"#footnote-anchor-4\">[4]</a>\n 4\n</p>\n<p id=\"footnote-5\">\n<a href=\"#footnote-anchor-5\">[5]</a>\n 5\n</p>\n<p id=\"footnote-6\">\n<a href=\"#footnote-anchor-6\">[6]</a>\n 6\n</p>\n<p id=\"footnote-7\">\n<a href=\"#footnote-anchor-7\">[7]</a>\n 7\n</p>\n<p id=\"footnote-8\">\n<a href=\"#footnote-anchor-8\">[8]</a>\n 8\n</p>\n<p id=\"footnote-9\">\n<a href=\"#footnote-anchor-9\">[9]</a>\n 9\n</p>\n<p id=\"footnote-10\">\n<a href=\"#footnote-anchor-10\">[10]</a>\n 10\n</p>\n<p id=\"footnote-11\">\n<a href=\"#footnote-anchor-11\">[11]</a>\n 11\n</p>\n<p id=\"footnote-12\">\n<a href=\"#footnote-anchor-12\">[12]</a>\n 12\n</p>",
		},
		{
			name:   "footnotes at the end should be renumbered if footnotes in text were renumbered",
			input:  "Throwaway line\n\nThis paragraph references a footnote.[^2]\n\nThis paragraph[^1] also has a footnote.\n\n[^1]: This is the reference.\n[^2]: This is a footnote.",
			output: "Throwaway line\n<p>\nThis paragraph references a footnote.<a id=\"footnote-anchor-1\" href=\"#footnote-1\">[1]</a>\n</p>\n<p>\nThis paragraph<a id=\"footnote-anchor-2\" href=\"#footnote-2\">[2]</a> also has a footnote.\n</p>\n\n<p id=\"footnote-2\">\n<a href=\"#footnote-anchor-2\">[2]</a>\n This is the reference.\n</p>\n<p id=\"footnote-1\">\n<a href=\"#footnote-anchor-1\">[1]</a>\n This is a footnote.\n</p>",
		},
		{
			name:   "unordered lists should have <ul> tags and <li> tags",
			input:  "# Unordered List!\n\n- This is an unordered list with a - dash.\n- One,\n- Two,\n- Three.",
			output: "<h1> Unordered List!</h1>\n<p>\n<ul>\n<li> This is an unordered list with a &ndash; dash.</li>\n<li> One,</li>\n<li> Two,</li>\n<li> Three.</li>\n</ul>\n</p>",
		},
		{
			name:   "the head of a table should be added correctly",
			input:  "| Table | Head |",
			output: "<table class=\"table is-hoverable\">\n<thead>\n<tr>\n<th> Table </th>\n<th> Head </th>\n</tr>\n</thead>\n<tbody>\n</tbody>\n</table>",
		},
		{
			name:   "the border line after the header of a table should be added correctly",
			input:  "| Table | Head |\n|--|--|",
			output: "<table class=\"table is-hoverable\">\n<thead>\n<tr>\n<th> Table </th>\n<th> Head </th>\n</tr>\n</thead>\n<tbody>\n</tbody>\n</table>",
		},
		{
			name:   "simple tables without markdown characters in them should have the appropriate table tags added",
			input:  "| col name one | col name two |\n|-|-|\n| row contents one | row contents two |",
			output: "<table class=\"table is-hoverable\">\n<thead>\n<tr>\n<th> col name one </th>\n<th> col name two </th>\n</tr>\n</thead>\n<tbody>\n<tr>\n<td> row contents one </td>\n<td> row contents two </td>\n</tr>\n</tbody>\n</table>",
		},
		{
			name:   "a table with html entities should have them replaced",
			input:  "| col name one | col name two |\n|-|-|\n| A non-entity / | Some entities - ' |\n| < More entities > | \"And I quote...\" |",
			output: "<table class=\"table is-hoverable\">\n<thead>\n<tr>\n<th> col name one </th>\n<th> col name two </th>\n</tr>\n</thead>\n<tbody>\n<tr>\n<td> A non&ndash;entity / </td>\n<td> Some entities &ndash; &apos; </td>\n</tr>\n<tr>\n<td> &lt; More entities &gt; </td>\n<td> &quot;And I quote...&quot; </td>\n</tr>\n</tbody>\n</table>",
		},
		{
			name:   "images should be placed into <figure> and <img> tags",
			input:  "![[image_name.png]]",
			output: "<figure class=\"image\">\n<img src=\"/directory_name/image_name.png\">\n</figure>",
		},
		{
			name:   "! at the end of a file should be written correctly.",
			input:  "A sentence!",
			output: "A sentence!",
		},
		{
			name:   "html elements in a header should be replaced correctly",
			input:  "# A header with html < > \" ' - elements",
			output: "<h1> A header with html &lt; &gt; &quot; &apos; &ndash; elements</h1>",
		},
		{
			name:   "unordered lists should have their tags closed correctly before the next piece of content",
			input:  "# Header\n\n- Unordered\n- List\n\nEnd of file.",
			output: "<h1> Header</h1>\n<p>\n<ul>\n<li> Unordered</li>\n<li> List</li>\n</ul>\n</p>\n<p>\nEnd of file.\n</p>",
		},
		{
			name:   "integration test: a small file",
			input:  "# Introduction\n\n## A Small File\n\nThis is a *small* file. It contains - neigh - requires the program to correctly translate a variety of different Obsidian Markdown elements into the HTML elements I want.\n\n![[image_name.png]]\n\nFor example:\n\n- paragraphs[^1]\n- \"0 < 1\"\n- \"2 > 1\"\n- **and**\n- ***headings***\n- `Code blocks`\n\n```Pseudocode\nfn removeCharacterFromList(remList list, charToRemove char) list {\n    match remList {\n        case x::[]:\n            match x {\n                charToRemove: []\n                _: x\n            }\n        case x::xs:\n            match x {\n                charToRemove: removeCharacterFromList(xs, charToRemove)\n                _: x::removeCharacterFromList(xs, charToRemove)\n            }\n    }\n}\n\nremoveCharacterFromList(['a', 'b', 'c'], 'a')\n```\n\n## A table conclusion\n\nAnother footnote.[^2]\n\n| A table | must have | columns |\n|--|--|--|\n| and rows. | which may have an arbitrary amount of content | |\n\n[^1]: With footnotes!\n[^2]: Pseudocode.",
			output: "<h1> Introduction</h1>\n<p>\n<h2> A Small File</h2>\n</p>\n<p>\nThis is a <i>small</i> file. It contains &ndash; neigh &ndash; requires the program to correctly translate a variety of different Obsidian Markdown elements into the HTML elements I want.\n</p>\n<p>\n<figure class=\"image\">\n<img src=\"/directory_name/image_name.png\">\n</figure>\n</p>\n<p>\nFor example:\n</p>\n<p>\n<ul>\n<li> paragraphs<a id=\"footnote-anchor-1\" href=\"#footnote-1\">[1]</a></li>\n<li> &quot;0 &lt; 1&quot;</li>\n<li> &quot;2 &gt; 1&quot;</li>\n<li> <b>and</b></li>\n<li> <i><b>headings</b></i></li>\n<li> <code>Code blocks</code></li>\n</ul>\n</p>\n<p>\n<pre><code>\nfn removeCharacterFromList(remList list, charToRemove char) list {\n    match remList {\n        case x::[]:\n            match x {\n                charToRemove: []\n                _: x\n            }\n        case x::xs:\n            match x {\n                charToRemove: removeCharacterFromList(xs, charToRemove)\n                _: x::removeCharacterFromList(xs, charToRemove)\n            }\n    }\n}\n\nremoveCharacterFromList([&apos;a&apos;, &apos;b&apos;, &apos;c&apos;], &apos;a&apos;)\n</code></pre>\n</p>\n<p>\n<h2> A table conclusion</h2>\n</p>\n<p>\nAnother footnote.<a id=\"footnote-anchor-2\" href=\"#footnote-2\">[2]</a>\n</p>\n<p>\n<table class=\"table is-hoverable\">\n<thead>\n<tr>\n<th> A table </th>\n<th> must have </th>\n<th> columns </th>\n</tr>\n</thead>\n<tbody>\n<tr>\n<td> and rows. </td>\n<td> which may have an arbitrary amount of content </td>\n<td> </td>\n</tr>\n</tbody>\n</table>\n</p>\n<p id=\"footnote-1\">\n<a href=\"#footnote-anchor-1\">[1]</a>\n With footnotes!\n</p>\n<p id=\"footnote-2\">\n<a href=\"#footnote-anchor-2\">[2]</a>\n Pseudocode.\n</p>",
		},
		{
			name:   "an unordered list may contain italics tags, bold tags, and inline code blocks",
			input:  "For example:\n\n- paragraphs[^1]\n- \"0 < 1\"\n- \"2 > 1\"\n- **and**\n- ***headings***\n- `Code blocks`",
			output: "For example:\n<p>\n<ul>\n<li> paragraphs<a id=\"footnote-anchor-1\" href=\"#footnote-1\">[1]</a></li>\n<li> &quot;0 &lt; 1&quot;</li>\n<li> &quot;2 &gt; 1&quot;</li>\n<li> <b>and</b></li>\n<li> <i><b>headings</b></i></li>\n<li> <code>Code blocks</code></li>\n</ul>\n</p>",
		},
		{
			name:   "paragraph tags should be added correctly after an image is added",
			input:  "# Introduction\n\n![[image_name.png]]\n\nFor example:",
			output: "<h1> Introduction</h1>\n<p>\n<figure class=\"image\">\n<img src=\"/directory_name/image_name.png\">\n</figure>\n</p>\n<p>\nFor example:\n</p>",
		},
		{
			name:   "paragraph tags should be added correctly after an header in text",
			input:  "# Introduction\n\n## A Small File\n\nThis is a *small* file.",
			output: "<h1> Introduction</h1>\n<p>\n<h2> A Small File</h2>\n</p>\n<p>\nThis is a <i>small</i> file.\n</p>",
		},
		// Below are some additional tests for optional extensions.
		//{
		//	name:   "paragraph tags should be added correctly after an h2 title",
		//	input:  "# Introduction\n\n## A Small File\n\nThis is a *small* file.",
		//	output: "<h1> Introduction</h1>\n<p>\n</h2> A Small File</h2>\n</p>\n<p>\nThis is a <i>small</i> file.\n</p>",
		//},
		//{
		//	name:   "a table may have footnotes in it and should replace them correctly",
		//	input:  "| col name one | col name two |\n|-|-|\n| first row contents one[^1] | first row[^2] contents two |\n| second row contents[^3] one | second row contents two[^4] |",
		//	output: "<table class=\"table table-hover\">\n<thead>\n<tr>\n<th scope=\"col\"> col name one </th>\n<th scope=\"col\"> col name two </th>\n</tr>\n</thead>\n<tbody>\n<tr>\n<td> first row contents one<a id=\"footnote-anchor-1\" href=\"#footnote-1\">[1]</a> </td>\n<td> first row<a id=\"footnote-anchor-2\" href=\"#footnote-2\">[2]</a> contents two </td>\n</tr>\n<tr>\n<td> second row contents<a id=\"footnote-anchor-3\" href=\"#footnote-3\">[3]</a> one </td>\n<td> second row contents two<a id=\"footnote-anchor-4\" href=\"#footnote-4\">[4]</a> </td>\n</tr>\n</tbody>\n</table>",
		//},
		//{
		//	name:   "italics tags should correctly surround text with an '*' in it which has spaces either side",
		//	input:  "*This text contains * an asterisk.*",
		//	output: "<i>This text contains * an asterisk.</i>",
		//},
		//{
		//	name:   "a solitary '*' at the end should not create an italics tag",
		//	input:  "*This text contains* an asterisk.*",
		//	output: "<i>This text contains</i> an asterisk.*",
		//},
	}

	for i, tst := range testCases {
		res := convertMarkdownFileToBlogHTML(bytes.NewReader([]byte(tst.input)), imageDirectoryName)
		if res != tst.output {
			t.Errorf(
				"TestConvertMarkdownFileToBlogHTML test number: %d \nTest name: %s \nexpected: \n%s \nbut got: \n%s",
				i, tst.name, tst.output, res,
			)
		}
	}
}
