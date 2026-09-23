ARG BASE_IMAGE=${BASE_IMAGE:-registry.redhat.io/openshift4/ose-operator-registry-rhel9:v4.22}
FROM ${BASE_IMAGE}
ARG OCP_VERSION=${OCP_VERSION:-fbc-v4-22}

ENTRYPOINT ["/bin/opm"]
CMD ["serve", "/configs", "--cache-dir=/tmp/cache"]

ADD licenses/ /licenses/
ADD ${OCP_VERSION}/catalog/ /configs
RUN ["/bin/opm", "serve", "/configs", "--cache-dir=/tmp/cache", "--cache-only"]

LABEL operators.operatorframework.io.index.configs.v1=/configs
