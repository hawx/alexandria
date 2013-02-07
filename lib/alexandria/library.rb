# coding: UTF-8
require 'highline/import'

module Alexandria
  
  class Library
  
    EXTENSIONS = %w(epub mobi)
  
    def initialize(path)
      @path = Pathname.new(path)
    end
    
    # Physical
    def add!(book_path)
      book_path = File.expand_path(book_path)
    
      # Get title and author
      temp = Book.create(book_path)
      
      dir  = @path + temp.author + temp.title
      dest = dir + (temp.title + temp.extension)
      
      # If it exists, ask to replace
      if dest.exist?
        return unless agree(" Book already exists, replace? [y/n] ".red)
      end
      
      puts " moving".grey + " #{book_path}"
      puts "     to".grey + " #{dest}"
      
      # Create dirs, and move
      FileUtils.mkdir_p dir
      FileUtils.cp book_path, dest
      
      convert dest
    end
    
    def convert(book)
      ext = book.extname[1..-1]
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
      @books ||= Dir[@path + '*' + "*.{#{EXTENSIONS.join(',')}}"]
                  .map {|path| Book.create(path) }
    end
  end
end