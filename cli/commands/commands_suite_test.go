// Copyright 2016 IBM Corporation
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package commands_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCommands(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Commands Suite")
}

var response = make(map[string][]byte)

var _ = Describe("Commands", func() {
	response["reviews_traffic_stopped"] = []byte(`
		{
		  "services": {
		    "reviews": [
				{
					"id": "41202250-8b4d-4fb4-a000-74ebb04857e4",
					"priority": 1,
					"destination": "reviews",
					"route": {
						"backends": [
							{
								"tags": [
									"v1"
								]
							}
						]
					}
				}
			]
		}
	}`)

	response["reviews_traffic_started"] = []byte(`
		{
		  "services": {
		    "reviews": [
				{
					"id": "41202250-8b4d-4fb4-a000-74ebb04857e4",
					"priority": 1,
					"destination": "reviews",
					"route": {
						"backends": [
							{
								"tags": [
									"v2"
								],
								"weight": 0.1
							},
							{
								"tags": [
									"v1"
								]
							}
						]
					}
				}
			]
		}
	}`)

	response["weight_not_zero"] = []byte(`
			{
				"rules": [
					{
						"id": "41202250-8b4d-4fb4-a000-74ebb04857e4",
						"priority": 1,
						"destination": "reviews",
						"route": {
							"backends": [
								{
									"tags": [
										"v1"
									],
									"weight": 0.1
								}
							]
						}
					}
				],
				"revision": 1
			}`)

	response["no_rules"] = []byte(`
		{"services":{}}
	`)

	response["inactive"] = []byte(`
		{"Error":"Failed to enumerate service names"}
	`)

	response["reviews"] = []byte(`
		{
			"service_name": "reviews",
			"instances": [
				{
					"id": "5f940f0ddee732bb",
					"service_name": "reviews",
					"endpoint": {
						"type": "http",
						"value": "172.17.0.9:9080"
					},
					"ttl": 60,
					"status": "UP",
					"last_heartbeat": "2016-11-22T22:25:56.02658653Z",
					"tags": [
						"v3"
					]
				},
				{
					"id": "9b9776db6ac79b56",
					"service_name": "reviews",
					"endpoint": {
						"type": "http",
						"value": "172.17.0.14:9080"
					},
					"ttl": 60,
					"status": "UP",
					"last_heartbeat": "2016-11-22T22:25:51.192761423Z",
					"tags": [
						"v2"
					]
				},
				{
					"id": "eea7a5a4d9b10a1f",
					"service_name": "reviews",
					"endpoint": {
						"type": "http",
						"value": "172.17.0.11:9080"
					},
					"ttl": 60,
					"status": "UP",
					"last_heartbeat": "2016-11-22T22:25:51.011945888Z",
					"tags": [
						"v1"
					]
				}
			]
		}`)

	response["reviews_two_rules"] = []byte(`
		{
		  "services": {
		    "reviews": [
				{
					"id": "41202250-8b4d-4fb4-a000-74ebb04857e4",
					"priority": 1,
					"destination": "reviews",
					"route": {
						"backends": [
							{
								"tags": [
									"v1"
								]
							}
						]
					}
				},
				{
					"id": "41202250-8b4d-4fb4-a000-74ebb04857e5",
					"priority": 1,
					"destination": "reviews",
					"route": {
						"backends": [
							{
								"tags": [
									"v2"
								]
							}
						]
					}
				}
			]
		}
	}`)
})
