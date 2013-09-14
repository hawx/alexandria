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
      ext = File.extname(path)

      if Mobi::EXTENSIONS.include?(ext)
        Mobi.new(path)

      elsif Epub::EXTENSIONS.include?(ext)
        Epub.new(path)

      else
        warn "Unrecognised file of type: #{File.extname(path)}"
        exit
      end
    end

    def initialize(dir)
      @dir = dir
    end

    def path
      @dir
    end

    def versions
      Pathname.glob @dir + '*'
    end

    def extensions
      versions.map &:extname
    end

    def either
      found = self.class.registered.find { |type|
        send("#{type}?")
      }

      if found
        send(found)
      else
        EmptyBook.new('null')
      end
    end

    def self.registered
      @__registered
    end

    def self.register(name, klass)
      ivar = "@__#{name}".to_sym

      define_method name do
        book = self.instance_variable_get(ivar)
        return book if book

        self.instance_variable_set ivar, klass.new(
          versions.find {|v| klass::EXTENSIONS.include? v.extname }
        )
      end

      define_method "#{name}?".to_sym do
        klass::EXTENSIONS.any? {|ext|
          extensions.include?(ext)
        }
      end

      registered = self.instance_variable_get(:@__registered) || []
      self.instance_variable_set(:@__registered, registered << name)
    end

    extend Forwardable
    def_delegators :either, :author, :title

    def inspect
      "#<#{self.class} '#{self.title}'>"
    end
    alias_method :to_s, :inspect

    # @abstract Implement {#author}, {#title} and {#extension} and define a list
    #   of possible {EXTENSIONS}.
    class EmptyBook
      EXTENSIONS = []

      attr_reader :path

      def initialize(path)
        @path = path.to_s
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

      def inspect
        "#<#{self.class} '#{self.title}'>"
      end
      alias_method :to_s, :inspect
    end

    class Epub < EmptyBook
      EXTENSIONS = ['.epub']

      def metadata
        @meta ||= Peregrin::Epub.read(@path)
                                .to_book
                                .properties
                                .inject({}) {|a,e| a.merge(e.to_h) }
      end

      def author
        metadata['creator'] || super
      end

      def title
        metadata['title'].force_encoding("utf-8") || super
      end

      def extension
        ".epub"
      end
    end

    register :epub, Epub

    class Mobi < EmptyBook
      EXTENSIONS = ['.mobi', '.azw', '.azw3']

      def metadata
        ::Mobi.metadata File.open(@path)
      end

      def author
        metadata.author || super
      end

      def title
        metadata.title.force_encoding("utf-8") || super
      end

      def extension
        ".mobi"
      end
    end

    register :mobi, Mobi

  end
end
