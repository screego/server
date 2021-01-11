DOCKER_MANIFEST=DOCKER_CLI_EXPERIMENTAL=enabled docker manifest

docker-login-ci:
	docker login -u "$$DOCKER_USER" -p "$$DOCKER_PASS";

docker-manifest-annotate:
	echo ${VERSION}
	${DOCKER_MANIFEST} create --amend "screego/server:${VERSION}" "screego/server:amd64-${VERSION}" "screego/server:386-${VERSION}" "screego/server:armv7-${VERSION}" "screego/server:arm64-${VERSION}" "screego/server:ppc64le-${VERSION}"
	${DOCKER_MANIFEST} annotate "screego/server:${VERSION}" "screego/server:amd64-${VERSION}"   --os=linux --arch=amd64
	${DOCKER_MANIFEST} annotate "screego/server:${VERSION}" "screego/server:386-${VERSION}"     --os=linux --arch=386
	${DOCKER_MANIFEST} annotate "screego/server:${VERSION}" "screego/server:armv7-${VERSION}"   --os=linux --arch=arm --variant=v7
	${DOCKER_MANIFEST} annotate "screego/server:${VERSION}" "screego/server:arm64-${VERSION}"   --os=linux --arch=arm64
	${DOCKER_MANIFEST} annotate "screego/server:${VERSION}" "screego/server:ppc64le-${VERSION}" --os=linux --arch=ppc64le

docker-manifest-annotate-unstable:
	echo ${VERSION}
	${DOCKER_MANIFEST} create --amend "screego/server:unstable" "screego/server:amd64-unstable" "screego/server:386-unstable" "screego/server:armv7-unstable" "screego/server:arm64-unstable" "screego/server:ppc64le-unstable"
	${DOCKER_MANIFEST} annotate "screego/server:unstable" "screego/server:amd64-unstable"   --os=linux --arch=amd64
	${DOCKER_MANIFEST} annotate "screego/server:unstable" "screego/server:386-unstable"     --os=linux --arch=386
	${DOCKER_MANIFEST} annotate "screego/server:unstable" "screego/server:armv7-unstable"   --os=linux --arch=arm --variant=v7
	${DOCKER_MANIFEST} annotate "screego/server:unstable" "screego/server:arm64-unstable"   --os=linux --arch=arm64
	${DOCKER_MANIFEST} annotate "screego/server:unstable" "screego/server:ppc64le-unstable" --os=linux --arch=ppc64le


docker-manifest-push:
	${DOCKER_MANIFEST} push "screego/server:${VERSION}"

docker-manifest-push-unstable:
	${DOCKER_MANIFEST} push "screego/server:unstable"
