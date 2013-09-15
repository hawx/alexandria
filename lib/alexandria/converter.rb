module Alexandria

  class Converter

    def initialize(original_path, options={})
      @original_path = original_path
      @book = Book.create(@original_path)
    end

    def dir
      File.dirname(@original_path)
    end

    def normalised_title
      Helpers.normalise(@book.title)
    end

    def new_path(ext)
      File.join(dir, normalised_title + ext)
    end

    def converted_to?(ext)
      File.exist? new_path(ext)
    end

    def convert_to(new_ext)
      new_path = new_path(new_ext)

      if converted_to?(new_ext)
        return unless agree("File already exists '#{new_path}', convert again? [y/n]".red)
      end

      if execute(@original_path, new_path, @options[:quiet])
        puts "  created".grey + " #{new_path}"
      else
        puts "  problem".red + " converting #{@original_path} to #{new_ext}"
      end
    end

    private

    def execute(from, to, quiet=true)
      if quiet
        `#{EBOOK_CONVERT} \"#{from}\" \"#{to}\"`
      else
        system "#{EBOOK_CONVERT} \"#{from}\" \"#{to}\""
      end
    end

  end
end
