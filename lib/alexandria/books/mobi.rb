require 'mobi'

module Alexandria::Book

  class Mobi < Base
    def author
      metadata.author || super
    end

    def title
      metadata.title.force_encoding("utf-8") || super
    end

    def self.extension
      ".mobi"
    end

    private

    def metadata
      ::Mobi.metadata File.open(@path)
    end
  end

  register %w(.mobi .azw .azw3), Mobi
end
