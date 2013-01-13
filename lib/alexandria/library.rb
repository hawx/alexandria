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
        return unless ask(" Book already exists!".red + " Replace? [y/n] ")
      end
      
      puts " moving".grey + " #{book_path}"
      puts "     to".grey + " #{dest}"
      
      # Create dirs, and move
      FileUtils.mkdir_p dir
      FileUtils.cp book_path, dest
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