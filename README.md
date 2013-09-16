# Alexandria

Alexandria is a command-line ebook manager.

## Installing

``` bash
$ gem install hawx-alexandria
$ cat >> ~/.bash_rc
export ALEXANDRIA_LIBRARY='~/books' # set to location of library
```

## Adding a book

When a book is added to alexandria it is automatically converted to each
supported format (at the moment `.epub` and `.mobi`). This is done with the
`ebook-convert` tool provided by [Calibre][]. You will need to install
[Calibre][] then set the environment variable `EBOOK_CONVERT` to be the path to
the `ebook-convert` file.

You can the add books like,

``` bash
$ alexandria add /path/to/book.epub
...
$ alexandria add /path/to/another-book.mobi
...
```


## Listing books

You can list all books with

``` bash
$ alexandria list
```


[Calibre]: http://calibre-ebook.com
