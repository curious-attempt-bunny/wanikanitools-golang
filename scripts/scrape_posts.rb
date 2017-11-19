require 'json'
require 'cgi'

brokenLinks = [
    'https://anonmgur.com/up/690971a092473f53f6784a155cf46f1a.png',
    'https://s3.amazonaws.com/s3.wanikani.com/assets/v03/loading-100x100.gif',
    'https://s3.amazonaws.com/s3.wanikani.com/assets/v03/loading-100x100.gif'
]

scriptToForum = Hash.new

Dir.glob('data/topic.*.json').each do |file|
    topic = JSON.parse(File.read(file))
    post = topic['post_stream']['posts'][0]
    m = post['cooked'].match(/href="(https:\/\/greasyfork.org\/(?:[^\/]+\/)scripts\/[^\/"?]+)/m)
    if m
        script_url = m[1]
        
        topic_url = "https://community.wanikani.com/t/#{topic['slug']}/#{topic['id']}"
        next if topic_url == "https://community.wanikani.com/t/the-new-and-improved-list-of-api-and-third-party-apps/7694"
        puts topic_url

        img_urls = post['cooked'].scan(/<img[^>]+src="([^"]+)"/m).flatten.map { |url| CGI::unescapeHTML(url).gsub('</em>', '_') }
        puts img_urls.inspect
        img_url = img_urls.find do |url|
            if url.include?('/emoji/')
                false 
            elsif brokenLinks.include?(url)
                false
            else
                true
            end
        end
        puts "  -> #{img_url}"

        likes = post['actions_summary'].find { |action| action['count'] }
        likes = likes ? likes['count'] : 0

        # print "#{'%03d' % likes}♥, #{topic_url}, "
        # puts "#{script_url}, #{img_url}, #{post['username']}"

        entry = {img_url: img_url, likes: likes, author: post['username'], topic_url: topic_url, script_url: script_url, topic_id: topic['id'] }
        # puts "ALREADY LISTED:\n  #{JSON.generate(scriptToForum[script_url])}\n  #{JSON.generate(entry)}" if scriptToForum[script_url]
        next if scriptToForum[script_url]
        scriptToForum[script_url] = entry
    end
end

File.write('data/scripts.json', JSON.generate({scripts: scriptToForum}))