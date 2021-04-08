DOCKER_MANIFEST=DOCKER_CLI_EXPERIMENTAL=enabled docker manifest

docker-login-ci:
	docker login -u "$$DOCKER_USER" -p "$$DOCKER_PASS";

docker-manifest-annotate:
	echo ${VERSION}
	${DOCKER_MANIFEST} create --amend "l11r/screego:unstable"     "l11r/screego:amd64-unstable"     "l11r/screego:386-unstable"     "l11r/screego:armv7-unstable"     "l11r/screego:arm64-unstable"     "l11r/screego:ppc64le-unstable"
	${DOCKER_MANIFEST} create --amend "l11r/screego:${VERSION}" "l11r/screego:amd64-${VERSION}" "l11r/screego:386-${VERSION}" "l11r/screego:armv7-${VERSION}" "l11r/screego:arm64-${VERSION}" "l11r/screego:ppc64le-${VERSION}"
	${DOCKER_MANIFEST} annotate "l11r/screego:unstable"     "l11r/screego:amd64-unstable"       --os=linux --arch=amd64
	${DOCKER_MANIFEST} annotate "l11r/screego:${VERSION}" "l11r/screego:amd64-${VERSION}"   --os=linux --arch=amd64
	${DOCKER_MANIFEST} annotate "l11r/screego:unstable"     "l11r/screego:386-unstable"         --os=linux --arch=386
	${DOCKER_MANIFEST} annotate "l11r/screego:${VERSION}" "l11r/screego:386-${VERSION}"     --os=linux --arch=386
	${DOCKER_MANIFEST} annotate "l11r/screego:unstable"     "l11r/screego:armv7-unstable"       --os=linux --arch=arm --variant=v7
	${DOCKER_MANIFEST} annotate "l11r/screego:${VERSION}" "l11r/screego:armv7-${VERSION}"   --os=linux --arch=arm --variant=v7
	${DOCKER_MANIFEST} annotate "l11r/screego:unstable"     "l11r/screego:arm64-unstable"       --os=linux --arch=arm64
	${DOCKER_MANIFEST} annotate "l11r/screego:${VERSION}" "l11r/screego:arm64-${VERSION}"   --os=linux --arch=arm64
	${DOCKER_MANIFEST} annotate "l11r/screego:unstable"     "l11r/screego:ppc64le-unstable"     --os=linux --arch=ppc64le
	${DOCKER_MANIFEST} annotate "l11r/screego:${VERSION}" "l11r/screego:ppc64le-${VERSION}" --os=linux --arch=ppc64le


docker-manifest-push:
	${DOCKER_MANIFEST} push "l11r/screego:${VERSION}"
	${DOCKER_MANIFEST} push "l11r/screego:unstable"

