# procstat
observe cpu and memory usage of a single process and create an easy chart out of it

install:

    $ go get github.com/advincze/procstat


run:

    $ procstat -pid 12345 -html /tmp/chart.html -tick 200ms > /dev/null
