require 'json'

similar = Hash.new
File.read("similar.kanji.txt").lines.map { |line| line.strip.split('') }.each do |kanjis|
    kanjis.each do |kanji|
        similar[kanji] = (similar[kanji] || [])
        kanjis.each do |other|
            similar[kanji] << other unless similar[kanji].include?(other) || kanji == other
        end
    end
end

# puts similar.inspect
# exit

subjects = JSON.parse(File.read("../data/subjects.json"))
subjects = subjects['data'].reject { |subject| subject['object'] == 'radical' }

# puts subjects['data'][0].inspect

subjects.each do |subject|
    subject['name'] = subject['data']['characters'] == nil || subject['data']['characters'] == '' ? subject['data']['character'] : subject['data']['characters']
    subject['shortname'] = subject['name'].gsub(/[あいうえおかがきぎくぐけげこごさざしじすずせぜそぞただちぢつづてでとどなにぬねのはばぱひびぴふぶぷへべぺほぼぽまみむめもゃやゅゆょよらりるれろわゐん]/, '')
end

subjectsByName = subjects.group_by { |subject| subject['shortname'] }

def subjectKey(subject)
    "#{subject['object']}/#{subject['name']}"
end

subjects.each do |subject|
    others = subjectsByName[subject['shortname']].reject { |s| subjectKey(s) == subjectKey(subject) }
    subject['shortname'].split('').each_with_index do |k, i|
        # if !similar[k]
        #     if subject['object'] == 'kanji'
        #         html = `curl 'https://www.wanikani.com/kanji/#{subject['name']}' -H 'Accept-Language: en-US,en;q=0.9,de;q=0.8' -H 'Upgrade-Insecure-Requests: 1' -H 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36' -H 'Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8' -H 'Cookie: wanikanitour=ridden; newbie_message_status=false; __utma=7847860.1403867517.1500696500.1505588044.1505604490.4; __utmz=7847860.1500696500.1.1.utmcsr=duckduckgo.com|utmccn=(referral)|utmcmd=referral|utmcct=/; __stripe_mid=93b86bea-3397-47cd-ad2d-7f084c1dd741; undefined_message_status=false; _wanikani_session=3b5e0f57054a577856fe9ee02f53b9ed; _ga=GA1.2.1403867517.1500696500; _gid=GA1.2.145920130.1511311488' -H 'Connection: keep-alive' | grep -E '(<span class="character")|section|h2'`
        #         started = false
        #         puts "Fetched for #{subject['name']}"
        #         html.lines.each do |line|
        #             started = true if line.include?("Visually Similar Kanji")
        #             started = false if line.include?("</section>")
        #             if started && line.include?("<span class=\"character\"")
        #                 puts line
        #                 open('similar.kanji.txt', 'a') { |f|
        #                   f.puts "#{subject['name']}#{line.scan(/<span class="character" lang="ja">(.*)<\/span>/)[0][0]}"
        #                 }
        #             end
        #         end
        #     end
        # end
        next unless similar[k]
        similar[k].each do |s|
            alternate = subject['shortname'][0...i]+s+subject['shortname'][i+1..-1]
            (subjectsByName[alternate] || []).each do |alt|
                unless alt['name'] == subject['name'] || others.include?(alt)
                    #puts "  Adding #{JSON.generate(alt)}"
                    others << alt
                end
            end
        end
    end

    next if others.empty?
    print "#{subjectKey(subject)},"
    print others.map { |s| subjectKey(s) }.join(",")
    puts
end
