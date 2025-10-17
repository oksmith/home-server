SERVICES = shutdown-service

.PHONY: deploy-all restart-all

deploy-all:
	@for service in $(SERVICES); do \
		$(MAKE) -C $$service deploy; \
	done

restart-all:
	@for service in $(SERVICES); do \
		$(MAKE) -C $$service restart; \
	done