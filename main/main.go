package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
)

var htmlEntityMap = map[rune]string{
	'\'': "&apos;",
	'<':  "&lt;",
	'>':  "&gt;",
	'"':  "&quot;",
	'-':  "&ndash;",
}

// Default directory name. To be overwritten by user with
// a command-line flag.
var imageDirectoryName = "/directory_name"
var footnoteNumberMap = map[int]int{}
var inlineFootnoteNumber int

func main() {
	pathName := os.Args
	br := getByteReadForFile(pathName[1])
	res := convertMarkdownFileToBlogHTML(br, pathName[3])
	saveToFile(res, pathName[2])
}

func getByteReadForFile(pathAndFilename string) *bytes.Reader {
	bytesReadIn, err := os.ReadFile(pathAndFilename)
	if err != nil {
		log.Fatal("unable to find file:", err)
	}

	// remove carriage returns
	bytesReadIn = bytes.ReplaceAll(bytesReadIn, []byte{'\r'}, []byte{})

	return bytes.NewReader(bytesReadIn)
}

// https://gobyexample.com/writing-files
func saveToFile(res string, outputPathAndFileName string) {
	f, err := os.Create(outputPathAndFileName)
	if err != nil {
		log.Fatal("unable to create file:", err)
	}
	defer f.Close()

	numBytesWritten, err := f.WriteString(res)
	if err != nil {
		log.Fatal("error when writing to file:", err)
	}
	fmt.Printf("wrote %d bytes to file", numBytesWritten)

	f.Sync()
}

func convertMarkdownFileToBlogHTML(br *bytes.Reader, newImageDirectoryName string) string {
	sb := strings.Builder{}

	lastCharacterWasANewLine := false
	thereIsAParagraphToClose := false
	addNewLineCharBeforeOpeningPara := false
	inlineFootnoteNumber = 0
	imageDirectoryName = newImageDirectoryName

	for {
		r, _, err := br.ReadRune()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("unable to read rune:", err)
		}

		switch r {
		case '#':
			addHeaderTagsOrPoundRune(br, &sb, lastCharacterWasANewLine)
			lastCharacterWasANewLine = true
			addNewLineCharBeforeOpeningPara = true

		case ' ':
			sb.WriteRune(r)
			lastCharacterWasANewLine = false

		case '\n':
			addNewLine(br, &sb, &thereIsAParagraphToClose, &lastCharacterWasANewLine, &addNewLineCharBeforeOpeningPara)
			lastCharacterWasANewLine = true

		case '*':
			addItalicsAndOrBoldTags(br, &sb)
			lastCharacterWasANewLine = false

		case '\'':
			sb.WriteString(htmlEntityMap[r])
			lastCharacterWasANewLine = false

		case '<':
			sb.WriteString(htmlEntityMap[r])
			lastCharacterWasANewLine = false

		case '>':
			sb.WriteString(htmlEntityMap[r])
			lastCharacterWasANewLine = false

		case '"':
			sb.WriteString(htmlEntityMap[r])
			lastCharacterWasANewLine = false

		case '-':
			if lastCharacterWasANewLine {
				addUnorderedList(br, &sb)
				addNewLineCharBeforeOpeningPara = true
				lastCharacterWasANewLine = true

			} else {
				sb.WriteString(htmlEntityMap[r])
				lastCharacterWasANewLine = false
			}

		case '`':
			addCodeBlock(br, &sb)
			lastCharacterWasANewLine = false

		case '[':
			addFootNote(br, &sb)
			lastCharacterWasANewLine = false

		case '|':
			addTable(br, &sb)
			lastCharacterWasANewLine = true
			addNewLineCharBeforeOpeningPara = true

		case '!':
			addImageTags(br, &sb)
			lastCharacterWasANewLine = true
			addNewLineCharBeforeOpeningPara = true

		default:
			lastCharacterWasANewLine = false
			sb.WriteRune(r)
		}
	}

	if thereIsAParagraphToClose {
		sb.WriteRune('\n')
		sb.WriteString("</p>")
	}

	return sb.String()
}

