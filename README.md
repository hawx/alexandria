# alexandria

A webapp for managing your ebook library.

First of all you will need the `ebook-convert` tool from [Calibre][] on your
`$PATH`, as this is used for converting between `.mobi` and `.epub`. Then,

``` bash
$ go get alexandria
$ cat > settings.toml
secret = "32 or 64 random bytes, like `head -c64 /dev/urandom | openssl base64`"
database = "./some-db-path"
library = "./some-books-dir"

[uberich]
appName = "alexandria"
appURL = "https://alexandria.example.com"
uberichURL = "https://uberich.example.com"
secret = "shared app secret"
$ mkdir some-books-dir
$ alexandria
...
```


[Calibre]: http://calibre-ebook.com/
