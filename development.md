# Amalgam8 Developer Instructions

The easiest way to set up an environment where you can compile and experiment with Amalgam8 source code
is by using the same vagrant VM that is used for the sample microservice demos in
the [examples](https://github.com/amalgam8/examples) project. However, in this case you must
checkout the master branch and pull more git repos before doing the "vagrant up".

Alternatively you can set up the required prerequisites that are described in the 
[Vagrantfile](https://github.com/amalgam8/examples/blob/master/Vagrantfile) by yourself,
and then run the samples on your own machine of choice.

To get started using the provided Vagrant file, run the following commands:

```bash
git clone https://github.com/amalgam8/examples.git
git clone https://github.com/amalgam8/registry.git
git clone https://github.com/amalgam8/controller.git
git clone https://github.com/amalgam8/sidecar.git
git clone https://github.com/amalgam8/a8ctl.git
git clone https://github.com/amalgam8/testing.git

cd examples; git checkout master
cd ../registry; git checkout master
cd ../controller; git checkout master
cd ../sidecar; git checkout master
cd ../a8ctl; git checkout master
cd ../testing; git checkout master
cd ..

cd examples

vagrant up
vagrant ssh

cd $GOPATH/src/github.com/amalgam8/examples
```

In this environment, you can run all the same samples and demos that are described in https://github.com/amalgam8/examples/blob/master/README.md,
and you have the ability to also compile the code and build the images locally.

If you need to use the latest (potentially unreleased) version of the Amalgam8 CLI, run the following commands:

```
sudo pip uninstall a8ctl
pushd ../a8ctl
sudo python setup.py develop
popd
```

You can compile and build the control plane or sidecar images locally, by running one of the following commands:

```
../testing/build-scripts/build-controller.sh
../testing/build-scripts/build-registry.sh
../testing/build-scripts/build-sidecar.sh
```

If you change the base sidecar image and rebuild it, ensure that you also locally build the sample images before trying to run them:

```
../testing/build-scripts/build-examples.sh
```

If want to build all three projects and the samples, you can run the following command instead:

```
../testing/build-scripts/build-amalgam8.sh
```

To remove locally-compiled Amalgam8 images and use the Amalgam8 images from Docker Hub:

```
docker rmi $(docker images | grep "amalgam8/" | awk "{print \$3}")
```
