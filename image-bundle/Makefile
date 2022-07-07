.docker-bundler.stamp: Dockerfile image.list
	docker build -t ekebuild$(basename $@) .
	touch $@

bundle.tar: .docker-bundler.stamp
	docker run --rm -v $(shell pwd):/hostdir -v /var/lib/containerd --privileged ekebuild.docker-bundler


clean:
	rm -f bundle.tar
	rm -f .docker-bundler.stamp