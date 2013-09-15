module Alexandria

  class Device
    @@registered = {}

    def self.register(name, klass)
      @@registered[name] = klass
    end

    def self.find
      @@registered.map(&:find).compact.first
    end
  end
end
