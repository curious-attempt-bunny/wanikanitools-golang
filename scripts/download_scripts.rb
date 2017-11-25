require 'json'

JSON.parse(File.read('data/scripts.json'))['scripts'].each do |url, script|
    filename = "data/script.#{script['topic_id']}.html"
    next if File.exists?(filename)

    cmd = "curl -L #{url} > #{filename}"
    puts cmd
    `#{cmd}`
    sleep(1)
end

JSON.parse(File.read('data/scripts.json'))['scripts'].each do |url, script|
    filename = "data/script.#{script['topic_id']}.js"
    next if File.exists?(filename)

    cmd = "curl -L #{url}/code.js > #{filename}"
    puts cmd
    `#{cmd}`
    sleep(1)
end