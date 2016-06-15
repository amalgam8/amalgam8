To try out the samples, you should first deploy the gateway. Then deploy
any of the apps in this directory. The gateway serves as the user facing
proxy that can be controlled for various purposes such as active deploy,
resiliency testing, etc.

To try out the apps: 
* Launch gateway (cd ../gateway; kubectl create -f gateway.yaml)
* Launch helloword (cd helloworld; ./run.sh) See helloworld/README.md for more info)
* Launch bookinfo (cd bookinfo; ./run.sh) See bookinfo/README.md for more info)
