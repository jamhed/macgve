FROM golang:1.13-alpine AS build
WORKDIR /home
COPY src src
RUN cd src && go build -o ../macgve

FROM alpine:3.10
COPY --from=build /home/macgve macgve
CMD ["./macgve"]
