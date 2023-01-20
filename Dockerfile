FROM registry.access.redhat.com/ubi8/ubi

ENV APP_DIR /opt/metaconflux
WORKDIR ${APP_DIR}

ADD _build/* ${APP_DIR}/



CMD ${APP_DIR}/server