require 'rake/testtask'

GEM_SPEC = eval(File.read('alexandria.gemspec'))

Rake::TestTask.new(:test) do |t|
  t.libs << 'lib' << 'test'
  t.pattern = 'test/**/*_test.rb'
  t.verbose = true
end

# require 'rspec/core/rake_task'
# RSpec::Core::RakeTask.new(:test)

task :man do
  ENV['RONN_ORGANIZATION'] = "#{GEM_SPEC.name} #{GEM_SPEC.version}"
  sh "ronn -5r -stoc man/*.ronn"
end

task :default => :test

