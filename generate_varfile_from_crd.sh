#!/usr/bin/env sh

set -o errexit
set -o nounset
set -o pipefail

apiserver=https://kubernetes.default.svc
serviceaccount=/var/run/secrets/kubernetes.io/serviceaccount
ns=$(cat ${serviceaccount}/namespace)
token=$(cat ${serviceaccount}/token)
cacert=${serviceaccount}/ca.crt

printf cert_dir=\"/certificates\"\\nuser_permissions= > terraform.tfvars

curl -s --cacert ${cacert} --header "Authorization: Bearer ${token}" -X GET ${apiserver}/apis/example.com/v1/users|jq '[.items[]|{user_name: .metadata.name, user_group: .spec.userGroup, role_rules: .spec.roleRules}]'|sed 's/apiGroups/api_groups/g;s/":/"=/g' >> terraform.tfvars