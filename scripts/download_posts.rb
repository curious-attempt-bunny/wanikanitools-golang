require 'json'

Dir.glob("data/forum.search.page.*.json").each do |file|
    # puts file
    JSON.parse(File.read(file))['posts'].each do |post|
        filename = "data/topic.#{post['topic_id']}.json"
        next if File.exists?(filename)
        cmd = "curl https://community.wanikani.com/t/#{post['topic_id']}.json > #{filename}"
        puts cmd
        `#{cmd}`
        sleep(1)
    end
end

# m = post['cooked'].match(/href="(https:\/\/greasyfork.org\/scripts\/[^"]+)"/)
        # puts "#{post['username']}, #{m[1]}"
