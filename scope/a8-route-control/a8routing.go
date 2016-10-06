package main

import (
    //"net/http"
    "os/exec"
    "log"
    //"io/ioutil"
)

func a8containerIPaddresses() {
    // map of IP address strings
    //m := make(map[string]string)

/*
    registryUrl := os.Getenv("A8_REGISTRY_URL")
    resp, err := http.Get("http://172.17.0.1:31300/api/v1/services")
    if err != nil {
        // handle error
        log.Println("error with request")
        log.Println(err)
        return
    }
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    log.Println("response: ")
    log.Println(body)
    */
    log.Println("hi")

    out, err := exec.Command("curl -H 'Accept: application/json' http://172.17.0.1:31300/api/v1/services/").Output()
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("%s", out)
}

func getContainerIDs() {
    return
}