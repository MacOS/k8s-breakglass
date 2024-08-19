FROM gcr.io/distroless/static:nonroot

ARG BINARY_SOURCE_PATH=./breakglass

WORKDIR /

COPY  ${BINARY_SOURCE_PATH} /breakglass
COPY  ./frontend/dist /frontend/dist

USER 65532
ENTRYPOINT [ "/breakglass" ]
