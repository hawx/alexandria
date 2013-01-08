# coding: UTF-8
require 'highline/import'

module Alexandria

  
  # .
  # |- author1
  # |    |- book1
  # |    |- book2
  # |- author2
  #      |- book1    
  
  class Library
  
    EXTENSIONS = %w(epub mobi)
  
    def initialize(path)
      @path = Pathname.new(path)
    end
    
    # Physical
    def add!(book_path)
      # Get title and author
      temp = Book.create(book_path)
      
      dest = @path + temp.author + (temp.title + temp.extension)
      
      # If it exists, ask to replace
      if dest.exist?
        return unless ask(" Book already exists!".red + " Replace?  ")
      end
      
      puts " moving".grey + " #{book_path}"
      puts "     to".grey + " #{dest}"
      
      # Create dirs, and move
      FileUtils.mkdir_p @path + temp.author
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