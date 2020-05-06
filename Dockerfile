FROM fluent/fluentd:v1.10.3-1.0

ADD fluentd.conf /fluentd/etc/fluentd.conf
ADD agent /usr/bin/agent

ENTRYPOINT ["/usr/bin/agent"]