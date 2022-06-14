# yafp: Yet Another Feed Proxy

### feedly_filter:
    author:FEEDLY_LEO_MUTE_ME

### feed_url:
    http://rutor.info/rss.php?full=10
    http://fast-torrent.ru/feeds/rss/

### title_query:
    \((19[0-9][0-9]|20[0-1][0-9]|19[0-9][0-9]-19[0-9][0-9]|19[0-9][0-9]-20[0-1][0-9]|20[0-1][0-9]-20[0-1][0-9])\)

### online_(de/en)coder:
    https://www.urlencoder.org/

### url_format:
    http://localhost:8080/mute?feed_url=URL&title_query=QUERY

    http://localhost:8080/mute?feed_url=URL&title_query=QUERY&rewrite_author=FEEDLY_LEO_MUTE_ME

    http://localhost:8080/mute?feed_url=URL&description_query=QUERY&rewrite_author=FEEDLY_LEO_MUTE_ME

    http://localhost:8080/mute?feed_url=URL&title_query=QUERY&description_query=QUERY&rewrite_author=FEEDLY_LEO_MUTE_ME

### full_url_example:
    http://localhost:8080/mute?feed_url=http%3A%2F%2Ffast-torrent.ru%2Ffeeds%2Frss%2F&title_query=%5C%28%2819%5B0-9%5D%5B0-9%5D%7C20%5B0-1%5D%5B0-9%5D%7C19%5B0-9%5D%5B0-9%5D-19%5B0-9%5D%5B0-9%5D%7C19%5B0-9%5D%5B0-9%5D-20%5B0-1%5D%5B0-9%5D%7C20%5B0-1%5D%5B0-9%5D-20%5B0-1%5D%5B0-9%5D%29%5C%29