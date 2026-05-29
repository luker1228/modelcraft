# Errors

Command failures and integration errors.

---

## [ERR-20260511-001] docker-compose-detection

**Logged**: 2026-05-11T00:00:00+08:00
**Priority**: medium
**Status**: resolved
**Area**: infra

### Summary
当前环境不支持 `docker compose -f ...` 子命令，但提供了独立的 `docker-compose` 可执行文件。

### Error
```
unknown shorthand flag: 'f' in -f
```

### Context
- Command attempted: `docker compose -f deploy/compose/docker-compose.local.yml config`
- Environment had `docker-compose version v5.1.0`
- Root cause: 当前机器没有可用的 Docker Compose plugin，`docker compose` 路径不可用

### Suggested Fix
在根目录 `justfile` 中添加 compose 命令探测：优先 `docker compose`，不可用时回退到 `docker-compose`。

### Metadata
- Reproducible: yes
- Related Files: justfile

### Resolution
- **Resolved**: 2026-05-11T00:00:00+08:00
- **Commit/PR**: local workspace change
- **Notes**: 根目录 `justfile` 已加入 compose 命令探测与回退逻辑。

---

## [ERR-20260525-001] action-status-check

**Logged**: 2026-05-25T00:00:00Z
**Priority**: low
**Status**: pending
**Area**: infra

### Summary
本地环境无 `gh` CLI 且 GitHub REST API 匿名请求触发 rate limit，导致无法直接用 API 拉取 workflow run。

### Error
```
/bin/bash: line 1: gh: command not found

{"message":"API rate limit exceeded ..."}
```

### Context
- Command attempted: `gh run list --repo patientCat/modelcraft --workflow "Release CLI" --limit 5`
- Fallback API: `https://api.github.com/repos/patientCat/modelcraft/actions/runs?event=push&per_page=20`
- Both unavailable paths blocked run status check in this environment.

### Suggested Fix
优先使用 Web 页面抓取（`/actions/workflows/release-cli.yml`）确认 run，再用 release 下载 URL 的 HEAD 请求验证资产可访问。

### Metadata
- Reproducible: unknown
- Related Files: .github/workflows/release-cli.yml

---

## [ERR-20260529-001] brainstorming-visual-server

**Logged**: 2026-05-29T05:33:21Z
**Priority**: medium
**Status**: pending
**Area**: docs

### Summary
visual companion 的 `start-server.sh` 在当前仓库中用相对 `--project-dir .` 启动会失败，因为脚本内部会切换到脚本目录后再写 session 文件。

### Error
```
.../start-server.sh: line 119: ./.superpowers/brainstorm/.../state/server.log: No such file or directory
{"error": "Server failed to start within 5 seconds"}
```

### Context
- Command attempted: `start-server.sh --project-dir .`
- Root cause: 脚本在解析参数后 `cd "$SCRIPT_DIR"`，导致相对 `PROJECT_DIR` 不再指向仓库根目录。
- Successful workaround: 改为传入绝对路径形式 `--project-dir "$PWD"`。

### Suggested Fix
以后在这个 visual companion 脚本中一律传绝对 `--project-dir`，或在脚本中提前把 `PROJECT_DIR` 规范化为绝对路径。

### Metadata
- Reproducible: yes
- Related Files: .codebuddy/plugins/marketplaces/superpowers-marketplace/plugins/superpowers/skills/brainstorming/scripts/start-server.sh

---
