# yafp: Yet Another Feed Proxy

### feedly_filter:
    author:FEEDLY_LEO_MUTE_ME

### feed_url:
    http://rutor.info/rss.php?full=10

### title_query:
    \((19[0-9][0-9]|20[0-1][0-9]|19[0-9][0-9]-19[0-9][0-9]|19[0-9][0-9]-20[0-1][0-9]|20[0-1][0-9]-20[0-1][0-9])\)

### online_(de/en)coder:
    https://www.urlencoder.org/

### url_format:
    http://localhost:8080/mute?feed_url=URL&title_query=QUERY&rewrite_author=FEEDLY_LEO_MUTE_ME

    http://localhost:8080/mute?feed_url=URL&description_query=QUERY&rewrite_author=FEEDLY_LEO_MUTE_ME

    http://localhost:8080/mute?feed_url=URL&title_query=QUERY&description_query=QUERY&rewrite_author=FEEDLY_LEO_MUTE_ME
