# cronjob-name-labels-admission-webhook

Kubernetes [MutatingAdmissionWebhook](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#mutatingadmissionwebhook) which adds a label to Jobs whose value is the name of the owner CronJob.  

```bash
$ kubectl get cronjobs
NAME      SCHEDULE      SUSPEND   ACTIVE   LAST SCHEDULE   AGE
her-job   */1 * * * *   False     0        17s             71m
his-job   */1 * * * *   False     0        17s             71m
my-job    */1 * * * *   False     0        17s             75m

$ kubectl get jobs --show-labels
NAME               COMPLETIONS   DURATION   AGE     LABELS
her-job-27320400   1/1           9s         2m47s   uzimihsr.github.io/cronjob-name=her-job
her-job-27320401   1/1           3s         107s    uzimihsr.github.io/cronjob-name=her-job
her-job-27320402   1/1           4s         47s     uzimihsr.github.io/cronjob-name=her-job
his-job-27320400   1/1           4s         2m47s   uzimihsr.github.io/cronjob-name=his-job
his-job-27320401   1/1           6s         107s    uzimihsr.github.io/cronjob-name=his-job
his-job-27320402   1/1           3s         47s     uzimihsr.github.io/cronjob-name=his-job
my-job-27320400    1/1           5s         2m47s   uzimihsr.github.io/cronjob-name=my-job
my-job-27320401    1/1           4s         107s    uzimihsr.github.io/cronjob-name=my-job
my-job-27320402    1/1           7s         47s     uzimihsr.github.io/cronjob-name=my-job

$ kubectl get jobs -l uzimihsr.github.io/cronjob-name=my-job
NAME              COMPLETIONS   DURATION   AGE
my-job-27320400   1/1           5s         2m52s
my-job-27320401   1/1           4s         112s
my-job-27320402   1/1           7s         52s
```

## install

create TLS Secret.  
ServerName should be `<service-name>.<namespace>.svc` : https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#service-reference  

```bash
# create private key
openssl genrsa -out tls.key

# create self-signed certificate
openssl req -x509 -key tls.key -out tls.crt -days 3650 -addext 'subjectAltName = DNS:cronjob-name-labels-admission-webhook.default.svc'
# Common Name (e.g. server FQDN or YOUR name) []:cronjob-name-labels-admission-webhook.default.svc

# create TLS secret
kubectl create secret tls cronjob-name-labels-admission-webhook-tls-secret --cert=tls.crt --key=tls.key
```

create Secret, Deployment, Service, and MutatingWebhookConfiguration at `default` namespace.  

```bash
# apply manifest
curl https://raw.githubusercontent.com/uzimihsr/cronjob-name-labels-admission-webhook/main/manifests/manifest.yaml \
  | sed "s/CA_BUNDLE/$(kubectl get secret cronjob-name-labels-admission-webhook-tls-secret -o jsonpath='{.data.tls\.crt}')/g" \
  | kubectl apply -f -
```

## develop

create secret.  

```bash
kubectl create secret tls cronjob-name-labels-admission-webhook-tls-secret --cert=tls.crt --key=tls.key
```

build the image and load it to kind node.  

```bash
docker image build -t cronjob-name-labels-admission-webhook:develop .
kind load docker-image cronjob-name-labels-admission-webhook:develop
```

create Secret, Deployment, Service, and MutatingWebhookConfiguration.  

```bash
kubectl apply -f manifests/deployment.yaml
kubectl wait --for=condition=ready pod --selector=app=cronjob-name-labels-admission-webhook --timeout=90s
kubectl apply -f manifests/service.yaml
sed "s/CA_BUNDLE/$(kubectl get secret cronjob-name-labels-admission-webhook-tls-secret -o jsonpath='{.data.tls\.crt}')/g" manifests/mutatingwebhookconfiguration.yaml | kubectl apply -f -
```

if you want to delete it :(  

```bash
kubectl delete mutatingwebhookconfiguration cronjob-name-labels-admission-webhook.default.svc
kubectl delete service cronjob-name-labels-admission-webhook
kubectl delete deployment cronjob-name-labels-admission-webhook
kubectl delete secret cronjob-name-labels-admission-webhook-tls-secret
```