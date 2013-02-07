# coding: UTF-8

require 'pathname'

require 'clive/output'

EBOOK_CONVERT = ENV['EBOOK_CONVERT']
# Fall back to default mac location. I know, mac.
EBOOK_CONVERT ||= '/Applications/calibre.app/Contents/MacOS/ebook-convert'

require_relative 'alexandria/book'
require_relative 'alexandria/device'
require_relative 'alexandria/library'