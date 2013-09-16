# -*- encoding: utf-8 -*-
require File.expand_path("../lib/alexandria/version", __FILE__)

Gem::Specification.new do |s|
  s.name         = "hawx-alexandria"
  s.author       = "Joshua Hawxwell"
  s.email        = "m@hawx.me"
  s.summary      = "A library for your ebooks."
  s.homepage     = "http://github.com/hawx/alexandria"
  s.version      = Alexandria::VERSION

  s.description  = <<-DESC
    An ebook library manager, with one-way kindle syncing.
  DESC

  s.add_dependency 'peregrin', '~> 1.2'
  s.add_dependency 'mobi', '~> 0.2.0'
  s.add_dependency 'clive', '~> 1.2'
  s.add_dependency 'highline', '~> 1.6'
  s.add_dependency 'data_mapper', '~> 1.2'
  s.add_dependency 'dm-sqlite-adapter', '~> 1.2'

  s.files        = %w(README.md LICENSE)
  s.files       += Dir["{bin,lib}/**/*"] & `git ls-files`.split("\n")
  s.executables  = %w(alexandria)
end
