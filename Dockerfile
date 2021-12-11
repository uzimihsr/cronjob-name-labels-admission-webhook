FROM golang AS build

WORKDIR /cronjob-name-labels-admission-webhook

COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /out/cronjob-name-labels-admission-webhook .

FROM scratch

COPY --from=build /out/cronjob-name-labels-admission-webhook /

ENTRYPOINT [ "/cronjob-name-labels-admission-webhook" ]