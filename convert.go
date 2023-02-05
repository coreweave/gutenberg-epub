package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
	"unicode"

	termbox "github.com/nsf/termbox-go"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/taylorskalyo/goreader/epub"
)

// parser is a part of the goreader repo for parsing epubs
type parser struct {
	tagStack  []atom.Atom
	tokenizer *html.Tokenizer
	doc       cellbuf
	items     []epub.Item
	sb        strings.Builder
}

// cellbuf is a part of the goreader repo for parsing epubs
type cellbuf struct {
	cells   []termbox.Cell
	width   int
	lmargin int
	col     int
	row     int
	space   bool
	fg, bg  termbox.Attribute
}

func main() {
	//flags used: -url is the url to scrape,
	// -data_dir is the directory to save the files to
	inputPTR := flag.String("inputDir", "./input",
		"directory that the book files will convert from. Defaults to './input'")

	outputPTR := flag.String("outputDir", "./output",
		"directory that the book files will convert to. Defaults to './output'")

	writeHeaderPtr := flag.Bool("writeHeader", true,
		"Saves the book title and author to the top of the file. Defaults to true")

	flag.Parse()
	ConvertEpubGo(*inputPTR, *outputPTR, *writeHeaderPtr)
}

// A lot of the actual parsing is done with this repo: https://github.com/taylorskalyo/goreader
func ConvertEpubGo(inputdir string, outputdir string, writeHeader bool) {
	//get all files in directory
	files, err := ioutil.ReadDir(inputdir)
	if err != nil {
		log.Fatal(err)
	}

	//we time the parsing
	start := time.Now()

	//we count the number of characters
	charCount := 0

	//for each file, if it is an epub, convert it to txt
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".epub") {
			filepath := inputdir + "/" + file.Name()

			//We use the goreader library to parse the epub
			rc, err := epub.OpenReader(filepath)
			if err != nil {
				panic(err)
			}
			defer rc.Close()
			// The rootfile (content.opf) lists all of the contents of an epub file.
			// There may be multiple rootfiles, although typically there is only one.
			book := rc.Rootfiles[0]

			//generate output file name and file
			outputFileName := strings.TrimSuffix(file.Name(), ".epub") + ".txt"
			outputFilePath := outputdir + "/" + outputFileName
			outputFile, err := os.Create(outputFilePath)
			if err != nil {
				log.Fatal(err)
			}
			defer outputFile.Close()

			// Print book title.
			fmt.Println("Parsing book: ", book.Title, "(file: ", file.Name()+")")

			//write the book title and author to the top of the file if writeHeader is true
			if writeHeader {
				outputFile.Write([]byte(book.Title + " by " + book.Metadata.Creator + " \n"))
				outputFile.Write([]byte("Published by " + book.Metadata.Publisher + " \n"))
				outputFile.Write([]byte("Language: " + book.Metadata.Language + " \n"))
				outputFile.Write([]byte("Description: " + book.Metadata.Description + " \n"))
				outputFile.Write([]byte("---------------------\n"))

			}

			//stringbuilder to hold the text instead of using goreader's cell system
			var sb strings.Builder

			//iterate through each chapter in the book
			for _, itemref := range book.Spine.Itemrefs {
				f, err := itemref.Open()
				if err != nil {
					panic(err)
				}

				//parse the chapter into the stringbuilder
				sbret, err := parseText(f, book.Manifest.Items, sb)
				if err != nil {
					log.Fatal(err)
				}

				//clean up returned string by removing tabs and extra spaces and newlines, other junk too
				chapterStr := strings.ReplaceAll(sbret.String(), "	", "")
				chapterStr = strings.ReplaceAll(chapterStr, "  ", " ")
				chapterStr = strings.ReplaceAll(chapterStr, "\r", "\n")
				chapterStr = strings.ReplaceAll(chapterStr, "\n\n\n", "\n")
				chapterStr = strings.ReplaceAll(chapterStr, "\n\n", "\n")
				chapterStr = strings.ReplaceAll(chapterStr, "\n\n", "\n")
				chapterStr = strings.TrimFunc(chapterStr, func(r rune) bool {
					return !unicode.IsGraphic(r)
				})

				//get the string from the stringbuilder
				charCount += len(chapterStr)

				//writes to file
				outputFile.Write([]byte(chapterStr))

				// Close the itemref.
				f.Close()

				//clear the stringbuilder
				sb.Reset()
			}

		}

	}
	if charCount > 0 {
		elapsed := time.Since(start)
		fmt.Printf("Parsing took %s, parsed %d characters at a rate of %d characters per second.\n", elapsed, charCount, int(float64(charCount)/elapsed.Seconds()))
	}
}

