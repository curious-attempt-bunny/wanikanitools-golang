require 'json'

# 
        # puts "#{post['username']}, #{m[1]}"

scriptToForum = Hash.new

Dir.glob('data/topic.*.json').each do |file|
    topic = JSON.parse(File.read(file))
    post = topic['post_stream']['posts'][0]
    m = post['cooked'].match(/href="(https:\/\/greasyfork.org\/(?:[^\/]+\/)scripts\/[^\/"?]+)/m)
    if m
        script_url = m[1]
        m = post['cooked'].match(/<img src="([^"]+)"/m)
        img_url = m ? m[1].to_s : nil
        img_url = nil if img_url && img_url.include?('/emoji/')
    
        likes = post['actions_summary'].find { |action| action['count'] }
        likes = likes ? likes['count'] : 0

        topic_url = "https://community.wanikani.com/t/#{topic['slug']}/#{topic['id']}"
        next if topic_url == "https://community.wanikani.com/t/the-new-and-improved-list-of-api-and-third-party-apps/7694"
        # print "#{'%03d' % likes}♥, #{topic_url}, "
        # puts "#{script_url}, #{img_url}, #{post['username']}"

        entry = {img_url: img_url, likes: likes, author: post['username'], topic_url: topic_url, script_url: script_url, topic_id: topic['id'] }
        # puts "ALREADY LISTED:\n  #{JSON.generate(scriptToForum[script_url])}\n  #{JSON.generate(entry)}" if scriptToForum[script_url]
        next if scriptToForum[script_url]
        scriptToForum[script_url] = entry
    end
end

File.write('data/scripts.json', JSON.generate({scripts: scriptToForum}))