func addHeaderTagsOrPoundRune(br *bytes.Reader, sb *strings.Builder, lastCharacterWasANewLine bool) {
	if sb.Len() == 0 {
		addHeaderTags(br, sb)

	} else if lastCharacterWasANewLine {
		addHeaderTags(br, sb)

	} else {
		sb.WriteRune('#')
	}
}

func addHeaderTags(br *bytes.Reader, sb *strings.Builder) {
	var finishedCountingHeaderTagsForLine = false
	var headerCount = 1 // assume 1, as a '#' has been seen in order to get to here
	var nextR rune

	for nextR != '\n' {
		nextR, _, err := br.ReadRune()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("unable to read next rune in header:", err)
		}
		if !finishedCountingHeaderTagsForLine {
			if nextR == '#' {
				headerCount++
			}
			if nextR == ' ' {
				finishedCountingHeaderTagsForLine = true
				sb.WriteString("<h" + strconv.Itoa(headerCount) + ">")
				sb.WriteRune(' ')
			}

		} else if nextR == '\n' {
			break

		} else {
			addRuneOrHTMLEntity(nextR, sb)
		}
	}

	sb.WriteString("</h" + strconv.Itoa(headerCount) + ">")
}

func addNewLine(br *bytes.Reader, sb *strings.Builder, thereIsAParagraphToClose *bool, lastCharacterWasANewLine *bool, thereIsAnUnorderedListOpen *bool) {
	if *thereIsAParagraphToClose {
		sb.WriteRune('\n')
		sb.WriteString("</p>")
		*thereIsAParagraphToClose = false
	}
	if *lastCharacterWasANewLine {
		nextR, _, err := br.ReadRune()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Fatal("unable to read rune:", err)
		}
		if nextR == '[' {
			err = br.UnreadRune()
			if err != nil {
				log.Fatal("unable to unread rune:", err)
			}

			sb.WriteRune('\n')
			return
		}
		if *thereIsAnUnorderedListOpen {
			sb.WriteRune('\n')
			*thereIsAnUnorderedListOpen = false
		}

		sb.WriteString("<p>")
		*thereIsAParagraphToClose = true

		err = br.UnreadRune()
		if err != nil {
			log.Fatal("unable to unread rune:", err)
		}

	} else {
		*lastCharacterWasANewLine = true
	}

	sb.WriteRune('\n')
}

// Optional extension: Extend to recursively allow for italics and bold text to exist within each other.
func addItalicsAndOrBoldTags(br *bytes.Reader, sb *strings.Builder) {
	asteriskCount := 1
	asteriskCountNeededToCloseTags := 0
	stillCountingAsterisks := true
	italicsOrBoldTagOpen := false

	for {
		nextR, _, err := br.ReadRune()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("unable to read ahead by one rune when adding italics or bold tags:", err)
		}

		if nextR == '*' {
			if stillCountingAsterisks {
				asteriskCount++

				if asteriskCount == asteriskCountNeededToCloseTags {
					break
				}

			} else {
				addRuneOrHTMLEntity(nextR, sb)
			}
		}

		if nextR != '*' {
			if !italicsOrBoldTagOpen {
				stillCountingAsterisks = false

				switch asteriskCount {
				case 1:
					sb.WriteString("<i>")
					asteriskCountNeededToCloseTags = 1
				case 2:
					sb.WriteString("<b>")
					asteriskCountNeededToCloseTags = 2
				case 3:
					sb.WriteString("<i><b>")
					asteriskCountNeededToCloseTags = 3
				}
				asteriskCount = 0

				addRuneOrHTMLEntity(nextR, sb)
				italicsOrBoldTagOpen = true

			} else {
				stillCountingAsterisks = true
				addRuneOrHTMLEntity(nextR, sb)
			}
		}
	}

	switch asteriskCount {
	case 1:
		sb.WriteString("</i>")
	case 2:
		sb.WriteString("</b>")
	case 3:
		sb.WriteString("</b></i>")
	}
}

