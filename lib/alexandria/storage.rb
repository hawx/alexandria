module Alexandria

  module Storage

    class Book
      include ::DataMapper::Resource

      property :id,    Serial
      property :title, String
      property :path,  String

      belongs_to :author

      def self.from_path(path, options={})
        book = ::Alexandria::Book.create(path)

        path = ::Alexandria::Helpers.book_path(book.author, book.title)

        if found = Book.first(:path => path)
          puts "Book already exists!".red
          return found unless options[:force]
        end

        book.write(path)

        author = Author.first_or_create(:name => book.author)
        created = Book.first_or_create(:title => book.title, :author => author, :path => path)
        created.save!

        puts "  created".grey + " #{path}"

        created
      end
    end

    class Author
      include ::DataMapper::Resource

      property :id,   Serial
      property :name, String

      has n, :books

      def books
        Book.all(:author => self, :order => [:title.asc])
      end
    end

    ::DataMapper.finalize
    ::DataMapper.auto_upgrade!
  end
end
