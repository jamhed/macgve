ARG BUILTBY
ARG DATE
ARG COMMIT
ARG VERSION

FROM golang:1.17-alpine AS build
ARG BUILTBY
ARG DATE
ARG COMMIT
ARG VERSION

WORKDIR /home
COPY src src
RUN cd src && go build -o ../macgve -ldflags "-X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE -X main.builtBy=$BUILTBY"

FROM alpine:3
COPY --from=build /home/macgve macgve
CMD ["./macgve"]
