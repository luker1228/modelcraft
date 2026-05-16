// ---------------------------------------------------------------------------
// 辅助函数：判断是否为有效的 CI 目标引用（PR / main 分支 / tag）
// ---------------------------------------------------------------------------
def isCiTargetRef() {
  return (env.CHANGE_ID?.trim()) ||
         env.BRANCH_NAME == 'main' ||
         (env.TAG_NAME?.trim())
}

// ---------------------------------------------------------------------------
// 辅助函数：返回镜像 tag
//   - git tag  → v1.2.3
//   - main     → abc1234（7 位 commit hash）
//   - PR       → pr-42
// ---------------------------------------------------------------------------
def imageTag() {
  if (env.TAG_NAME?.trim())      return env.TAG_NAME
  if (env.BRANCH_NAME == 'main') return env.GIT_COMMIT.take(7)
  return "pr-${env.CHANGE_ID}"
}

// ---------------------------------------------------------------------------
// Pipeline 主体
// ---------------------------------------------------------------------------
pipeline {
  agent any

  options {
    timestamps()
    disableConcurrentBuilds()
    buildDiscarder(logRotator(numToKeepStr: '30'))
    timeout(time: 45, unit: 'MINUTES')
  }

  environment {
    // compose 文件位于根目录
    COMPOSE_FILE = 'docker-compose.yml'
    // 项目名固定，避免 compose 因工作目录不同产生多个 project
    COMPOSE_PROJECT_NAME = 'modelcraft'
    // Node.js 版本（前端），与 modelcraft-front 保持一致
    NODE_VERSION = "${env.NODE_VERSION ?: '20'}"
  }

  stages {

    // -----------------------------------------------------------------------
    // Checkout：拉代码 + 初始化 submodule
    // -----------------------------------------------------------------------
    stage('Checkout') {
      steps {
        checkout scm
        sh 'git submodule update --init --recursive'
      }
    }

    // -----------------------------------------------------------------------
    // CI：并行跑各服务的 lint / test / build
    // 所有分支（PR / main / tag）都执行
    // -----------------------------------------------------------------------
    stage('CI') {
      when {
        expression { isCiTargetRef() }
      }
      parallel {

        stage('backend') {
          when {
            anyOf {
              changeset pattern: 'modelcraft-backend/**', comparator: 'GLOB'
              tag pattern: '*', comparator: 'GLOB'
            }
          }
          steps {
            dir('modelcraft-backend') {
              sh '''
                set -euo pipefail
                go mod tidy
                git diff --exit-code go.mod go.sum
                go mod verify
              '''
              sh 'just lint'
              sh 'just test-unit'
              sh 'just build'
            }
          }
          post {
            always {
              archiveArtifacts artifacts: 'modelcraft-backend/bin/**,modelcraft-backend/coverage.out',
                               allowEmptyArchive: true
            }
          }
        }

        stage('front') {
          when {
            anyOf {
              changeset pattern: 'modelcraft-front/**', comparator: 'GLOB'
              tag pattern: '*', comparator: 'GLOB'
            }
          }
          steps {
            dir('modelcraft-front') {
              sh 'npm ci'
              sh 'npm run lint'
              sh 'npm run test'
              sh 'npm run build'
            }
          }
          post {
            always {
              archiveArtifacts artifacts: 'modelcraft-front/.next/**',
                               allowEmptyArchive: true
            }
          }
        }

        stage('cli') {
          when {
            allOf {
              anyOf {
                changeset pattern: 'modelcraft-cli/**', comparator: 'GLOB'
                tag pattern: '*', comparator: 'GLOB'
              }
              expression { return false } // cli 目录非空时删除此行
            }
          }
          steps {
            echo 'cli 暂无构建内容，跳过'
          }
        }

      } // end parallel
    } // end CI

    // -----------------------------------------------------------------------
    // Deploy：仅 tag 触发
    //   1. 记录当前运行中的旧镜像 tag（回滚用）
    //   2. docker compose build（重新构建所有镜像）
    //   3. docker compose up -d（滚动替换容器）
    //   4. healthcheck（等待服务就绪）
    //   5. 失败时自动回滚到旧镜像
    // -----------------------------------------------------------------------
    stage('Deploy') {
      when {
        tag pattern: 'v.*', comparator: 'REGEXP'
      }
      environment {
        // 当前 tag 将作为镜像版本号注入 compose
        IMAGE_TAG = "${env.TAG_NAME}"
      }
      steps {
        script {
          // ── Step 1: 记录旧镜像，用于回滚 ──────────────────────────────────
          def prevBackendImg = sh(
            script: "docker inspect --format='{{.Config.Image}}' modelcraft-backend 2>/dev/null || echo ''",
            returnStdout: true
          ).trim()
          def prevApisixImg = sh(
            script: "docker inspect --format='{{.Config.Image}}' modelcraft-apisix 2>/dev/null || echo ''",
            returnStdout: true
          ).trim()
          env.PREV_BACKEND_IMG = prevBackendImg
          env.PREV_APISIX_IMG = prevApisixImg
          echo "旧版本 backend=${prevBackendImg ?: '(未运行)'} apisix=${prevApisixImg ?: '(未运行)'}"

          // ── Step 2: 构建新镜像 ─────────────────────────────────────────────
          sh "IMAGE_TAG=${env.IMAGE_TAG} docker compose build --no-cache backend"

          // ── Step 3: 启动/更新容器（基础设施服务如 MySQL/Redis 不重建）──────
          sh "IMAGE_TAG=${env.IMAGE_TAG} docker compose up -d --no-deps backend apisix"

          // ── Step 4: 健康检查（等待最多 120 秒）───────────────────────────
          def healthy = sh(
            script: '''
              set +e
              for i in $(seq 1 24); do
                BACKEND=$(docker inspect --format="{{.State.Health.Status}}" modelcraft-backend 2>/dev/null)
                APISIX=$(docker inspect --format="{{.State.Health.Status}}" modelcraft-apisix 2>/dev/null)
                echo "[$i/24] backend=${BACKEND} apisix=${APISIX}"
                if [ "$BACKEND" = "healthy" ] && [ "$APISIX" = "healthy" ]; then
                  exit 0
                fi
                sleep 5
              done
              exit 1
            ''',
            returnStatus: true
          )

          if (healthy != 0) {
            // ── Step 5: healthcheck 失败 → 回滚 ───────────────────────────
            echo "❌ Healthcheck 失败，开始回滚..."
            if (env.PREV_BACKEND_IMG) {
              sh "docker compose stop backend || true"
              sh "docker run -d --name modelcraft-backend-rollback --network modelcraft_modelcraft-network ${env.PREV_BACKEND_IMG} || true"
              sh "docker rename modelcraft-backend-rollback modelcraft-backend || true"
            }
            error("部署失败，已回滚到旧版本。请检查日志：docker logs modelcraft-backend / modelcraft-apisix")
          }

          echo "✅ 部署成功 — 版本: ${env.IMAGE_TAG}"
        }
      }
      post {
        always {
          // 清理悬空镜像，防止磁盘堆积
          sh 'docker image prune -f || true'
        }
      }
    } // end Deploy

  } // end stages

  post {
    success {
      echo "✅ Pipeline 通过 — ${env.BRANCH_NAME ?: env.TAG_NAME}"
    }
    failure {
      echo "❌ Pipeline 失败 — ${env.BRANCH_NAME ?: env.TAG_NAME}"
      // 按需开启通知：
      // slackSend channel: '#ci-alerts', color: 'danger',
      //           message: "❌ ${env.JOB_NAME} #${env.BUILD_NUMBER} 失败"
    }
  }
}
