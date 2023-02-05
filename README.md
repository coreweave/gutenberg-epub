# fast-epubtotxt
A fast implementation of Golang epub to txt conversion.

Speeds of 2.5-5 million characters/second in testing (the novel 1984 in ~100-150ms)

![Icon](./iconEpub.png)


## Usage

``` shell
./convert <inputDir> <outputDir> <writeHeader>

Example
./convert -inputDir ./input -outputDir ./output -writeHeader=false
```


---
## Arguments

Binary included with example folders.

Call ./convert <inputDir> <outputDir> <writeHeader>

inputDir - string - the path to the input folder, defaults to ./input
outputDir - string - the path to the output folder, defaults to ./output
writeHeader - bool - If you want to write a header with some of the epub's metadata, defaults to true


---
## Build with golang
```shell
go build convert.go
```

Significant References:
https://github.com/soskek/bookcorpus/blob/master/epub2txt.py
https://github.com/taylorskalyo/goreader/tree/master/epub

Taken from a section of code I wrote while working @ Coreweave in Feb 2023
