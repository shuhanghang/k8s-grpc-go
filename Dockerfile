FROM golang:1.20 AS build-stage
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn
WORKDIR /app
COPY go.mod go.sum .
COPY ./server ./server
COPY ./utils ./utils
COPY ./pb ./pb
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /grpc_server ./server/main.go


FROM alpine AS build-release-stage
WORKDIR /
COPY --from=build-stage /grpc_server /grpc_server
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories && \
    apk --no-cache add tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone 
EXPOSE 22222
ENTRYPOINT ["/grpc_server"]