func addUnorderedList(br *bytes.Reader, sb *strings.Builder) {
	var lastCharacterWasANewLine = false

	sb.WriteString("<ul>")
	sb.WriteRune('\n')
	sb.WriteString("<li>")

	for {
		nextR, _, err := br.ReadRune()
		if err == io.EOF {
			sb.WriteString("</li>")
			sb.WriteRune('\n')
			break
		}
		if err != nil {
			log.Fatal("unable to read rune when creating an unordered list:", err)
		}

		if nextR == '-' {
			if lastCharacterWasANewLine {
				sb.WriteString("<li>")

			} else {
				sb.WriteString(htmlEntityMap[nextR])
			}
			lastCharacterWasANewLine = false

		} else if nextR == '\n' {
			if lastCharacterWasANewLine {
				err = br.UnreadRune()
				if err != nil {
					log.Fatal("unable to unread rune at end of unordered list:", err)
				}
				break

			} else {
				sb.WriteString("</li>")
				sb.WriteRune('\n')
				lastCharacterWasANewLine = true
			}
		} else if nextR == '[' {
			addFootNote(br, sb)

		} else if nextR == '*' {
			addItalicsAndOrBoldTags(br, sb)

		} else if nextR == '`' {
			addCodeBlock(br, sb)

		} else {
			addRuneOrHTMLEntity(nextR, sb)
			lastCharacterWasANewLine = false
		}
	}

	sb.WriteString("</ul>")
}

func addCodeBlock(br *bytes.Reader, sb *strings.Builder) {
	var numberOfCurrentBackQuotes = 1
	var thereIsACodeBlockOpen = false

	for {
		nextR, _, err := br.ReadRune()
		if err == io.EOF {
			if thereIsACodeBlockOpen {
				sb.WriteString("</code>")
				if numberOfCurrentBackQuotes == 6 {
					sb.WriteString("</pre>")
				}
				thereIsACodeBlockOpen = false
				numberOfCurrentBackQuotes = 0
			}
			return
		}
		if err != nil {
			log.Fatal("unable to read next rune:", err)
		}

		if nextR == '`' {
			numberOfCurrentBackQuotes++

			if thereIsACodeBlockOpen {
				if numberOfCurrentBackQuotes == 2 {
					sb.WriteString("</code>")
					return
				}
			}

			if numberOfCurrentBackQuotes == 3 {
				sb.WriteString("<pre><code>")
				thereIsACodeBlockOpen = true

				for nextR != '\n' {
					nextR, _, err = br.ReadRune()
					if err == io.EOF {
						break
					}
					if err != nil {
						log.Fatal("unable to read next rune:", err)
					}
				}
				sb.WriteRune('\n')
			}

			if numberOfCurrentBackQuotes == 6 {
				sb.WriteString("</code></pre>")
				return
			}

		} else {
			if !thereIsACodeBlockOpen {
				if numberOfCurrentBackQuotes == 1 {
					sb.WriteString("<code>")
					thereIsACodeBlockOpen = true

				} else if numberOfCurrentBackQuotes == 2 {
					sb.WriteString("</code>")
					return
				}
			}

			addRuneOrHTMLEntity(nextR, sb)
		}
	}
}

func addRuneOrHTMLEntity(r rune, sb *strings.Builder) {
	if slices.Contains([]rune{'\'', '<', '>', '"', '-'}, r) {
		sb.WriteString(htmlEntityMap[r])
	} else {
		sb.WriteRune(r)
	}
}

