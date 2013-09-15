module Alexandria

  module Helpers
    extend self

    def normalise(title)
      title.gsub(' ', '_').gsub(/\W/, '').gsub('_', '-').downcase
    end

    def book_path(author, title)
      normal_author = normalise(author)
      normal_title = normalise(title)
      path = File.join(normal_author, normal_title)

      File.join LIB_ROOT, path
    end
  end
end
