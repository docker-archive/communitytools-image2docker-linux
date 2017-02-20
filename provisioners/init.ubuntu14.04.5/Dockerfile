RUN apt-get update && apt-get install -y runit
ENTRYPOINT ["runsvdir","-P","/etc/service"]
STOPSIGNAL SIGHUP
