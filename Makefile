NAME     = trigger-proxy
REGISTRY = vebis
VERSION  = 1.1.0

.PHONY: build clean

all: build

build:
	@docker build --rm=true -t $(REGISTRY)/$(NAME):$(VERSION) .
	@docker tag $(REGISTRY)/$(NAME):$(VERSION) $(REGISTRY)/$(NAME):latest
	@docker images $(REGISTRY)/$(NAME)

push: build
	@docker push $(REGISTRY)/$(NAME):$(VERSION)
	@docker push $(REGISTRY)/$(NAME):latest

clean:
	@docker rmi $(REGISTRY)/$(NAME):$(VERSION)
	@docker rmi $(REGISTRY)/$(NAME):latest

default: build