func addFootNote(br *bytes.Reader, sb *strings.Builder) {
	inTextFootnoteNumber := strings.Builder{}
	for {
		nextR, _, err := br.ReadRune()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("unable to read next rune:", err)
		}
		if nextR == ']' {

			nextR, _, err = br.ReadRune()
			if err == io.EOF {
				sb.WriteString(
					"<a id=\"footnote-anchor-" +
						strconv.Itoa(inlineFootnoteNumber) +
						"\" href=\"#footnote-" +
						strconv.Itoa(inlineFootnoteNumber) +
						"\">[" + strconv.Itoa(inlineFootnoteNumber) +
						"]</a>",
				)
				break
			}
			if err != nil {
				log.Fatal("unable to read next rune:", err)
			}
			if nextR == ':' {
				footnoteNumber, err := strconv.Atoi(inTextFootnoteNumber.String())
				if err != nil {
					log.Fatal("unable to convert string to number:", err)
				}

				sb.WriteString(
					"<p id=\"footnote-" +
						strconv.Itoa(footnoteNumberMap[footnoteNumber]) +
						"\">\n<a href=\"#footnote-anchor-" +
						strconv.Itoa(footnoteNumberMap[footnoteNumber]) +
						"\">[" +
						strconv.Itoa(footnoteNumberMap[footnoteNumber]) +
						"]</a>",
				)
				sb.WriteRune('\n')

				for nextR != '\n' {
					nextR, _, err = br.ReadRune()
					if err == io.EOF {
						sb.WriteRune('\n')
						break
					}
					if err != nil {
						log.Fatal("unable to read next rune:", err)
					}

					addRuneOrHTMLEntity(nextR, sb)
				}

				sb.WriteString("</p>")
				if err == io.EOF {
					break
				}
				sb.WriteRune('\n')

			} else {
				sb.WriteString(
					"<a id=\"footnote-anchor-" +
						strconv.Itoa(inlineFootnoteNumber) +
						"\" href=\"#footnote-" +
						strconv.Itoa(inlineFootnoteNumber) +
						"\">[" + strconv.Itoa(inlineFootnoteNumber) +
						"]</a>",
				)

				footnoteOriginalNumber, err := strconv.Atoi(inTextFootnoteNumber.String())
				if err != nil {
					log.Fatal("unable to convert string to number:", err)
				}
				footnoteNumberMap[footnoteOriginalNumber] = inlineFootnoteNumber

				err = br.UnreadRune()
				if err != nil {
					log.Fatal("unable to unread rune:", err)
				}
			}

			break
		} else if nextR == '^' {
			inlineFootnoteNumber++

		} else if nextR != '^' {
			inTextFootnoteNumber.WriteRune(nextR)
		}
	}
}

func addTable(br *bytes.Reader, sb *strings.Builder) {
	sb.WriteString("<table class=\"table is-hoverable\">")
	sb.WriteRune('\n')

	addTableHeader(br, sb)
	skipTableHeaderLine(br)
	addTableBody(br, sb)

	sb.WriteString("</table>")
}

func addTableHeader(br *bytes.Reader, sb *strings.Builder) {
	sb.WriteString("<thead>")
	sb.WriteRune('\n')
	sb.WriteString("<tr>")
	sb.WriteRune('\n')
	sb.WriteString("<th>")

	var nextR rune
	var err error
	for nextR != '\n' {
		nextR, _, err = br.ReadRune()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("unable to read next rune:", err)
		}
		if nextR == '|' {
			nextR, _, err = br.ReadRune()
			if err == io.EOF {
				sb.WriteString("</th>")
				sb.WriteRune('\n')
				break
			}
			if err != nil {
				log.Fatal("unable to read next rune:", err)
			}
			if nextR == '\n' {
				sb.WriteString("</th>")
				sb.WriteRune('\n')
				break
			}
			err = br.UnreadRune()
			if err != nil {
				log.Fatal("unable to unread rune:", err)
			}
			sb.WriteString("</th>")
			sb.WriteRune('\n')
			sb.WriteString("<th>")

		} else {
			addRuneOrHTMLEntity(nextR, sb)
		}
	}
	sb.WriteString("</tr>")
	sb.WriteRune('\n')
	sb.WriteString("</thead>")
	sb.WriteRune('\n')
}

