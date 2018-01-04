FROM alpine

RUN apk add --update curl jq && \
    rm -rf /var/cache/apk/*

# get the latest version from github API

RUN curl -s https://api.github.com/repos/jaxxstorm/change-aws-credentials/releases/latest | jq -r '.assets[]| select(.browser_download_url | contains("linux")) | .browser_download_url' | xargs curl -L -o /tmp/change-aws-credentials.tar.gz

RUN tar zxvf /tmp/change-aws-credentials.tar.gz

RUN mv change-aws-credentials /usr/local/bin/change-aws-credentials

ENTRYPOINT ["/usr/local/bin/change-aws-credentials"] 
