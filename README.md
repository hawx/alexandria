# alexandria

A webapp for managing your ebook library.

First of all you will need the `ebook-convert` tool from [Calibre][] on your
`$PATH`, as this is used for converting between `.mobi` and `.epub`. Then,

``` bash
$ go get alexandria
$ cat > settings.toml
user = "m@hawx.me"
^D
$ mkdir alexandria-books
$ alexandria
...
```


[Calibre]: http://calibre-ebook.com/
