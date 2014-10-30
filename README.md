# alexandria

A webapp for managing your ebook library.

First of all you will need the `ebook-convert` tool from [Calibre][] on your
`$PATH`, as this is used for converting between `.mobi` and `.epub`. Then,

``` bash
$ go get alexandria
$ cat > settings.toml
users = ["you@domain.com"]
secret = "some 32 or 64 byte string"
audience = "hostname and port of place this will be"
database = "./some-db-path"
library = "./some-dir-for-the-books"
^D
$ mkdir some-dir-for-the-books
$ alexandria
...
```


[Calibre]: http://calibre-ebook.com/
