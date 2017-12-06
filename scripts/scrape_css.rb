require 'set'
require 'json'

lines = `grep -E '#[^0-9]|\\.[^0-9]' data/wanikani.css`.lines

wkids = Set.new
wkclasses = Set.new

lines.each do |line|
    m = line.scan(/(#[a-zA-Z_\-][a-zA-Z_\-0-9]*)/)

    # puts line
    # puts m.inspect
    m.flatten.each do |id|
        wkids.add(id) unless id.match(/#[0-9a-fA-F]{3}[0-9a-fA-F]{3}?/)
    end

    m = line.scan(/(\.[a-zA-Z_\-][a-zA-Z_\-0-9]*)/)
    # puts m.inspect
    m.flatten.each do |clazz|
        wkclasses.add(clazz)
    end
end

css = `grep -E 'id\s*=\s*['"'"'"]|class\s*=\s*['"'"'"]' data/script*.js`.lines

topicToScript = Hash.new

css.each do |line|
    filename = line[0...line.index(':')]
    code = line[line.index(':')+1..-1]
    topic_id = filename.match(/^data\/script\.([0-9]+)\.js$/)[1].to_i
    topicToScript[topic_id] = {ids: [], classes: []} unless topicToScript[topic_id]

    m = code.scan(/id\s*=\s*['"]([a-zA-Z_\-][a-zA-Z_\-0-9]*)['"]/)
    # puts m.flatten.inspect
    topicToScript[topic_id][:ids] = (topicToScript[topic_id][:ids] + m.flatten).uniq

    m = code.scan(/class(?:[Nn]ame)?\s*=\s*['"]([a-zA-Z_\-][a-zA-Z_\-0-9 ]*)['"]/)
    # puts m.flatten.map { |cn| cn.split(' ') }.flatten.inspect
    topicToScript[topic_id][:classes] = (topicToScript[topic_id][:classes] + m.flatten.map { |cn| cn.split(' ') }.flatten).uniq
end

ids = []
classes = []

topicToScript.values.each do |script|
    ids += script[:ids]
    classes += script[:classes]
end

sharedIds = ids.group_by { |i| i }.select { |k,v| v.size >= 2 }
sharedClasses = classes.group_by { |i| i }.select { |k,v| v.size >= 2 }

puts JSON.pretty_generate(sharedIds.keys)
puts JSON.pretty_generate(sharedClasses.keys)