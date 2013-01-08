# coding: UTF-8

module Alexandria
  class Device
    def initialize(path)
      @path = path
    end
    
    def has?(book)
      false
    end
  end
end