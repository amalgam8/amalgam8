# Amalgam8 Bookinfo sample

## Overview

The bookinfo sample is a simple application that implements a web page that displays information about a book, 
similar to a single catalog entry of an online book store. Displayed on the page is a description of the book,
book details (ISBN, number of pages, and so on), and a few book reviews.

The sample application is broken into four separate microservices:

* *productpage*. The productpage microservice calls the *details* and *reviews* microservices to populate the page. It provides a good example to experiment with both mid-tier and edge service routing.
* *details*. The details microservice contains book information.
* *reviews*. The reviews microservice contains book reviews. It also calls the *ratings* microservice, to provide two levels on downstream mid-tier routing.
* *ratings*. The ratings microservice contains booking ranking information that accompanies a book review. 

There are 3 versions of the reviews microservice:

* Version v1 doesn't call the ratings service.
* Version v2 calls the ratings service, and displays each rating as 1 to 5 black stars.
* Version v3 calls the ratings service, and displays each rating as 1 to 5 red stars.

![Microservice dependencies](https://github.com/amalgam8/examples/blob/master/apps/bookinfo/dependencies.jpg)

## Running the bookinfo demo

Before you begin, follow the environment set up instructions at https://github.com/amalgam8/examples/blob/master/README.md

1. Start the bookinfo sample app by running the following commands:

    ```
    cd $GOPATH/src/github.com/amalgam8/examples/apps/bookinfo
    ./run.sh
    ```
    The run.sh script starts the 4 bookinfo microservices. 

1. Confirm the microservices are running by running the following command:

    ```bash
    a8ctl service-list
    ```
    
    The following 4 microservices are displayed:
    
    ```
    +-------------+---------------------+
    | Service     | Instances           |
    +-------------+---------------------+
    | productpage | v1(1)               |
    | ratings     | v1(1)               |
    | details     | v1(1)               |
    | reviews     | v1(1), v2(1), v3(1) |
    +-------------+---------------------+
    ```

    The script started the v1 version of the `productpage`, `ratings`, and `details` services, and one 
    instance of each of the 3 different versions of the `reviews` service (v1, v2, and v3).

1. Route all of the incoming traffic to version v1 only for each service, by running the following commands:

    ```bash
    a8ctl route-set productpage --default v1
    a8ctl route-set ratings --default v1
    a8ctl route-set details --default v1
    a8ctl route-set reviews --default v1
    ```
    
1. Open http://192.168.33.33:32000/productpage/productpage from your browser
    and you should see the bookinfo application displayed. Notice that the
    product page is displayed, with no rating stars since `reviews:v1` does not
    access the ratings service.

1. Run the following command to send traffic to the v2 and v3 versions of the `reviews` service:

    ```
    a8ctl route-set reviews --default v1 --selector 'v2(user="john")' --selector 'v3(user="katherine")'
    ```

    This command specifies that only the designated user (john) will see v2 of the reviews service,
    and another designated user (katherine) will see v3 of the reviews service. 
    Any other user (including anonymous or not logged in) will use v1. 
  
1. Before you log in (or if you log in as anyone other than "john" or "katherine"), notice that there is no "rating" associated with the
    book reviews on the page. This is the v1 version of the reviews microservice. 

1. Click the *Sign in* button at the top-right corner of the web page, and log in as "john" (no password required). You will see a rating of 1-5 black stars next to each review, which was defined in the v2 version.

1. Click the *Sign in* button at the top-right corner of the web page, and log in as "katherine" (no password required). You will see the ratings displayed using red stars, which was defined in the v3 version.

There are many other features of the Amalgam8 microservice fabric that are highlighted using the bookinfo sample.
For example, you might want to take a look at the end-to-end
[test & deploy demo](https://github.com/amalgam8/examples/blob/master/demo-script.md) next.

## Shutting down

When you are finished, you can stop the bookinfo services with the following command:

```
./kill-services.sh
```
