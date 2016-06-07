#!/bin/bash
#gunicorn -D -w 1 -b 0.0.0.0:10081 --reload details:app
#gunicorn -D -w 1 -b 0.0.0.0:10082 --reload reviews:app
#gunicorn -w 1 -b 0.0.0.0:19080 --reload --access-logfile prod.log --error-logfile prod.log productpage:app >>prod.log 2>&1 &
kubectl create -f bookinfo.yaml
if [ $? -ne 0 ]; then
    echo -e "\n***********\nFAILED: bookinfo create.\n***********\n"
    exit 1
fi
#echo "waiting for a few seconds for pods to get created.."
#sleep 10
echo "setting version routing rules for productpage, reviews, ratings, and details"
AC=$(kubectl get svc/controller --template={{.spec.clusterIP}}:{{\("index .spec.ports 0"\).port}})
curl -X PUT ${AC}/v1/tenants/local/versions/productpage -d '{"default": "v1"}' -H "Content-Type: application/json" 
#curl -X PUT ${AC}/v1/tenants/local/versions/reviews -d '{"default": "v1", "selectors": "{v2={weight=0.3},v3={weight=0.2}}" }' -H "Content-Type: application/json" 
curl -X PUT ${AC}/v1/tenants/local/versions/reviews -d '{"default": "v1", "selectors": "{v2={user=\"frankb\"},v3={user=\"shriram\"}}" }' -H "Content-Type: application/json" 
curl -X PUT ${AC}/v1/tenants/local/versions/ratings -d '{"default": "v1"}' -H "Content-Type: application/json" 
curl -X PUT ${AC}/v1/tenants/local/versions/details -d '{"default": "v1"}' -H "Content-Type: application/json" 