// parseText takes in html content via an io.Reader and returns a buffer
// containing only plain text.
func parseText(r io.Reader, items []epub.Item, sb strings.Builder) (strings.Builder, error) {
	tokenizer := html.NewTokenizer(r)
	doc := cellbuf{width: 80}
	p := parser{tokenizer: tokenizer, doc: doc, items: items, sb: sb}
	err := p.parse(r)
	if err != nil {
		return p.sb, err
	}
	return p.sb, nil
}

// parse walks an html document and renders elements to a cell buffer document.
func (p *parser) parse(io.Reader) (err error) {
	for {
		tokenType := p.tokenizer.Next()
		token := p.tokenizer.Token()
		switch tokenType {
		case html.ErrorToken:
			err = p.tokenizer.Err()
		case html.StartTagToken:
			p.tagStack = append(p.tagStack, token.DataAtom) // push element
			fallthrough
		case html.SelfClosingTagToken:
			p.handleStartTag(token)
		case html.TextToken:
			p.handleText(token)
		case html.EndTagToken:
			p.tagStack = p.tagStack[:len(p.tagStack)-1] // pop element
		}
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
	}
}

// handleText appends text elements to the parser buffer. It filters elements
// that should not be displayed as text (e.g. style blocks).
func (p *parser) handleText(token html.Token) {
	// Skip style tags
	if len(p.tagStack) > 0 && p.tagStack[len(p.tagStack)-1] == atom.Style {
		return
	}
	p.doc.style(p.tagStack)
	//I think the appendText is needed to properly parse the tags
	p.doc.appendText(string(token.Data))
	p.sb.WriteString(string(token.Data))

}

// handleStartTag appends text representations of non-text elements (e.g. image alt
// tags) to the parser buffer.
func (p *parser) handleStartTag(token html.Token) {
	switch token.DataAtom {
	case atom.Img:
		// Display alt text in place of images.
		for _, a := range token.Attr {
			switch atom.Lookup([]byte(a.Key)) {
			case atom.Alt:
				text := fmt.Sprintf("Alt text: %s", a.Val)
				p.doc.appendText(text)
				p.doc.row++
				p.doc.col = p.doc.lmargin
			case atom.Src:
				for _, item := range p.items {
					if item.HREF == a.Val {

						break
					}
				}
			}
		}
	case atom.Br:
		p.doc.row++
		p.doc.col = p.doc.lmargin
	case atom.H1, atom.H2, atom.H3, atom.H4, atom.H5, atom.H6, atom.Title,
		atom.Div, atom.Tr:
		p.doc.row += 2
		p.doc.col = p.doc.lmargin
	case atom.P:
		p.doc.row += 2
		p.doc.col = p.doc.lmargin
		p.doc.col += 2
	case atom.Hr:
		p.doc.row++
		p.doc.col = 0
		p.doc.appendText(strings.Repeat("-", p.doc.width))
	}
}

// style sets the foreground/background attributes for future cells in the cell
// buffer document based on HTML tags in the tag stack.
func (c *cellbuf) style(tags []atom.Atom) {
	fg := termbox.ColorDefault
	for _, tag := range tags {
		switch tag {
		case atom.B, atom.Strong, atom.Em:
			fg |= termbox.AttrBold
		case atom.I:
			fg |= termbox.ColorYellow
		case atom.Title:
			fg |= termbox.ColorRed
		case atom.H1:
			fg |= termbox.ColorMagenta
		case atom.H2:
			fg |= termbox.ColorBlue
		case atom.H3, atom.H4, atom.H5, atom.H6:
			fg |= termbox.ColorCyan
		}
	}
	c.fg = fg
}

// appendText appends text to the cell buffer document.
func (c *cellbuf) appendText(str string) {
	if len(str) <= 0 {
		return
	}
	if c.col < c.lmargin {
		c.col = c.lmargin
	}

	scanner := bufio.NewScanner(strings.NewReader(str))
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		if c.col != c.lmargin && c.space {
			c.col++
		}
		word := []rune(scanner.Text())
		if len(word) > c.width-c.col {
			c.row++
			c.col = c.lmargin
		}
		for _, r := range word {
			c.setCell(c.col, c.row, r, c.fg, c.bg)
			c.col++
		}
		//c.space = true
	}
}

// setCell changes a cell's attributes in the cell buffer document at the given
// position.
func (c *cellbuf) setCell(x, y int, ch rune, fg, bg termbox.Attribute) {
	// Grow in steps of 1024 when out of space.
	for y*c.width+x >= len(c.cells) {
		c.cells = append(c.cells, make([]termbox.Cell, 1024)...)
	}
	c.cells[y*c.width+x] = termbox.Cell{Ch: ch, Fg: fg, Bg: bg}
}
