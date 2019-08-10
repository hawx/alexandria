# alexandria

A webapp for managing your ebook library.

First of all you will need the `ebook-convert` tool from [Calibre][] on your
`$PATH`, as this is used for converting between `.mobi` and `.epub`. Then,
install with `go get` and see the options with `alexandria --help`. The url set
for `me` should define links for [IndieAuth][].

``` bash
$ go get alexandria
$ mkdir some-books-dir
$ alexandria \
    --db ./some-db-path \
    --books ./some-books-dir \
    --me https://john.example.com/ \
    --secret "$(head -c64 /dev/urandom | openssl base64)"
...
```

[Calibre]:   http://calibre-ebook.com/
[IndieAuth]: https://indieweb.org/IndieAuth
