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
  
    def initialize(dir)
      @dir = dir
    end
    
    def versions
      Pathname.glob @dir + '*'
    end
    
    def extensions
      versions.map &:extname
    end
    
    def either
      if epub?
        epub
      elsif mobi?
        mobi
      else
        EmptyBook.new
      end
    end
    
    def epub?
      extensions.include? '.epub'
    end
    
    def epub
      @epub ||= Epub.new(
        versions.find {|v| v.extname == '.epub' }
      )
    end
    
    def mobi?
      extensions.include? '.mobi'
    end
    
    def mobi
      @mobi ||= Mobi.new(
        versions.find {|v| v.extname == '.mobi' }
      )
    end
    
    def method_missing(sym, *args, &block)
      either.send sym, *args, &block
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
    
    
    class EmptyBook
      def initialize(path)
        @path = path.to_s
      end
      
      def data
        {}
      end
      
      def metadata
        {}
      end
      
      def chapters
        []
      end
      
      def author
        "None"
      end
      
      def title
        "Untitled"
      end
      
      def extension
        ".missing"
      end
    end
    
    class Epub < EmptyBook
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
      
      def extension
        ".epub"
      end
    end
    
    class Mobi < EmptyBook
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
      
      def extension
        ".mobi"
      end
    end
    
  end
end