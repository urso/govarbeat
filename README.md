# Govarbeat

Welcome to Govarbeat.

Ensure that this folder is at the following location:
`${GOPATH}/github.com/urso`

## Getting Started with Govarbeat

### Init Project
To get running with Govarbeat, run the following commands:

```
glide update --no-recursive
make update
```


To push Govarbeat in the git repository, run the following commands:

```
git init
git add .
git commit
git remote set-url origin https://github.com/ursogovarbeat
git push origin master
```

For further development, check out the [beat developer guide](https://www.elastic.co/guide/en/beats/libbeat/current/new-beat.html).

### Build

To build the binary for Govarbeat run the command below. This will generate a binary
in the same directory with the name govarbeat.

```
make
```


### Run

To run Govarbeat with debugging output enabled, run:

```
./govarbeat -c govarbeat.yml -e -d "*"
```


### Test

To test Govarbeat, run the following commands:

```
make testsuite
```

alternatively:
```
make unit-tests
make system-tests
make integration-tests
make coverage-report
```

The test coverage is reported in the folder `./build/coverage/`


### Update

Each beat has a template for the mapping in elasticsearch and a documentation for the fields
which is automatically generated based on `etc/fields.yml`.
To generate etc/govarbeat.template.json and etc/govarbeat.asciidoc

```
make update
```


### Cleanup

To clean  Govarbeat source code, run the following commands:

```
make fmt
make simplify
```

To clean up the build directory and generated artifacts, run:

```
make clean
```


### Clone

To clone Govarbeat from the git repository, run the following commands:

```
mkdir -p ${GOPATH}/github.com/urso
cd ${GOPATH}/github.com/urso
git clone https://github.com/urso/govarbeat
```


For further development, check out the [beat developer guide](https://www.elastic.co/guide/en/beats/libbeat/current/new-beat.html).
