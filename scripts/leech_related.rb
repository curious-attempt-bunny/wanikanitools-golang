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
