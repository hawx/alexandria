class Hash
  def compact
    self.map {|k,v|
      v.respond_to?(:compact) ? [k, v.compact] : [k,v]
    }.reject {|k,v|
      v.nil? || v.empty?
    }.to_h
  end
end

class Object
  def to_h
    Hash[self]
  end
end
