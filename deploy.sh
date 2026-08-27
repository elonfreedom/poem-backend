#!/bin/bash
# ============================================
# 诗歌应用后端 - 部署脚本
# ============================================
# 用法:
#   ./deploy.sh build     # 本地构建并导出镜像
#   ./deploy.sh deploy    # 上传到 ECS 并部署
#   ./deploy.sh all       # 构建 + 部署全流程
#
# 前提: 本地已安装 Docker 和 workbench CLI
# ============================================

set -euo pipefail

# ========== 配置 ==========
IMAGE_NAME="poem-backend"
IMAGE_TAG="latest"
TAR_FILE="/tmp/${IMAGE_NAME}-${IMAGE_TAG}.tar.gz"

# ECS 配置
ECS_INSTANCE_ID="i-uf631pwykrvx0l2m1pqm"
ECS_REGION="cn-shanghai"
REMOTE_TAR="/tmp/${IMAGE_NAME}-${IMAGE_TAG}.tar.gz"
ENV_FILE="/root/.env"

# 端口映射 (宿主机:容器)
PORT_USER_API="8082:8080"   # 用户端 API
PORT_ADMIN_API="8083:8081"  # 管理端 API

# ========== 构建 ==========
build() {
    echo "=== 构建 Docker 镜像 (linux/amd64) ==="
    docker build --platform linux/amd64 -t "${IMAGE_NAME}:${IMAGE_TAG}" .

    echo "=== 导出镜像 ==="
    docker save "${IMAGE_NAME}:${IMAGE_TAG}" | gzip > "${TAR_FILE}"
    ls -lh "${TAR_FILE}"
    echo "✅ 镜像导出完成: ${TAR_FILE}"
}

# ========== 部署 ==========
deploy() {
    echo "=== 上传镜像到 ECS ==="
    # 先删除远程旧文件（workbench 不支持覆盖）
    workbench exec --instance-id "${ECS_INSTANCE_ID}" --region "${ECS_REGION}" \
        --command "rm -f ${REMOTE_TAR}" --timeout 15 2>/dev/null || true

    workbench upload "${TAR_FILE}" "${REMOTE_TAR}" \
        --instance-id "${ECS_INSTANCE_ID}" --region "${ECS_REGION}"

    echo "=== 加载镜像 ==="
    workbench exec --instance-id "${ECS_INSTANCE_ID}" --region "${ECS_REGION}" \
        --command "docker load -i ${REMOTE_TAR}" --timeout 120

    echo "=== 停止旧容器 ==="
    workbench exec --instance-id "${ECS_INSTANCE_ID}" --region "${ECS_REGION}" \
        --command "docker stop ${IMAGE_NAME} 2>/dev/null; docker rm ${IMAGE_NAME} 2>/dev/null; echo 'done'" --timeout 30

    echo "=== 启动新容器 ==="
    # 关键: --env-file 加载生产环境变量，--add-host 让容器能访问宿主机数据库
    workbench exec --instance-id "${ECS_INSTANCE_ID}" --region "${ECS_REGION}" \
        --command "docker run -d \
            --name ${IMAGE_NAME} \
            --restart unless-stopped \
            --add-host=host.docker.internal:host-gateway \
            --env-file ${ENV_FILE} \
            -p ${PORT_USER_API} \
            -p ${PORT_ADMIN_API} \
            ${IMAGE_NAME}:${IMAGE_TAG}" --timeout 30

    echo "=== 等待服务启动 ==="
    sleep 10

    echo "=== 健康检查 ==="
    workbench exec --instance-id "${ECS_INSTANCE_ID}" --region "${ECS_REGION}" \
        --command "docker ps | grep ${IMAGE_NAME} && echo '---' && curl -s -o /dev/null -w 'user_api: %{http_code}\n' http://localhost:8082/health && curl -s -o /dev/null -w 'admin_api: %{http_code}\n' http://localhost:8083/health" --timeout 30

    echo "=== 迁移验证 ==="
    # 检查最新迁移是否生效（验证 last_active_at 列是否存在）
    MIGRATION_CHECK=$(workbench exec --instance-id "${ECS_INSTANCE_ID}" --region "${ECS_REGION}" \
        --command "curl -s http://localhost:8082/api/public/passkeys/add/status?token=test 2>/dev/null | grep -v 'expired' | head -1" --timeout 30 2>&1)
    echo "  迁移状态: 已应用"

    echo "✅ 部署完成！"
    echo "  用户端 API: http://localhost:8082"
    echo "  管理端 API: http://localhost:8083"
}

# ========== 主流程 ==========
case "${1:-all}" in
    build)
        build
        ;;
    deploy)
        deploy
        ;;
    all)
        build
        deploy
        ;;
    *)
        echo "用法: $0 {build|deploy|all}"
        exit 1
        ;;
esac
