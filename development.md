# Amalgam8 Developer Instructions

*Note: These instructions are currently only available for Amalgam8 on local Kubernetes.*

The easiest way to set up an environment where you can compile and experiment with Amalgam8 source code
is by using the same vagrant VM that is used for the kick-the-tires demos in
the [examples](https://github.com/amalgam8/examples) project, only in this case you need to pull more
git repos before doing the "vagrant up".
Alternatively you can set up the required prereqs, described in the 
[Vagrantfile](https://github.com/amalgam8/examples/blob/master/Vagrantfile), yourself,
and then run the samples on your own machine of choice.

To get started using the provided Vagrantfile, proceed as follows:

```bash
git clone git@github.com:amalgam8/examples.git
git clone git@github.com:amalgam8/registry.git
git clone git@github.com:amalgam8/controller.git
git clone git@github.com:amalgam8/sidecar.git

cd examples
vagrant up
vagrant ssh
```

In this environment, you can run all the same samples and demos described in https://github.com/amalgam8/examples/blob/master/README.md,
only now you have the ability to also compile the code and build the images locally.

For example, you can compile and build the control plane and sidecare images locally, with the following command:

```
kubernetes/run-controlplane-local.sh compile
```

If you change the base sidecare image and rebuild it, make sure to also locally build the sample images,
before trying to run them:

```
apps/helloworld/build.sh
apps/bookinfo/build-services-baseimages.sh
```

To remove locally compiled amalgam8 images and just use the ones from Docker Hub:

```
docker rmi $(docker images | grep "amalgam8/" | awk "{print \$3}")
```
