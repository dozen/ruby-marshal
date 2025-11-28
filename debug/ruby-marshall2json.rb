#!/usr/bin/env ruby
# frozen_string_literal: true

require 'json'

def load_and_dump_json(marshalled_file = nil)
  begin
    # Read from file if provided, otherwise read from stdin
    marshalled_data = if marshalled_file
                        File.binread(marshalled_file)
                      else
                        $stdin.read
                      end

    data = Marshal.load(marshalled_data)

    # Dump the data as JSON
    puts JSON.pretty_generate(data)
  rescue => e
    puts "Error loading or dumping data: #{e.message}"
  end
end

# Check for command-line arguments
if ARGV.empty?
  load_and_dump_json # Use stdin
else
  load_and_dump_json(ARGV[0]) # Use the first argument as the file path
end
