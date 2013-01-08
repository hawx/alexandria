# coding: UTF-8

module Alexandria
  class Library
  
    EXTENSIONS = %w(epub mobi)
  
    def initialize(path)
      @path = Pathname.new(path)
    end
    
    def authors
      books.group_by(&:author)
    end
    
    def books
      @books ||= Dir[@path + '**' + "*.{#{EXTENSIONS.join(',')}}"]
                  .map {|path| Book.create(path) }
    end
  end
end