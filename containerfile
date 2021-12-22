FROM golang:latest AS build-layer

COPY . /
RUN go mod tidy && go mod vendor
ARG RUN_TYPE=""
RUN CGO_ENABLED=0 GOARCH=arm64 GOOS=linux go build -o server.app cmd/${RUN_TYPE}


FROM scratch

COPY --from=build-layer server.app .
EXPOSE 8080

ARG DEBUG=false
ARG DATABASEURI
ARG DATASETURI
ARG BUCKET
ARG OBJECTNAME
ARG NAME
ARG PUBLICPORT
ARG DOWNLOADDIR

ENV debug=${DEBUG}
ENV database_uri=${DATABASEURI}
ENV dataset_uri=${DATASETURI}
ENV bucket=${BUCKET}
ENV objectname=${OBJECTNAME}
ENV name=${NAME}
ENV publicport=${PUBLICPORT}
ENV downloaddir=${DOWNLOADDIR}

CMD [ "server.app" ]