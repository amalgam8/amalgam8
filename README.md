# Controller
Amalgam8 Controller

## Running
Locally:

    go run ./main.go

For help on command line arguments, run:

    go run ./main.go -h

In a container:

    ./build.sh
    docker run -p 6379:6379 --name controller controller

To use the a8 cli, install the Gremlin SDK in development mode

    git clone github.com/ResilienceTesting/gremlinsdk-python .
    cd gremlinsdk-python/python
    sudo python setup.py develop

Then install and run the a8ctl tools in development mode

    cd cli
    sudo python setup.py develop

To run the gremlin recipes (Note: untested):

    a8ctl recipe-set --topology topology.json --scenario scenario.json
    --checks assertions.json
