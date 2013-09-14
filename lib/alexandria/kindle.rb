require 'pathname'

module Alexandria
  class Kindle
    def self.find
      return false unless Dir.exist?('/Volumes/Kindle')
      new '/Volumes/Kindle'
    end

    def initialize(path)
      @path = Pathname.new(path)
    end

    def books
      Dir[@path + 'documents' + '**/*.{azw,azw3,mobi}'].map {|path|
        Book.create File.expand_path(path)
      }
    end
  end
end
