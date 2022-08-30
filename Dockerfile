FROM golang:1.16

# 将代码复制到容器中
COPY . /app

WORKDIR /app

RUN export  GOPROXY=https://goproxy.cn
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go mod tidy
RUN go build ./main.go


#打包
FROM centos:7.9

COPY --from=builder /app/main /app/main

COPY --from=builder /app/conf/ /app/conf/

RUN yum -y install git && \
    mkdir /root/.ssh && \
    chmod 700 /root/.ssh && \
    cp /app/conf/id_rsa /root/.ssh/ && \
    chmod 600 /root/.ssh/id_rsa && \
    cp /app/conf/id_rsa.pub /root/.ssh/ && \
    chmod 644 /root/.ssh/id_rsa.pub && \
    ssh  -o StrictHostKeyChecking=no git@github.com

WORKDIR /app

CMD ["./main"]