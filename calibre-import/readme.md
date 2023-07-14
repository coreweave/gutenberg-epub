# Import Project Gutenberg books to Calibre

## Background

[Project Gutenberg](https://www.gutenberg.org/) offers a large collection of public-domain books, but it's difficult to download books from a specific author or genre without crawling their site, which is strongly discouraged. An easier solution is to download the entire Project Gutenberg ZIM archive from the Kiwix project, then process it with OpenZIM and Calibre.

The [Kiwix project](https://wiki.kiwix.org/wiki/Content) is an open-source initiative that makes online content available for offline use. Kiwix allows users to download entire websites, including Project Gutenberg, in a compressed file format known as ZIM. Project Gutenberg's ZIM archive is available via [direct download](https://download.kiwix.org/zim/gutenberg/), [BitTorrent](https://download.kiwix.org/zim/gutenberg_en_all.zim.torrent), or a [magnet link](https://download.kiwix.org/zim/gutenberg_en_all.zim.magnet). Kiwix also maintains [OpenZIM](https://wiki.openzim.org/wiki/OpenZIM), a set of open-source tools to used manipulate ZIM archives.

[Calibre](https://calibre-ebook.com/about) is a powerful open-source e-book management app for viewing, organizing, and converting e-books in various formats. Importing the entire Project Gutenberg catalog is time-consuming, but once indexed, a typical personal computer can easily manage the entire library.

## Process overview

The main steps:

1. Install `zim-tools` on a new Ubuntu server.
1. Download the Project Gutenberg ZIM archive.
1. Extract the EPUBs with `zimdump`.
1. Install Calibre.
1. Import the books into Calibre.

This process is time-consuming. Intermediate files are available for download from a filebrowser app hosted in CoreWeave prod. 

These aren't secret, it's public information from Project Gutenberg.

* Login URL: <https://gutenberg-dataset.tenant-96362f-dev.ord1.ingress.coreweave.cloud/files/>
* User UI: `gutenberg`
* Password: `WGE0vwm8dhe.hew!pqh`

| Filename | Description |
| --- |  --- |
| gutenberg_en_all_2023-05.zim | CoreWeave mirror of the full ZIM file ([original source](https://download.kiwix.org/zim/gutenberg/))
| pg-raw.tar.gz | EPUB files extracted from the May 2023 Project Gutenberg ZIM |
| pg-calibre-library.tar.gz | The Project Gutenberg catalog in a Calibre library |

Each file has a corresponding `*.sha256` for verification.

## Step 1: Install zim-tools

On a freshly-installed Ubuntu 22.04 server, install the prerequisite packages.

```bash
$ sudo apt-get install liblzma-dev \
    libicu-dev \
    libzstd-dev \
    libxapian-dev \
    meson \
    libdocopt-dev \
    libkainjow-mustache-dev \
    libmagic-dev \
    zlib1g-dev \
    libgumbo-dev \
    libicu-dev \
    cmake
```

Create a project directory and clone the [libzim repo](https://github.com/openzim/libzim).

```bash
mkdir /tmp/pg-files/OpenZIM
cd /tmp/pg-files/OpenZIM
git clone https://github.com/openzim/libzim.git
cd libzim
```

Compile and install `libzim`.

```bash
meson . build
ninja -C build
sudo ninja -C build install
```

Clone the [zim-tools repo](https://github.com/openzim/zim-tools).

```bash

cd /tmp/pg-files/OpenZIM
git clone https://github.com/openzim/zim-tools.git
cd zim-tools
```

Compile and install `zim-tools`.

```bash
meson . build
ninja -C build
sudo ninja -C build install
```

Test `zimdump`.

```bash
$ zimdump --version
zim-tools 3.2.0

libzim 8.2.0
+ libzstd 1.4.8
+ liblzma 5.2.5
+ libxapian 1.4.18
+ libicu 70.1.0

```

## Step 2: Download the ZIM archive

Download the ZIM from the file list above. 

Or, the latest ZIM file from Kiwix and verify the [sha-256 sum matches](https://download.kiwix.org/zim/gutenberg/gutenberg_en_all_2023-05.zim.sha256).

```bash
$ mkdir /tmp/pg-files
$ cd /tmp/pg-files

$ curl -O --progress-bar https://download.kiwix.org/zim/gutenberg/gutenberg_en_all_2023-05.zim

$ sha256sum gutenberg_en_all_2023-05.zim
c57133c971c7cf82df907e8fe037e84d7ee2d54ec6bd72af97b6ba509e33d9cf  gutenberg_en_all_2023-05.zim

```

> These files are from May 2023, [check the website](https://wiki.kiwix.org/wiki/Content) for the latest version.

## Step 3: Extract the EPUBs

Dump the ZIM to a `dump` directory.

```bash
mkdir dump
zimdump dump --dir=dump gutenberg_en_all.zim
```

Wait about 20 minutes for it to complete, then check the directory. There are 1,423,653 total files, including 55,756 `*.epub` files. The rest are HTML, images, CSS, and other things we don't need.

```bash
$ ls -l dump | wc -l
1423653

$ ls -l dump | grep epub$ | wc -l
55756
```

Move `*.epub` to a new directory.

```bash
mkdir pg-epub
find dump -name *.epub -type f -print0 | xargs -0 -I {} mv {} pg-epub
```

Verify the number of `*.epub` in `pg-epub` matches the earlier count.

```bash
$ ls -l pg-epub | grep epub$ | wc -l
55756
```

Delete the `dump` directory.

```bash
rm -rf dump
```

## Step 4: Install Calibre

Install [Calibre](https://calibre-ebook.com/download_linux) according to the instructions at [calibre-ebook.com](https://calibre-ebook.com/download):

```bash
$ sudo apt update
$ sudo -v && wget -nv -O- https://download.calibre-ebook.com/linux-installer.sh | sudo sh /dev/stdin

$ calibredb --version
calibredb (calibre 6.17)
```

## Step 5: Import the books into Calibre

```bash
$ for file in /tmp/pg-files/pg-epub/*
  do
      calibredb add --automerge new_record \
        --library-path /tmp/pg-calibre-library "$file"
  done
```

Output:

```bash
Added book ids: 1
Added book ids: 2
Added book ids: 3
...
# Many hours later...
...
Added book ids: 55754
Added book ids: 55755
Added book ids: 55756
```

Importing the EPUB files to Calibre takes up to 48 hours, depending on the speed of your machine.
