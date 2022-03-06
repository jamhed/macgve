ARG BUILTBY DATE COMMIT VERSION

FROM golang:1.17-alpine AS build
ARG BUILTBY DATE COMMIT VERSION

WORKDIR /home/src
COPY src/go.mod src/go.sum ./
RUN go mod download

COPY src/ .
RUN --mount=type=cache,id=gobuild,target=/root/.cache/go-build \
	go build -v -o ../macgve -ldflags \
	"-X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE -X main.builtBy=$BUILTBY"

FROM alpine:3
COPY --from=build /home/macgve macgve
CMD ["./macgve"]
