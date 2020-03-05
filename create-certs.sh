#! /bin/bash

WEBHOOK_NS=${1:-"mgmt"}
WEBHOOK_NAME=${2:-"labeling-validator"}
WEBHOOK_SVC="${WEBHOOK_NAME}-webhook"

# Create certs for our webhook
openssl genrsa -out webhookCA.key 4096
openssl req -new -key ./webhookCA.key -subj "/CN=${WEBHOOK_SVC}.${WEBHOOK_NS}.svc" -out ./webhookCA.csr
openssl x509 -req -days 365 -in webhookCA.csr -signkey webhookCA.key -out webhook.crt

# Create certs secrets for k8s
kubectl create secret generic \
    ${WEBHOOK_SVC}-certs \
    --namespace=${WEBHOOK_NS} \
    --from-file=key.pem=./webhookCA.key \
    --from-file=cert.pem=./webhook.crt \
    --dry-run -o yaml > ./deploy/webhook-certs.yaml

# Set the CABundle on the webhook registration
CA_BUNDLE=$(cat ./webhook.crt | base64 -w0)
sed -e "s/CA_BUNDLE/${CA_BUNDLE}/" -e "s/WEBHOOK_SVC/${WEBHOOK_SVC}/" -e "s/WEBHOOK_NS/${WEBHOOK_NS}/" ./deploy/webhook-registration.yaml.tpl > ./deploy/webhook-registration.yaml
sed -e "s/CA_BUNDLE/${CA_BUNDLE}/" -e "s/WEBHOOK_SVC/${WEBHOOK_SVC}/" -e "s/WEBHOOK_NS/${WEBHOOK_NS}/" ./deploy/webhook.yaml.tpl > ./deploy/webhook.yaml

# Clean
rm ./webhookCA* && rm ./webhook.crt
