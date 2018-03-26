FROM golang:1.9

WORKDIR /go/src/github.com/jaxxstorm/change-aws-credentials

COPY . .

RUN go get -v github.com/Masterminds/glide

RUN cd $GOPATH/src/github.com/Masterminds/glide && git checkout tags/v0.12.3 && go install && cd -

RUN ls .

ENTRYPOINT ["./change-aws-credentials"] 