func skipTableHeaderLine(br *bytes.Reader) {
	var afterR rune
	var err error

	for afterR != '\n' {
		afterR, _, err = br.ReadRune()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("unable to read rune:", err)
		}
	}
}

func addTableBody(br *bytes.Reader, sb *strings.Builder) {
	sb.WriteString("<tbody>")
	sb.WriteRune('\n')

	nextR, _, err := br.ReadRune()
	if err == io.EOF {
		sb.WriteString("</tbody>")
		sb.WriteRune('\n')
		return
	}
	if err != nil {
		log.Fatal("unable to read next rune:", err)
	}
	if nextR == '|' {
		sb.WriteString("<tr>")
		sb.WriteRune('\n')
		sb.WriteString("<td>")

		numOfConsecutiveNewLines := 0

		for numOfConsecutiveNewLines < 2 {
			nextR, _, err = br.ReadRune()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal("unable to read next rune:", err)
			}
			if nextR == '\n' {
				numOfConsecutiveNewLines++

				if numOfConsecutiveNewLines < 2 {
					sb.WriteString("</tr>")
					sb.WriteRune('\n')
				}

				if numOfConsecutiveNewLines == 2 {
					err = br.UnreadRune()
					if err != nil {
						log.Fatal("unable to unread new line character after table:", err)
					}
					break
				}

			} else {
				if numOfConsecutiveNewLines == 1 {
					sb.WriteString("<tr>")
					sb.WriteRune('\n')
					sb.WriteString("<td>")

				} else {
					if nextR == '|' {
						sb.WriteString("</td>")

						nextR, _, err = br.ReadRune()
						if err == io.EOF {
							sb.WriteRune('\n')
							sb.WriteString("</tr>")
							sb.WriteRune('\n')
							break
						}
						if err != nil {
							log.Fatal("unable to read next rune:", err)
						}
						sb.WriteRune('\n')
						if nextR != '\n' {
							sb.WriteString("<td>")
						}
						err = br.UnreadRune()
						if err != nil {
							log.Fatal("unable to unread rune:", err)
						}

					} else {
						addRuneOrHTMLEntity(nextR, sb)
					}
				}
				numOfConsecutiveNewLines = 0
			}
		}
	}

	sb.WriteString("</tbody>")
	sb.WriteRune('\n')
}

func addImageTags(br *bytes.Reader, sb *strings.Builder) {
	// Assumes structure of ![[image_name.png]]
	nextR, _, err := br.ReadRune()
	if err == io.EOF {
		sb.WriteRune('!')
		return
	}
	if err != nil {
		log.Fatal("unable to read rune:", err)
	}
	if nextR != '[' {
		sb.WriteRune('!')
		err = br.UnreadRune()
		if err != nil {
			log.Fatal("unable to unread rune:", err)
		}
		return
	}

	var imageNameAndExtension = strings.Builder{}

	for nextR != '\n' {
		nextR, _, err = br.ReadRune()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("unable to read rune:", err)
		}
		if nextR != '[' && nextR != ']' {
			imageNameAndExtension.WriteRune(nextR)
		}
		if nextR == ']' {
			_, _, err = br.ReadRune()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal("unable to read last ] of an image:", err)
			}
			break
		}
	}
	_, _, err = br.ReadRune()
	if err != nil && err != io.EOF {
		log.Fatal("unable to read rune:", err)
	}

	sb.WriteString("<figure class=\"image\">")
	sb.WriteRune('\n')
	sb.WriteString("<img src=\"" + imageDirectoryName + "/" + imageNameAndExtension.String() + "\">")
	sb.WriteRune('\n')
	sb.WriteString("</figure>")
}
