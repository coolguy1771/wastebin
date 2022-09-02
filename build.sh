
#!/usr/bin/env bash
# build, tag, and push docker images

# exit if a command fails
set -o errexit

# if no registry is provided, tag image as "local" registry
registry="ghcr.io/coolguy1771"


image_version="0.0.1"

# set image name
image_name="wastebin"


# copy native image to local image repository
docker buildx build \
    -t "${registry}/${image_name}:${image_version}" \
    $(if [ "${LATEST}" == "yes" ]; then echo "-t ${registry}/${image_name}:latest"; fi) \
    -f Dockerfile . \
    --load

# push amd64 and arm images to remote registry
docker buildx build --platform linux/amd64,linux/arm,linux/arm64 \
    -t "${registry}/${image_name}:${image_version}" \
    $(if [ "${LATEST}" == "yes" ]; then echo "-t ${registry}/${image_name}:latest"; fi) \
    -f Dockerfile . \
    --push