FROM alpine

COPY Digital_Ocean_Cluster /usr/local/bin

ENTRYPOINT ["Digital_Ocean_Cluster"]
