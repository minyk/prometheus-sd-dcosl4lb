FROM prom/prometheus:v2.8.0

RUN mkdir -p /etc/prometheus/runit/prometheus && mkdir -p /etc/prometheus/runit/sd
COPY build/prometheus-sd-dcosl4lb-linux /bin/prometheus-sd-dcosl4lb-linux
COPY docker/prometheus /etc/prometheus/runit/prometheus/run
COPY docker/prometheus-sd-dcosl4lb /etc/prometheus/runit/sd/run
COPY docker/prometheus.yml /etc/prometheus/prometheus.yml

ENTRYPOINT ["/bin/runsvdir"]
CMD ["-P", "/etc/prometheus/runit"]