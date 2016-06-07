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

## Running the microservices

Before you begin, follow the environment set up instructions at https://github.com/amalgam8/examples/blob/master/README.md

1. Start the bookinfo sample app by running the following commands:

  ```
    cd $GOPATH/src/github.com/amalgam8/examples/apps/bookinfo
    ./run.sh
  ```
  The run.sh script starts the 4 bookinfo microservices. 

2. Confirm the microservices are running by running the following cURL command:

  ```
    curl -X GET -H "Authorization: Bearer ${TOKEN}" http://${AR}/api/v1/services | jq .
     {
       "services": [
         "reviews",
         "ratings",
         "details",
         "productpage"
       ]
     }
   ```

3. Navigate in your browser to http://192.168.33.33:32000/productpage/productpage 

  The page displays information about a book titled, "NEW BOOK NAME HERE".

4. Review how the script operates, by running the following cURL command to see the instances of the reviews microservice:

  ```
    curl -X GET -H "Authorization: Bearer ${TOKEN}" http://${AR}/api/v1/services/reviews | jq .
    {
      "instances": [
        {
          "last_heartbeat": "2016-05-25T20:58:26.135709018Z",
          "metadata": {
            "version": "v1"
          },
          "status": "UP",
          "ttl": 45,
          "endpoint": {
            "value": "172.17.0.9:9080",
            "type": "http"
          },
          "service_name": "reviews",
          "id": "5f940f0ddee732bb"
        },
        {
          "last_heartbeat": "2016-05-25T20:58:26.262514572Z",
          "metadata": {
            "version": "v3"
          },
          "status": "UP",
          "ttl": 45,
          "endpoint": {
            "value": "172.17.0.11:9080",
            "type": "http"
          },
          "service_name": "reviews",
          "id": "eea7a5a4d9b10a1f"
        },
        {
          "last_heartbeat": "2016-05-25T20:58:26.367771984Z",
          "metadata": {
            "version": "v2"
          },
          "status": "UP",
          "ttl": 45,
          "endpoint": {
            "value": "172.17.0.10:9080",
            "type": "http"
          },
          "service_name": "reviews",
          "id": "05f853b7b4ab8b37"
        }
      ],
      "service_name": "reviews"
    }
  ```

 The script started one instance of each of the 3 different versions of the reviews service (v1, v2, and v3).

5. Open the run-services.sh script that was used to run the example. Review the following line of code:

  ```
   curl -X PUT ${AC}/v1/tenants/local/versions/reviews -d '{"default": "v1", "selectors": "{v2={user=\"frankb\"},v3={user=\"shriram\"}}" }'  -H "Content-Type: application/json"
  ```

 This command specifies that only the designated user (frankb) can use v2 of the reviews service, and another designated user (shriram) only uses v3 of the reviews service. Any other user (including anonymous or not logged in) will use v1. To review this in action:
 
6. Before you log in (or if you log in as anyone other than "frankb" or "shriram"), notice that there is no "rating" associated with the
book reviews on the page. This is the v1 version of the reviews microservice. 

7. Click the *Sign in* button at the top-right corner of the web page, and log in as "frankb" (no password required). You will see a rating of 1-5 black stars next to each review, which was defined in the v2 version.

8. Click the *Sign in* button at the top-right corner of the web page, and log in as "shriram" (no password required). You will see the ratings displayed using red stars, which was defined in the v3 version.

  There are many other features of the Amalgam8 microservice fabric that will be highlighted in future, using this bookinfo sample. Stay tuned!

## Shutting down

9. When you are finished, you can stop the bookinfo services with the following command:

  ```
    ./kill-services.sh
  ```

10. If you want to restart the services, without rebuilding the images, 
you can use the following script in place of run.sh from step (1):

  ```
    ./run-services.sh
  ```
