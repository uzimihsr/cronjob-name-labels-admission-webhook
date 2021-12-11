# cronjob-labels-admission-webhook


## develop

create private key and self-signed TLS certificate.  

```console
openssl genrsa -out tls.key

openssl req -x509 -key tls.key -out tls.crt -addext 'subjectAltName = DNS:cronjob-labels-admission-webhook.default.svc'
# specify CN: cronjob-labels-admission-webhook.default.svc
# reference : https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#service-reference
```

build the image and load it to kind node.  

```console
docker image build -t cronjob-labels-admission-webhook:develop .
kind load docker-image cronjob-labels-admission-webhook:develop
```

create Secret, Deployment, Service, and MutatingWebhookConfiguration.  

```console
kubectl create secret tls tls-secret --cert=tls.crt --key=tls.key
sed "s/CA_BUNDLE/$(kubectl get secret tls-secret -o jsonpath='{.data.tls\.crt}')/g" manifests/mutatingwebhookconfiguration.yaml | kubectl apply -f -
kubectl apply -f manifests/deployment.yaml
kubectl apply -f manifests/service.yaml
```