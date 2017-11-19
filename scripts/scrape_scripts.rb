require 'json'

scripts = JSON.parse(File.read('data/scripts.json'))['scripts']

scripts.each do |url, script|
    filename = "data/script.#{script['topic_id']}.html"
    puts filename
    html = File.read(filename)

    m = html.match(/<h2>([^<]+)<\/h2>.*?<p id="script-description">([^<]+)<\/p>.*?<dd class="script-show-total-installs"><span>([^<]+)<\/span><\/dd>.*?<dd class="script-show-version"><span>([^<]+)<\/span><\/dd>/m)

    if m
        script['name'] = m[1]
        script['description'] = m[2]
        script['installs'] = m[3].gsub(/,/, '').to_i
        script['version'] = m[4]
    else
        puts "Failed regex! (#{script['script_url']})"
        `rm #{filename}`
    end
end

File.write('data/scripts.json', JSON.generate({scripts: scripts}))
File.write('../static/scripts.json', JSON.generate(scripts.values))