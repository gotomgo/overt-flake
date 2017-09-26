FROM centurylink/ca-certs
EXPOSE 4444

ADD configs/ configs
ADD artifacts/ofsrvr /

ENTRYPOINT ["/ofsrvr", "-config","./configs/default.config.yml"]
CMD ["-auth", "test"]
