# coding: UTF-8

require 'pathname'

require 'clive/output'

EBOOK_CONVERT = ENV['EBOOK_CONVERT'] || 
                '/Applications/calibre.app/Contents/MacOS/ebook-convert'
# Fall back to default mac location. I know, mac.

require_relative 'alexandria/book'
require_relative 'alexandria/library'