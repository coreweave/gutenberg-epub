# fast-epubtotxt
A fast implementation of Golang epub to txt conversion.


![Icon](./iconEpub.png)

---

How to run? The binary is included for the faster golang implementation

Simply call ./convert <inputDir> <outputDir> <writeHeader>

inputDir - string - the path to the input folder, defaults to ./input
outputDir - string - the path to the output folder, defaults to ./output
writeHeader - bool - If you want to write a header with some of the epub's metadata, defaults to true


>./convert -inputDir ./input -outputDir ./output -writeHeader=false

---

Build with golang:

>go build convert.go

Significant References:
https://github.com/soskek/bookcorpus/blob/master/epub2txt.py
https://github.com/taylorskalyo/goreader/tree/master/epub

Taken from a section of code I wrote while working @ Coreweave in Feb 2023
