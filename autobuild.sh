#!/bin/bash

cur=`pwd`
kill=1

inotifywait -mqr --timefmt '%d/%m/%y %H:%M' --format '%T %w %f' \
   -e modify ./ | while read date time dir file; do
    ext="${file##*.}"
    if [[ "$ext" = "md" ]]; then
        echo "$file changed @ $time $date, rebuilding..."
        rt -t custom.jinja --prettify index.md
    fi
done

