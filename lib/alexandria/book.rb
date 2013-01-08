# coding: UTF-8

require 'peregrin'
require 'mobi'

class Peregrin::Property
  def to_h
    {key => value}
  end
end

module Alexandria
  class Book
  
    def self.create(path)
      case File.extname(path)
        when '.epub' then Epub.new(path)
        when '.mobi' then Mobi.new(path)
      end
    end
  
    def initialize(path)
      @path = path
    end
    
    # def chapters
    # def data
    # def metadata
    # def author
    # def title
    
    def inspect
      "#<#{self.class} '#{self.title}'>"
    end
    alias_method :to_s, :inspect
    
    class Epub < Book
      def data
        @data ||= Peregrin::Epub.read(@path).to_book
      end
      
      def metadata
        @meta ||= data.properties.inject({}) {|a,e| a.merge(e.to_h) }
      end
      
      def chapters
        data.chapters
      end
      
      def author
        metadata['creator']
      end
      
      def title
        metadata['title'].force_encoding("utf-8")
      end
    end
    
    class Mobi < Book
      def data
        # todo...
        
      end
      
      def metadata
        ::Mobi.metadata File.open(@path)
      end
      
      def chapters
        # todo...
      end
      
      def author
        metadata.author
      end
      
      def title
        metadata.title.force_encoding("utf-8")
      end
    end
    
  end
end