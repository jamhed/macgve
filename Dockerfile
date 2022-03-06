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
COPY src/go.mod src/go.sum src/
RUN cd src && go mod graph | awk '{if ($1 !~ "@") print $2}' | xargs go get -v

COPY src/ src/
RUN --mount=type=cache,id=gobuild,target=/root/.cache/go-build \
	cd src && go build -v -o ../macgve -ldflags \
	"-X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE -X main.builtBy=$BUILTBY"

FROM alpine:3
COPY --from=build /home/macgve macgve
CMD ["./macgve"]
