FROM martenseemann/quic-network-simulator-endpoint:latest

# Tutorial for use with QUIC Network Simulator: https://github.com/marten-seemann/quic-network-simulator

# Erro no serviço "sim":ip6tables v1.8.4 (legacy): can't initialize ip6tables table `filter': Table does not exist (do you need to insmod?)
# Solução: https://ilhicas.com/2018/04/08/Fixing-do-you-need-insmod.html (sudo modprobe ip6table_filter)

RUN apt-get update

# download and build your QUIC implementation
RUN apt-get install -y wget tar git vim python

RUN wget https://go.dev/dl/go1.19.5.linux-amd64.tar.gz && \
  tar xfz go1.19.5.linux-amd64.tar.gz && \
  rm go1.19.5.linux-amd64.tar.gz

ENV PATH="/go/bin:${PATH}"
ENV GO111MODULE=on

RUN git clone -b quic-simulator https://github.com/eriklima/http3-quic.git

WORKDIR /http3-quic
RUN go get ./...

# copy run script and run it
COPY run_endpoint.sh .
RUN chmod +x run_endpoint.sh
ENTRYPOINT [ "./run_endpoint.sh" ]