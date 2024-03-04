#!/bin/bash

venom=~/builds/venom-1.1.0/dist/venom.linux-amd64

mkdir -p out

email="foo@example.com"
password="FooFooFooFoo"

stem=`sls info | grep 'signup' | sed -r 's/.*(https:\/\/[^\/]+).*/\1/'`

echo "URL Stem: $stem"

echo "Sign Up"
curl ${stem}/signup?email=${email}\&password=${password} 

sleep 2

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

httpstem=${stem}

echo "HTTP stem:  ${httpstem}"

echo
echo "Tests"
${venom} run --var email=${email} --var password=${password} --var httpstem=${httpstem} --var token="$token" --output-dir out test/counter-api.test.yaml

