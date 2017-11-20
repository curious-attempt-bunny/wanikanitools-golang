require 'json'

topicLastUpdated = {}

4.times do |i| 
    latest = JSON.parse(`curl 'https://community.wanikani.com/c/wanikani/api-and-third-party-apps/l/latest?no_subcategories=false&page=#{i}&_=1511140377129' -H 'Accept: application/json'`)

    latest['topic_list']['topics'].each { |topic| topicLastUpdated[topic['id']] = topic['last_posted_at'] }
end

Dir.glob("data/forum.search.page.*.json").each do |file|
    # puts file
    JSON.parse(File.read(file))['posts'].each do |post|
        filename = "data/topic.#{post['topic_id']}.json"
        if File.exists?(filename)
            topic = JSON.parse(File.read(filename))
            next unless topicLastUpdated[topic['id']] && topicLastUpdated[topic['id']] != topic['last_posted_at']
            puts "#{topic['last_posted_at']} != #{topicLastUpdated[topic['id']]}"
        end 
        cmd = "curl https://community.wanikani.com/t/#{post['topic_id']}.json > #{filename}"
        puts cmd
        `#{cmd}`
        sleep(1)
    end
end

