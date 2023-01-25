FROM registry.access.redhat.com/ubi8/go-toolset:latest as build
COPY go.mod go.sum ./

RUN go mod download

COPY . ./

RUN make build

FROM registry.access.redhat.com/ubi8/ubi
ENV APP_DIR /opt/metaconflux
COPY --from=build /opt/app-root/src/_build/* ${APP_DIR}/

WORKDIR ${APP_DIR}

#ADD _build/* ${APP_DIR}/

CMD ${APP_DIR}/server
