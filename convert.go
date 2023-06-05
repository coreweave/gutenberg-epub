package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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

// tracks the metadata of the book
type metadata struct {
	title       string
	author      string
	publisher   string
	language    string
	description string
	charCount   int
	filename    string
	identifier  string
	categories  []string
	bookType    string
	format      string
	source      string
	relation    string
	coverage    string
	rights      string
}

//tracks the config of the program
type programConfig struct {
	writeHeader       bool
	writeMetadata     bool
	cleanOutput       bool
	seperateFolders   bool
	stopEarly         int
	silent            bool
	skipCopyRight     bool
	gutenbergCleaning bool
	createSubsets     string
}

// Mini struct for files
type fileTrack struct {
	name   string
	path   string
	isDir  bool
	isEpub bool
}

// Mini struct for counters
type programCounter struct {
	bookCount                     int
	fileCount                     int
	charCount                     int
	timeStart                     time.Time
	timeEnd                       time.Time
	finishedBooksCount            int
	skippedDueToCopyRight         int
	skippedDueToInsuffcientLength int
	charCleanedCount              int
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

	writeMetadataPtr := flag.Bool("writeMetadata", false,
		"Saves the book metadata to another file. Defaults to false")

	cleanOutputPtr := flag.Bool("cleanOutput", true,
		"Removes strange characters and spacing. Defaults to true")

	seperateFoldersPtr := flag.Bool("seperateFolders", false,
		"Creates a seperate folder for each book. Defaults to false")

	stopEarlyPtr := flag.Int("stopEarly", 0,
		"Stops after a certain number of books. Defaults to 0 (no limit)")

	silentPtr := flag.Bool("silent", false,
		"Doesn't print anything to the console. Defaults to false")

	skipCopyRightPtr := flag.Bool("skipCopyRight", false,
		"Skips books that have a copy right. Defaults to false")

	gutenbergCleaningPtr := flag.Bool("gutenbergCleaning", false,
		"Additions to the cleaning process for gutenberg books."+
			"Must be used with -cleanOutput. Defaults to false")

	createSubsetsPtr := flag.String("createSubsets", "book",
		"Creates subsets of the books based on the metadata."+
			"Options: author, category, book, categoryauthor. Defaults to 'book'")

	flag.Parse()

	//check createSubsets is valid
	if *createSubsetsPtr != "author" && *createSubsetsPtr != "category" &&
		*createSubsetsPtr != "book" && *createSubsetsPtr != "categoryauthor" {
		fmt.Println("Error: createSubsets must be one of the following: author, category, book, categoryauthor")
		*createSubsetsPtr = "book"
		return
	}

	config := programConfig{
		writeHeader:       *writeHeaderPtr,
		writeMetadata:     *writeMetadataPtr,
		cleanOutput:       *cleanOutputPtr,
		seperateFolders:   *seperateFoldersPtr,
		stopEarly:         *stopEarlyPtr,
		silent:            *silentPtr,
		skipCopyRight:     *skipCopyRightPtr,
		gutenbergCleaning: *gutenbergCleaningPtr,
		createSubsets:     *createSubsetsPtr,
	}
	counters := programCounter{
		bookCount:                     0,
		fileCount:                     0,
		charCount:                     0,
		timeStart:                     time.Now(),
		timeEnd:                       time.Now(),
		finishedBooksCount:            0,
		skippedDueToCopyRight:         0,
		charCleanedCount:              0,
		skippedDueToInsuffcientLength: 0,
	}
	//Write params
	if !config.silent {
		fmt.Println("Input Directory: ", *inputPTR)
		fmt.Println("Output Directory: ", *outputPTR)
		fmt.Println("Write Header: ", config.writeHeader)
		fmt.Println("Write Metadata: ", config.writeMetadata)
		fmt.Println("Clean Output: ", config.cleanOutput)
		fmt.Println("Seperate Folders: ", config.seperateFolders)
		fmt.Println("Stop Early: ", config.stopEarly)
		fmt.Println("Silent: ", config.silent)
		fmt.Println("Skip Copy Right: ", config.skipCopyRight)
		fmt.Println("Gutenberg Cleaning: ", config.gutenbergCleaning)
		fmt.Println("Create Subsets: ", config.createSubsets)
		fmt.Println("------------\nStarting...\n")
	}

	//create output directory if it doesn't exist
	if _, err := os.Stat(*outputPTR); os.IsNotExist(err) {
		os.Mkdir(*outputPTR, 0755)
	}

	//get all files in directory
	files := aquireEpubFilePaths(*inputPTR, config, &counters)

	ConvertEpubGo(files, *inputPTR, *outputPTR, config, &counters)
}

