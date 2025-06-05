FROM node:alpine AS build
RUN apk update && apk add go git typescript
ENV CGO_ENABLED=0
RUN mkdir /src
WORKDIR /src
COPY go.mod go.sum /src
RUN go mod download
COPY . /src
WORKDIR /src/ui
RUN yarn
RUN yarn build
WORKDIR /src
RUN go build \
	-trimpath \
	-tags="netgo osusergo" \
	-ldflags="-s -w -X main.mode=prod -X 'main.version=$(git describe --always --tags)' -X 'main.version=main.commitHash=$(git log -1 --format=%H)'" \
	-o /src/screego

FROM scratch
USER 1001
COPY --from=build /src/screego /screego
EXPOSE 3478/tcp
EXPOSE 3478/udp
EXPOSE 5050
WORKDIR "/"
ENTRYPOINT [ "/screego" ]
CMD ["serve"]
