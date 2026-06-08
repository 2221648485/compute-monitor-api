#!/usr/bin/env sh
set -eu

ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
IMAGE_NAME="${IMAGE_NAME:-compute-monitor-api}"
IMAGE_VERSION="${IMAGE_VERSION:-1.0.0}"
IMAGE_TAG="${IMAGE_NAME}:${IMAGE_VERSION}"
IMAGE_DIR="${ROOT_DIR}/deploy/images"
IMAGE_TAR="${IMAGE_DIR}/${IMAGE_NAME}-${IMAGE_VERSION}.tar"
COMPOSE_FILE="${ROOT_DIR}/deployments/docker-compose/docker-compose.yml"

echo "======================== Start build compute-monitor-api ========================"
echo "Root: ${ROOT_DIR}"
echo "Image: ${IMAGE_TAG}"

cd "${ROOT_DIR}"

echo "[STEP] docker build ${IMAGE_TAG}"
docker build -t "${IMAGE_TAG}" .

echo "[STEP] validate docker compose"
docker compose -f "${COMPOSE_FILE}" config --quiet

echo "[STEP] docker save image to tar"
mkdir -p "${IMAGE_DIR}"
docker save -o "${IMAGE_TAR}" "${IMAGE_TAG}"

echo "======================== Build finished ========================"
echo "Image: ${IMAGE_TAG}"
echo "Tar: ${IMAGE_TAR}"
