# coding: UTF-8
require 'highline/import'

module Alexandria

  class Library

    EXTENSIONS = %w(epub mobi)

    attr_reader :path

    def initialize(path)
      @path = Pathname.new(path)
    end

    def criteria_matcher(criteria)
      lambda {|book|
        criteria.all? {|k,v|
          if v.is_a?(Regexp)
            book.send(k) =~ v
          else
            book.send(k) == v
          end
        }
      }
    end
    private :criteria_matcher

    def find_all(criteria={})
      block = block_given? ? Proc.new : criteria_matcher(criteria)

      books.find_all {|book| block.call(book) }
    end

    def find(criteria={})
      block = block_given? ? Proc.new : criteria_matcher(criteria)

      books.find {|book| block.call(book) }
    end

    def include?(object)
      case object
      when Book, Book::EmptyBook
        ! find_all(author: object.author, title: object.title).empty?
      when Hash
        ! find_all(object).empty?
      else
        book_path = File.expand_path(object)
        temp = Book.create(book_path)

        include? Book.create(book_path)
      end
    end

    def add!(book)
      dir  = File.join(@path, book.author, book.title)
      dest = File.join(dir, (book.title + book.extension))

      # If it exists, ask to replace
      # if dest.exist?
      if File.exist?(dir)
        return unless agree(" Book '#{book.title}' already exists, replace? [y/n] ".red)
      end

      puts " moving".grey + " #{book.path}"
      puts "     to".grey + " #{dest}"

      # Create dirs, and move
      FileUtils.mkdir_p dir
      FileUtils.cp book.path, dest

      convert dest
    end

    def convert(book)
      ext = File.extname(book)[1..-1]
      needs = EXTENSIONS - [ext]
      base = book.to_s[0..-ext.size-1]

      needs.each do |new_ext|
        new_book = base + new_ext

        puts "Converting to #{new_ext}".blue.bold

        if File.exist?(new_book)
          return unless agree(" Book already converted, convert again? [y/n] ".red)
        end

        if system "#{EBOOK_CONVERT} \"#{book}\" \"#{new_book}\""
          puts " created".grey + " #{new_book}"
        else
          puts " problem converting".red
        end
      end
    end

    def authors
      books.group_by(&:author)
    end

    def books
      Pathname.glob(@path + '*' + "*").map {|dir| Book.new dir }
    end
  end
end
