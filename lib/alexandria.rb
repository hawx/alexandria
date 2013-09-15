# coding: UTF-8

require 'pathname'
require 'clive/output'

EBOOK_CONVERT = ENV['EBOOK_CONVERT'] ||
                '/Applications/calibre.app/Contents/MacOS/ebook-convert'
# Fall back to default mac location. I know, mac.

require_relative 'alexandria/core_ext'
require_relative 'alexandria/helpers'

require_relative 'alexandria/book'
require_relative 'alexandria/books/base'
require_relative 'alexandria/books/epub'
require_relative 'alexandria/books/mobi'

require_relative 'alexandria/device'
require_relative 'alexandria/devices/kindle'

require_relative 'alexandria/converter'
require_relative 'alexandria/library'
require_relative 'alexandria/storage'
