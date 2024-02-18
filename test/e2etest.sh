#!/bin/bash

mkdir -p out

email="foo@example.com"
password="FooFooFooFoo"

stem=`sls info | grep 'api/v1/{Path+}' | sed -r 's/.*(https:\/\/[^\/]+).*/\1/'`

echo "URL Stem: $stem"

echo "Sign Up"
curl ${stem}/signup?email=${email}\&password=${password} 2>/dev/null

echo
echo "Log In"
l=`curl ${stem}/login?email=${email}\&password=${password} 2>/dev/null`

echo $l

token=`echo $l | jq -r .Token`

echo "Token: $token."

if [[ -z "$token" || "$token" == "null" ]] ; then
    echo "FAIL:  Auth Token not obtained."
    exit 22
fi      

echo "Loop"
curl -H "Authorization: $token" ${stem}/loop 

l=`curl -X POST -H "Authorization: $token" ${stem}/api/v1/group/testgroup 2>/dev/null`

echo $l

groupid=`echo $l | jq -r .Id`

httpstem=${stem}/api/v1/group/$groupid/counter

echo "HTTP stem:  ${httpstem}"

echo
echo "Tests"
venom run --var httpstem=${httpstem} --var token="$token" --output-dir out test/counter-api.test.yaml

