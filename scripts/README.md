# Usage

To rebuild static/scripts.json:

    cd scripts
    ruby search_forum.rb && ruby download_posts.rb && ruby scrape_posts.rb && ruby download_scripts.rb && ruby scrape_scripts.rb
    
To rebuild static/similar.csv:

    ruby leech_related.rb > ../static/similar.csv