func aquireEpubFilePaths(inputdir string, config programConfig, counters *programCounter) []fileTrack {
	//get all files in directory recursively
	files := []fileTrack{}
	err := filepath.Walk(inputdir, func(path string, info fs.FileInfo, err error) error {
		fi := new(fileTrack)
		fi.name = info.Name()
		fi.path = path
		fi.isDir = info.IsDir()
		fi.isEpub = strings.HasSuffix(info.Name(), ".epub")
		counters.fileCount++
		if !fi.isDir && fi.isEpub {
			files = append(files, *fi)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	//trim files if stopEarly is set
	if config.stopEarly != 0 && config.stopEarly < len(files) {
		files = files[:config.stopEarly]
	}

	counters.bookCount = len(files)
	fmt.Printf("Converting %d files\n", len(files))
	return files
}

func checkMetaForCopyright(meta metadata) bool {
	if strings.Contains(meta.rights, "copy") || strings.Contains(meta.rights, "Copyrighted") {
		return true
	}
	return false
}

// A lot of the actual parsing is done with this repo: https://github.com/taylorskalyo/goreader
func ConvertEpubGo(files []fileTrack, inputdir string, outputdir string, config programConfig, counters *programCounter) {
	//we time the parsing
	counters.timeStart = time.Now()

	//for each file, if it is an epub, convert it to txt
	for _, file := range files {
		if strings.HasSuffix(file.name, ".epub") {
			//fmt.Printf("Open files %d\n", countOpenFiles()) //debugging
			//We use the goreader library to parse the epub
			rc, err := epub.OpenReader(file.path)
			if err != nil {
				panic(err)
			}
			defer rc.Close()
			// The rootfile (content.opf) lists all of the contents of an epub file.
			// There may be multiple rootfiles, although typically there is only one.
			book := rc.Rootfiles[0]

			// Print book title.
			if !config.silent {
				fmt.Println("Parsing book: ", book.Title, "(file: ", file.name+")")
			}

			//stringbuilder to hold the text instead of using goreader's cell system
			var sb strings.Builder

			bookstr := ""
			//iterate through each chapter in the book
			for _, itemref := range book.Spine.Itemrefs {
				f, err := itemref.Open()
				if err != nil {
					log.Fatal(fmt.Printf("Error opening %s: %s\n", itemref.ID, err))
				}

				//parse the chapter into the stringbuilder
				sbret, err := parseText(f, book.Manifest.Items, sb)
				if err != nil {
					log.Fatal(err)
				}
				bookstr += "CHAPTER_SEPERATOR"
				bookstr += sbret.String()

				// Close the itemref.
				f.Close()

				//clear the stringbuilder
				sb.Reset()
			}

			//clean the text if cleanOutput is true
			lenBefore := (len(bookstr))
			// count the number of characters
			counters.charCount += len(bookstr)
			bookstr = cleanEpubString(bookstr, config)
			//count the number of characters removed
			counters.charCleanedCount += lenBefore - len(bookstr)
			fmt.Printf("Removed %d characters from %d characters\n", lenBefore-len(bookstr), lenBefore)

			//if length is less than 2000 characters, skip the file
			if len(bookstr) < 2000 {
				fmt.Printf("Skipping file %s, too short (%d characters)\n", file.name, len(bookstr))
				counters.skippedDueToInsuffcientLength++
				continue
			}

			// Write metadata to a separate file if writeMetadata is true
			bookMeta := new(metadata)
			bookMeta.title = book.Title
			bookMeta.author = book.Metadata.Creator
			bookMeta.publisher = book.Metadata.Publisher
			bookMeta.language = book.Metadata.Language
			bookMeta.description = book.Metadata.Description
			bookMeta.filename = ""
			bookMeta.charCount = counters.charCount
			bookMeta.format = book.Metadata.Format
			bookMeta.categories = []string{}
			fmt.Printf("Categories: %s\n", book.Metadata.Subject)
			bookMeta.categories = append(bookMeta.categories, book.Metadata.Subject)
			//parse seperated categories
			if len(book.Metadata.Subject) > 0 && strings.Contains(book.Metadata.Subject, " -- ") {
				bookMeta.categories = strings.Split(book.Metadata.Subject, " -- ")
			}

			bookMeta.identifier = book.Metadata.Identifier
			bookMeta.relation = book.Metadata.Relation
			bookMeta.coverage = book.Metadata.Coverage
			bookMeta.rights = book.Metadata.Rights

			//if createSubsets is set to book, we don't change the output directory
			//if it is set to author, we create a folder for each author
			//if it is set to category, we create a folder for each category
			//if it is set to categoryauthor, we create a folder for each category and then a folder for each author in that category

			//generate output file name and file
			outputFileName := strings.TrimSuffix(file.name, ".epub") + ".txt"
			outputFilePath := ""
			seperateFoldersExtension := ""
			if config.seperateFolders {
				seperateFoldersExtension = strings.TrimSuffix(file.name, ".epub")
			}

			if config.createSubsets == "book" {
				outputFilePath = outputdir + "/" + seperateFoldersExtension + "/" + outputFileName
			} else if config.createSubsets == "author" {
				outputFilePath = outputdir + "/" + bookMeta.author + "/" + seperateFoldersExtension + "/" + outputFileName
			} else if config.createSubsets == "category" {
				outputFilePath = outputdir + "/" + bookMeta.categories[0] + "/" + seperateFoldersExtension + "/" + outputFileName
			} else if config.createSubsets == "categoryauthor" {
				outputFilePath = outputdir + "/" + bookMeta.categories[0] + "/" + bookMeta.author + "/" + seperateFoldersExtension + "/" + outputFileName
			} else {
				outputFilePath = outputdir + "/" + seperateFoldersExtension + "/" + outputFileName
			}

			fmt.Printf("Output file path: %s\n", outputFilePath)

			if config.skipCopyRight {
				isRestricted := checkMetaForCopyright(*bookMeta)
				if isRestricted {
					//close the file
					if rc != nil {
						rc.Close()
					}

					if !config.silent {
						fmt.Println("Skipping restricted book: ", book.Title, "(file: ", file.name+")")
					}
					counters.skippedDueToCopyRight++
					continue

				}
			}

			//creates the path including the folders if they don't exist
			err = os.MkdirAll(filepath.Dir(outputFilePath), os.ModePerm)
			if err != nil {
				log.Fatal(err)
			}

			outputFile, err := os.Create(outputFilePath)
			if err != nil {
				log.Fatal(err)
			}

			if config.writeMetadata {
				writeMetadataToFile(bookMeta, outputFilePath, config)
			}

			//write the book title and author to the top of the file if writeHeader is true
			if config.writeHeader {
				header := buildMetadataHeader(bookMeta)
				outputFile.Write([]byte(header))
			}

			//write the book to the file
			outputFile.Write([]byte(bookstr))
			counters.finishedBooksCount++

			//close the file if not already closed
			if outputFile != nil {
				outputFile.Close()
			}
			if rc != nil {
				rc.Close()
			}

		}

	}

	if counters.charCount > 0 {
		counters.timeEnd = time.Now()
		elapsed := counters.timeEnd.Sub(counters.timeStart)
		fmt.Printf("--------------------\n")
		fmt.Printf("Parsing took %s, parsed %d characters at a rate of %d characters per second.\n", elapsed, counters.charCount, int(float64(counters.charCount)/elapsed.Seconds()))
		fmt.Printf("Cleaned %d characters, %% of characters removed: %f%%\n", counters.charCleanedCount, float64(counters.charCleanedCount)/float64(counters.charCount)*100)
		fmt.Printf("Parsed %d books, %d finished and %d skipped due to copy right, %d skipped due to insufficient length after cleaning (2000 char).\n", counters.bookCount, counters.finishedBooksCount, counters.skippedDueToCopyRight, counters.skippedDueToInsuffcientLength)
	}
}

// writeMetadataToFile writes the metadata of a book to a file.
func writeMetadataToFile(bookMeta *metadata, outputdir string, config programConfig) {
	if !config.writeMetadata {
		return
	}
	//generate output file name and file
	outputFilePath := strings.TrimSuffix(outputdir, ".txt") + ".metadata"

	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error creating metadata file: %s", err))
	}
	defer outputFile.Close()

	//write the book title and author to the top of the file if writeHeader is true
	header := buildMetadataHeader(bookMeta)
	outputFile.Write([]byte(header))
	if outputFile != nil {
		outputFile.Close()
	}
}

func countOpenFiles() int {
	out, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("lsof -p %v", os.Getpid())).Output()
	if err != nil {
		log.Fatal(err)
	}
	lines := strings.Split(string(out), "\n")
	return len(lines) - 1
}

func buildMetadataHeader(bookMeta *metadata) string {
	var sb strings.Builder
	//Header in format [ Author: ; Title: ; Genre: ; ]
	sb.WriteString("[ ")
	sb.WriteString("Author: " + bookMeta.author + "; ")
	sb.WriteString("Title: " + bookMeta.title + "; ")
	sb.WriteString("Categories: " + strings.Join(bookMeta.categories, ", ") + "; ")
	//Language
	sb.WriteString("Language: " + bookMeta.language + "; ")

	sb.WriteString("]\n")
	return sb.String()
}

func basicCleanString(input string) string {
	input = strings.ReplaceAll(input, "	", "")
	input = strings.ReplaceAll(input, "  ", " ")
	input = strings.ReplaceAll(input, "\r", "\n")
	input = strings.ReplaceAll(input, "\n\n\n", "\n")
	input = strings.ReplaceAll(input, "\n\n", "\n")
	input = strings.ReplaceAll(input, "\n\n", "\n")
	input = strings.TrimFunc(input, func(r rune) bool {
		return !unicode.IsGraphic(r)
	})

	input = strings.ReplaceAll(input, " ", " ")

	//replace left-right quotes with normal quotes
	input = strings.ReplaceAll(input, "“", "\"")
	input = strings.ReplaceAll(input, "”", "\"")
	input = strings.ReplaceAll(input, "‘", "'")
	input = strings.ReplaceAll(input, "’", "'")
	return input
}

func gutenBergLineSubstitution(input string, config programConfig) []string {
	//TRIM for gutenburg
	if !config.gutenbergCleaning {
		return []string{input}
	}

	lines := strings.Split(input, "\n")
	lineCount := len(lines)
	if lineCount == 0 {
		fmt.Printf("No lines to clean\n")
		return lines
	}

	if config.gutenbergCleaning {
		//remove before Introduction and after Footnotes
		if lineCount > 0 {
			// Attempt to catch more chapters since epub is rarely 100% accurate
			for _, line := range lines {
				if strings.Contains(line, "Chapter") || strings.Contains(line, "CHAPTER") && line != "CHAPTER_SEPERATOR" {
					line = "CHAPTER_SEPERATOR"
				}
			}

			for i, line := range lines {
				if strings.Contains(line, "\"Cover\"") && i < 10 {
					lines = markLinesBeforeForDeletion(lines, i+1)
					break
				}
			}
			for i, line := range lines {
				if (strings.Contains(line, "Introduction") || strings.Contains(line, "Introduction.") || strings.Contains(line, "INTRODUCTION")) && i < 100 {
					lines = markLinesBeforeForDeletion(lines, i+1)
					break
				}
			}
			for i, line := range lines {
				if (strings.Contains(line, "Introduction") || strings.Contains(line, "Introduction.")) && i < 100 {
					lines = markLinesBeforeForDeletion(lines, i+1)
					break
				}
			}

			for i, line := range lines {
				if strings.Contains(line, "Bibliography") || strings.Contains(line, "BIBLIOGRAPHY.") && i < 100 {
					lines = markLinesBeforeForDeletion(lines, i+1)
					break
				}
			}

			for i, line := range lines {
				if (strings.Contains(line, "Part One") || strings.Contains(line, "PART ONE")) && i < 50 {
					lines = markLinesBeforeForDeletion(lines, i+2)
					break
				}
			}

			for i, line := range lines {
				if (strings.Contains(line, "Contents") || strings.Contains(line, "CONTENTS")) && i < 50 {
					lines = markLinesBeforeForDeletion(lines, i+3)
					break
				}
			}

			for i, line := range lines {
				if strings.Contains(line, "PREFACE") && i < 200 {
					lines = markLinesBeforeForDeletion(lines, i)
					break
				}
			}

			for i, line := range lines {
				if (strings.Contains(line, "START OF THE PROJECT GUTENBERG EBOOK") || strings.Contains(line, "The Project Gutenberg EBook")) && i < 150 {
					lines = markLinesBeforeForDeletion(lines, i+1)
					break
				}
			}

			// Some books have endings at the start strangely
			for i, line := range lines {
				linePercent := float64(i) / float64(lineCount)
				if strings.Contains(line, "Footnotes") && linePercent > 0.8 {
					lines = markLinesAfterForDeletion(lines, i)
					break
				}
			}

			for i, line := range lines {
				linePercent := float64(i) / float64(lineCount)
				if strings.Contains(line, "END OF THE PROJECT GUTENBERG EBOOK") && linePercent > 0.3 {
					lines = markLinesAfterForDeletion(lines, i)
					break
				}
			}

			for i, line := range lines {
				if strings.Contains(line, "APPENDIX") {
					lines = markLinesAfterForDeletion(lines, i)
					break
				}
			}

			//remove any line that has [Pages or [Page, just remove that line
			for i, line := range lines {
				if strings.Contains(line, "[Pages") || strings.Contains(line, "[Page") || strings.Contains(line, "[pg") {
					lines = markLineForDeletion(lines, i)
				}
			}

			//remove any line that has Gutenberg
			for i, line := range lines {
				if strings.Contains(line, "Gutenberg") {
					lines = markLineForDeletion(lines, i)
				}
			}
		}
	} else {
		return lines
	}
	return lines
}

func resolveAllMarks(lines []string) []string {
	for i, line := range lines {
		if line == "MARKED_FOR_DELETION" {
			lines[i] = ""
		}
	}
	return lines[:len(lines)-2]
}

func cleanLineList(lines []string) []string {
	//truncate leading spaces in each line
	for line := range lines {
		lines[line] = strings.TrimLeft(lines[line], " ")
	}
	//remove empty lines (lines that are just spaces or newlines)
	CleanedLines := []string{}
	for _, line := range lines {
		trimmed := strings.Trim(strings.Trim(line, " "), "\n\n")
		if trimmed != "" {
			//fmt.Printf("trimmed: %s\n", trimmed)
			CleanedLines = append(CleanedLines, line)
		}
	}
	return CleanedLines
}

func RemoveToCAndResolveChapterSeperators(lines []string, thresholdRemove int, rangeRemove int) string {
	offset := 0
	//if book does not have chapter seperator in first 10 lines, add it
	for i, line := range lines {
		if strings.Contains(line, "CHAPTER_SEPERATOR") && i < 10 {
			break
		}
		if i > 10 {
			lines = append([]string{"CHAPTER_SEPERATOR\n"}, lines...)
			offset = 1
			break
		}
	}

	// Split story by CHAPTER_SEPERATOR and count number of numbers in each chapter
	// If there are more than 10 numbers in a chapter, assume it is a chapter list and remove it
	storyBuffer := strings.Join(lines, "\n")
	chapterCount := 1
	bookByChapters := strings.Split(storyBuffer, "CHAPTER_SEPERATOR")
	totalChapterCount := len(bookByChapters)
	story := ""

	for _, chapter := range bookByChapters {
		//only check first thresholdRemove and last thresholdRemove chapters
		chapterTitle := ""
		if chapterCount > thresholdRemove && chapterCount < totalChapterCount-thresholdRemove {
			chapterCount++
			if len(chapter) > thresholdRemove {
				if strings.Count(chapter, "HEADER!") > 2 {
					headerSplit := strings.Split(chapter, "HEADER!")

					//try to get chapter title from header
					if len(headerSplit) > 1 {
						chapterTitle = strings.Split(headerSplit[1], "\n")[0]
						if chapterTitle == "" {
							chapterTitle = strings.Split(headerSplit[2], "\n")[0]
						}
						chapterTitle = strings.Trim(chapterTitle, " ")
					}

					//remove header tags and newlines
					chapter = strings.ReplaceAll(chapter, "HEADER!", "\n")
					chapter = strings.ReplaceAll(chapter, "\n ", "\n")
					chapter = strings.ReplaceAll(chapter, "\n\n", "\n")
				}

				//write chapter to story including chapter seperator
				story += fmt.Sprintf("\n***\n[ Chapter %d: %s ; ]\n", chapterCount-offset, chapterTitle)
				story += chapter
			}
			continue
		}

		//attempt to count numbers in chapter to see if it is a chapter list
		numbers := 0
		chapter_nopunct := strings.ReplaceAll(chapter, ".", "")
		chapter_nopunct = strings.ReplaceAll(chapter_nopunct, ",", "")
		chapter_nopunct = strings.ReplaceAll(chapter_nopunct, "-", "")
		words := strings.Split(chapter_nopunct, " ")

		for _, word := range words {
			if _, err := strconv.Atoi(word); err == nil {
				numbers++
			}
		}
		//add chapter to story if it is not a chapter list
		if numbers > rangeRemove && len(chapter) < 10000 {
			offset++
		} else if len(chapter) > 10 {
			//grab title if it exists
			if strings.Count(chapter, "HEADER!") > 2 {
				//check if there is more than 2 HEADER! in chapter
				headerSplit := strings.Split(chapter, "HEADER!")
				if len(headerSplit) > 1 {
					chapterTitle = strings.Split(headerSplit[1], "\n")[0]
					if chapterTitle == "" {
						chapterTitle = strings.Split(headerSplit[2], "\n")[0]
					}
					chapterTitle = strings.Trim(chapterTitle, " ")
				}
				chapter = strings.ReplaceAll(chapter, "HEADER!", "\n")
				chapter = strings.ReplaceAll(chapter, "\n ", "\n")
				chapter = strings.ReplaceAll(chapter, "\n\n", "\n")
			}
			story += fmt.Sprintf("\n***\n[ Chapter %d: %s ; ]\n", chapterCount-offset, chapterTitle)
			story += chapter

		}

		chapterCount++
	}
	return story
}

func cleanEpubString(input string, config programConfig) string {
	//first pass to make it a bit more readable
	input = basicCleanString(input)
	if !config.cleanOutput {
		input = strings.Replace(input, "PARAGRAPH", "\n", -1)
		input = strings.Replace(input, "HEADER!", "\n", -1)
		input = strings.Replace(input, "CHAPTER_SEPERATOR", "\n", -1)
		return input
	}

	//special gutenburg cleaning if enabled (trimming based on common gutenburg headers and footers)
	CleanedLines := []string{}
	if config.gutenbergCleaning {
		CleanedLines = gutenBergLineSubstitution(input, config)
		CleanedLines = resolveAllMarks(CleanedLines)
	} else {
		CleanedLines = cleanLineList(strings.Split(input, "\n"))
	}

	//resolve the paragraph and header marks from <p> tags

	storyBuffer := strings.Join(CleanedLines, " ")
	storyBuffer = strings.Replace(storyBuffer, "PARAGRAPH", "\n", -1)

	CleanedLines = cleanLineList(strings.Split(storyBuffer, "\n"))
	//CleanedLines = strings.Split(storyBuffer, "\n")
	story := RemoveToCAndResolveChapterSeperators(CleanedLines, 20, 15)
	story = strings.Replace(story, "HEADER!", "", -1)
	return story
}

func markLineForDeletion(s []string, index int) []string {
	//replace line with MARKED_FOR_DELETION
	s[index] = "MARKED_FOR_DELETION"
	return s
}

func markLinesBeforeForDeletion(s []string, index int) []string {
	//replace line with MARKED_FOR_DELETION
	for i := 0; i < index; i++ {
		s[i] = "MARKED_FOR_DELETION"
	}
	return s
}

func markLinesAfterForDeletion(s []string, index int) []string {
	//replace line with MARKED_FOR_DELETION
	for i := index; i < len(s); i++ {
		s[i] = "MARKED_FOR_DELETION"
	}
	return s
}

func RemoveIndex(s []string, index int) []string {
	ret := make([]string, 0)
	ret = append(ret, s[:index]...)
	return append(ret, s[index+1:]...)
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
	//case atom.Img:
	//	// Display alt text in place of images.
	//	for _, a := range token.Attr {
	//		switch atom.Lookup([]byte(a.Key)) {
	//		//case atom.Alt:
	//		//text := fmt.Sprintf("Alt text: %s", a.Val)
	//		//p.doc.appendText(text)
	//		//p.doc.row++
	//		//p.doc.col = p.doc.lmargin
	//		//we dont care about alt text, and we dont want to display it
	//		case atom.Src:
	//			for _, item := range p.items {
	//				if item.HREF == a.Val {
	//
	//					break
	//				}
	//			}
	//		}
	//	}
	case atom.Br:
		p.doc.row++
		p.doc.col = p.doc.lmargin
	case atom.H1, atom.H2, atom.H3, atom.H4, atom.H5, atom.H6, atom.Title,
		atom.Div, atom.Tr:
		p.doc.row += 2
		p.doc.col = p.doc.lmargin
		p.sb.WriteString("HEADER!")

	case atom.P:
		p.doc.row += 2
		p.doc.col = p.doc.lmargin
		p.doc.col += 2
		p.sb.WriteString("PARAGRAPH")
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
