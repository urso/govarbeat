# Govarbeat

Welcome to govarbeat.

To get running with your beat, run the following commands:

```
glide init
glide update --no-recursive
make update
make
```

To run your beat with debugging output enabled, run:

```
./govarbeat -c etc/govarbeat.yml -e -d "*"
```

For further development, check out the [beat developer guide](https://www.elastic.co/guide/en/beats/libbeat/current/new-beat.html).

To start it as a git repository, run

```
git init
```
