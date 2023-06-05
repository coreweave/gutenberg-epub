# fast-epubtotxt
A fast implementation of Golang epub to txt conversion.

Speeds of 4-6 million characters/second in testing (the novel 1984 in ~100-150ms)


![Icon](./iconEpub.png)


## Usage

``` shell
./convert -inputDir (dir) -outputDir (dir) -writeHeader=(bool) -writeMetadata=(bool) -cleanOutput=(bool) -seperateFolders=(bool) -stopEarly=(int) -silent=(bool) -skipCopyRight=(bool) -gutenbergCleaning=(bool)

Example
./convert -inputDir ../data/test-lib -outputDir ./output -writeHeader=true -writeMetadata=true -cleanOutput=true -seperateFolders=true -stopEarly=100 -skipCopyRight=false -gutenbergCleaning=true
```


---
## Arguments

note: not exhaustively tested, please contact me or raise an issue if code doesn't work for a parameter combination

Binary included with example folders.

```shell
./convert -inputDir (dir) -outputDir (dir) -writeHeader=(bool) -writeMetadata=(bool) -cleanOutput=(bool) -seperateFolders=(bool) -stopEarly=(int) -silent=(bool) -skipCopyRight=(bool) -gutenbergCleaning=(bool)
```

inputDir - string - the path to the input folder, defaults to ./input

outputDir - string - the path to the output folder, defaults to ./output

writeHeader - bool - If you want to write a header with some of the epub's metadata, defaults to true.

writeMetadata - bool - If you want to write all metadata to a seperate file, defaults to false.

cleanOutput - bool - If you want to clean the outputted text file for strange characters and spacing, defaults to true.

gutenbergCleaning - bool - Additional output cleaning for gutenberg format books, defaults to false.

seperateFolders - bool - If you want your epub and metadata to be written into a seperate folder per-book, defaults to false.

stopEarly - bool - If you want the epubs to be converted to be limited to that number, defaults to 0 (unlimited).

silent - bool - Silences console output, defaults to false.

skipCopyRight - bool - Skips all books that is marked as copyrighted in the metadata, defaults to false.



---
## Build with golang
```shell
go build convert.go
```

Significant References:
https://github.com/soskek/bookcorpus/blob/master/epub2txt.py

https://github.com/taylorskalyo/goreader/tree/master/epub

Taken from a section of code I wrote while working @ Coreweave in Feb 2023


---

## Parsing benchmark

Gutenberg Library

```
Parsing took 44m56.324860172s, parsed 16465734085 characters at a rate of 6106732 characters per second.
Parsed 55756 books, 55340 finished and 416 skipped due to copy right.
```