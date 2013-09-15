require 'peregrin'

class Peregrin::Property
  def to_h
    {key => value}
  end
end

module Alexandria::Book

  class Epub < Base
    def author
      metadata['creator'] || super
    end

    def title
      metadata['title'].force_encoding("utf-8") || super
    end

    def self.extension
      ".epub"
    end

    private

    def metadata
      @meta ||= Peregrin::Epub.read(@path)
        .to_book
        .properties
        .inject({}) {|a,e| a.merge(e.to_h) }
    end
  end

  register %w(.epub), Epub
end
