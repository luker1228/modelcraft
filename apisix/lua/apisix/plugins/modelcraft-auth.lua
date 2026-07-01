local core = require("apisix.core")
local jwt = require("resty.jwt")
local http = require("resty.http")
local cjson = require("cjson")

local plugin_name = "modelcraft-auth"

local schema = {
    type = "object",
    properties = {
        public_key = {
            type = "string",
            minLength = 1,
        },
        allow_pat = {
            type = "boolean",
            default = false,
        },
        whoami_url = {
            type = "string",
            minLength = 1,
        },
        inject_user_id = {
            type = "boolean",
            default = true,
        },
        inject_org_name = {
            type = "boolean",
            default = true,
        },
        inject_is_admin = {
            type = "boolean",
            default = true,
        },
    },
    required = {"public_key"},
}

local _M = {
    version = 0.1,
    priority = 2500,
    name = plugin_name,
    schema = schema,
}

local function unauthorized(message, request_id)
    local body = {message = message}
    if request_id then
        body.requestId = request_id
    end
    return core.response.exit(401, body)
end

local function inject_headers(conf, ctx, payload, auth_header)
    if conf.inject_user_id and payload.userId then
        core.request.set_header(ctx, "X-User-ID", payload.userId)
    elseif conf.inject_user_id and payload.user_id then
        core.request.set_header(ctx, "X-User-ID", payload.user_id)
    end

    if conf.inject_org_name and payload.orgName then
        core.request.set_header(ctx, "X-Org-Name", payload.orgName)
    elseif conf.inject_org_name and payload.org_name then
        core.request.set_header(ctx, "X-Org-Name", payload.org_name)
    end

    if conf.inject_is_admin then
        local is_admin = payload.isAdmin == true or payload.is_admin == true
        core.request.set_header(ctx, "X-Is-Admin", is_admin and "true" or "false")
    end

    if payload.accessToken and auth_header then
        core.request.set_header(ctx, "Authorization", "Bearer " .. payload.accessToken)
    end
end

local function decode_json(body)
    local ok, payload = pcall(cjson.decode, body)
    if not ok then
        return nil
    end
    return payload
end

local function verify_pat(conf, ctx, auth_header, request_id, uri)
    if not conf.whoami_url then
        core.log.error("[modelcraft-auth] allow_pat=true but whoami_url is empty")
        return unauthorized("PAT validation is not configured", request_id)
    end

    core.log.warn("[modelcraft-auth] validating PAT request_id=", request_id, " uri=", uri)

    local httpc = http.new()
    httpc:set_timeout(3000)

    local res, err = httpc:request_uri(conf.whoami_url, {
        method = "GET",
        headers = {
            ["Authorization"] = auth_header,
            ["X-Request-Id"] = request_id,
        },
        keepalive = false,
    })

    if err or not res or res.status ~= 200 then
        local status = res and res.status or "nil"
        local body = res and res.body or ""
        if #body > 300 then
            body = body:sub(1, 300)
        end

        core.log.error(
            "[modelcraft-auth] PAT whoami failed request_id=", request_id,
            " err=", err or "nil",
            " status=", status,
            " body=", body
        )

        local failed_request_id = request_id
        local message = "Invalid or expired PAT token"
        local payload = res and res.body and decode_json(res.body) or nil
        if payload then
            if payload.requestId then
                failed_request_id = payload.requestId
            end
            if payload.error and payload.error.message then
                message = payload.error.message
            end
        end

        return unauthorized(message, failed_request_id)
    end

    local payload = decode_json(res.body or "")
    if not payload or not payload.userId then
        local body = res and res.body or ""
        if #body > 300 then
            body = body:sub(1, 300)
        end

        core.log.error(
            "[modelcraft-auth] PAT whoami missing userId request_id=", request_id,
            " status=", res and res.status or "nil",
            " body=", body
        )

        return unauthorized("PAT validation failed", request_id)
    end

    core.log.info(
        "[modelcraft-auth] PAT ok request_id=", request_id,
        " user_id=", payload.userId,
        " org_name=", payload.orgName or "",
        " is_admin=", tostring(payload.isAdmin == true)
    )

    inject_headers(conf, ctx, payload, auth_header)
    return
end

local function verify_jwt(conf, token)
    local jwt_obj = jwt:verify(conf.public_key, token)
    if not jwt_obj or not jwt_obj.verified then
        local message = (jwt_obj and jwt_obj.reason == "jwt expired") and "Token expired" or "Invalid token"
        return nil, message
    end

    if not jwt_obj.payload then
        return nil, "Invalid token payload"
    end

    return jwt_obj.payload
end

function _M.check_schema(conf)
    return core.schema.check(schema, conf)
end

function _M.rewrite(conf, ctx)
    local auth_header = core.request.header(ctx, "Authorization")
    local request_id = core.request.header(ctx, "X-Request-Id")
    if not auth_header then
        return unauthorized("Authorization header required", request_id)
    end

    local token = auth_header:match("Bearer%s+(.+)")
    if not token then
        return unauthorized("Bearer token required", request_id)
    end

    if conf.allow_pat and token:sub(1, 7) == "mc_pat_" then
        return verify_pat(conf, ctx, auth_header, request_id, ctx.var.uri)
    end

    local payload, err_message = verify_jwt(conf, token)
    if not payload then
        return unauthorized(err_message, request_id)
    end

    inject_headers(conf, ctx, payload, nil)
end

return _M
