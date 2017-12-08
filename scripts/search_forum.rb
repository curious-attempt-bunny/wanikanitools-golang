10.times do |i|
    page = i+1
    `curl 'https://community.wanikani.com/search?q=%22script%22+%23wanikani%3Aapi-and-third-party-apps&page=#{page}' -H 'accept: application/json' -H 'authority: community.wanikani.com' > data/forum.search.page.#{page}.json`
end