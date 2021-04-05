FROM grafana/grafana:7.5.1
USER root
RUN  wget https://golang.org/dl/go1.16.2.linux-amd64.tar.gz \
 && tar -C /usr/local -xzf go1.16.2.linux-amd64.tar.gz \
 && rm go1.16.2.linux-amd64.tar.gz
USER grafana
RUN /usr/local/go/bin/go install github.com/go-delve/delve/cmd/dlv@latest
