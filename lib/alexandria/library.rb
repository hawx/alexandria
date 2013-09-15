module Alexandria
  module Library
    extend self

    # Adds a book to the current library, then converts it to all possible formats.
    #
    # @param path [String] Path to the book to add.
    def add(path)
      book = Storage::Book.from_path(File.expand_path(path))

      instance = Dir.glob(book.path + '/*').first

      extensions_missing = Book.extensions - [File.extname(instance)]

      converter = Converter.new(instance)

      extensions_missing.each do |extension|
        converter.convert_to(extension)
      end
    end

    def books(criteria={})
      Storage::Book.all criteria.compact.merge(:order => [:title.asc])
    end

    def authors(criteria={})
      Storage::Author.all criteria.compact.merge(:order => [:name.asc])
    end
  end
end
