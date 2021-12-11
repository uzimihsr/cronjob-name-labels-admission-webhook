# cronjob-labels-admission-webhook

Kubernetes [MutatingAdmissionWebhook](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#mutatingadmissionwebhook) which adds a label to Jobs whose value is the name of the owner CronJob.  

```console
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

WIP

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