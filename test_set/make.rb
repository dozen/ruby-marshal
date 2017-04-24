
EXT = ".dat"

test_set = { 
  "null" => nil,
  "int_0" => 0,
  "int_1" => 1,
  "int_-5" => -5,
  "int_777" => 777,
  "int_-777" => -7777,
  "int_65537" => 65537,
  "int_-65537" => -65537,
  "sym_name" => :name,
  "string_hoge" => "hoge",
  "hash_1" => {host: "localhost", db: 1},
  "hash_2" =>  {"name" => "taro", "age" => 21},
  "hash_3" => {user: {name: "matsumoto-yasunori", age: 31}, job: "voice-actor"},
}

test_set.each do |name, val|
  File.open(name + EXT, "w") do |f|
    f << Marshal.dump(val)
  end
end
