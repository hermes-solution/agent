FROM fluent/fluentd:v1.10.3-1.0

USER root
RUN mkdir -p /hermes \
	&& mkdir -p /hermes/bin

COPY out_http_with_injection.rb /etc/fluent/plugin/

COPY forward.conf /hermes/forward.conf
COPY agent /hermes/bin/agent
COPY entrypoint.sh /hermes/bin/entrypoint.sh

EXPOSE 8080

ENV PATH="/hermes/bin:${PATH}"
ENTRYPOINT ["tini",  "--", "/hermes/bin/entrypoint.sh"]
CMD ["agent"]