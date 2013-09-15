module Alexandria
  module Book
    @@registered = {}

    def self.register(extensions, klass)
      extensions.each do |extension|
        @@registered[extension] = klass
      end
    end

    def self.create(path)
      ext = File.extname(path)

      unless @@registered.has_key?(ext)
        warn "Unrecognised extension #{ext}"
        exit 2
      end

      @@registered[ext].new(path)
    end

    def self.extensions
      @@registered.values.map(&:extension)
    end
  end
end
