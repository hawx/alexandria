require 'pathname'

class Alexandria::Device

  class Kindle
    def self.find
      return false unless Dir.exist?('/Volumes/Kindle')
      new '/Volumes/Kindle'
    end

    def initialize(path)
      @path = Pathname.new(path)
    end

    def each_book
      Dir[@path + 'documents' + '**/*.{azw,azw3,mobi}'].each {|path|
        yield Book.create(File.expand_path(path))
      }
    end
  end

  register :kindle, Kindle
end
