# fast-epubtotxt

A fast `epub` to `txt` converter implemented in Golang. This converter acheives speeds of 4-6 million characters/second in testing. For example, it converts the novel _1984_ in ~100-150ms.

## Usage

```bash
./convert \
    -inputDir [INPUT_DIRECTORY] \
    -outputDir [OUTPUT_DIRECTORY] \
    -writeHeader=[true|false] \
    -writeMetadata=[true|false] \
    -cleanOutput=[true|false] \
    -seperateFolders=[true|false] \
    -stopEarly=[INT_NUMBER_OF_BOOKS] \
    -silent=[true|false] \
    -skipCopyRight=[true|false] \
    -gutenbergCleaning=[true|false]

```

Example:

```bash
./convert \
    -inputDir ../data/test-lib \
    -outputDir ./output \
    -writeHeader=true \
    -writeMetadata=true \
    -cleanOutput=true \
    -seperateFolders=true \
    -stopEarly=100 \
    -skipCopyRight=false \
    -gutenbergCleaning=true
```

## Arguments

| Argument | Type | Description | Default Value |
| -------- | ---- | ----------- | ------------- |
| `inputDir` | _string_ | Input folder path | `./input` |
| `outputDir` | _string_ | Output folder path | `./output` | 
| `writeHeader` | _bool_ | Write a metadata header to the `*.txt` file. | `true` |
| `writeMetadata` | _bool_ | Write metadata to a seperate file. | `false` |
| `cleanOutput` | _bool_ | Remove strange characters and spacing from the output. | `true` |
| `gutenbergCleaning` | _bool_ | Perform additional output cleaning for Gutenberg format books. | `false` |
| `seperateFolders` | _bool_ | Write epub and metadata to a seperate folder per book. | `false` |
| `stopEarly` | _int_ | The number of books to process before stopping. | `0` (unlimited) |
| `silent` | _bool_ | Suppress console output. | `false` |
| `skipCopyRight` | _bool_ | Skip all books marked as copyrighted in the metadata. | `false` |

## Build instructions

Build the converter with golang.

```shell
go build convert.go
```

## Official icon

![Icon](./iconEpub.png)

## Parsing benchmark

This converter processed 55,756 books from the Project Gutenberg library in less than 45 minutes.

```
Parsing took 44m56.324860172s, parsed 16465734085 characters at a rate of 6106732 characters per second.
Parsed 55756 books, 55340 finished and 416 skipped due to copy right.
```

## Notes

The converter is not exhaustively tested. Please contact me or raise an issue if errors are discovered.

### Significant references

* https://github.com/soskek/bookcorpus/blob/master/epub2txt.py
* https://github.com/taylorskalyo/goreader/tree/master/epub
* Taken from a section of code I wrote while working as Coreweave in February 2023.
