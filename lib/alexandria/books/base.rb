require 'fileutils'

module Alexandria::Book

  # @abstract Implement {#author}, {#title} and {.extension}.
  class Base
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

    def self.extension
      ".missing"
    end

    def write(dir)
      name = File.basename(dir)
      write_path = File.join(dir, "#{name}#{self.class.extension}")

      FileUtils.mkdir_p dir
      File.write(write_path, File.read(self.path))
    end

    def inspect
      "#<#{self.class} '#{self.title}'>"
    end
    alias_method :to_s, :inspect
  end
end
