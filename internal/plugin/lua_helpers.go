package plugin

// luaJSONHelper defines json.encode in Lua so plugin tables never need to be
// walked from Go. json.decode and __go_encode are Go-backed globals registered
// before this chunk runs. Keys are sorted to keep encoded objects deterministic.
const luaJSONHelper = `
local json = _G.json or {}
_G.json = json

local function escape(s)
  s = tostring(s)
  s = s:gsub("\\", "\\\\"):gsub('"', '\\"'):gsub("\b", "\\b"):gsub("\f", "\\f")
  s = s:gsub("\n", "\\n"):gsub("\r", "\\r"):gsub("\t", "\\t")
  s = s:gsub("[%c]", function(c) return string.format("\\u%04x", c:byte()) end)
  return s
end

local function encode(value, depth)
  if depth > 32 then error("json: nesting too deep") end
  local kind = type(value)
  if kind == "nil" then return "null" end
  if kind == "boolean" then return value and "true" or "false" end
  if kind == "number" then
    if value ~= value or value == math.huge or value == -math.huge then return "null" end
    local text = tostring(value)
    if text:sub(-2) == ".0" then text = text:sub(1, -3) end
    return text
  end
  if kind == "string" then return '"' .. escape(value) .. '"' end
  if kind ~= "table" then error("json: unsupported type " .. kind) end

  local count, arrayLen = 0, 0
  for k in pairs(value) do
    count = count + 1
    if type(k) == "number" and k == arrayLen + 1 then arrayLen = k end
  end
  if count > 0 and arrayLen == count then
    local parts = {}
    for i = 1, arrayLen do parts[i] = encode(value[i], depth + 1) end
    return "[" .. table.concat(parts, ",") .. "]"
  end

  local parts = {}
  for k, v in pairs(value) do
    if type(k) == "string" or type(k) == "number" then
      parts[#parts + 1] = '"' .. escape(k) .. '":' .. encode(v, depth + 1)
    end
  end
  table.sort(parts)
  return "{" .. table.concat(parts, ",") .. "}"
end

function json.encode(value) return encode(value, 0) end
__go_encode = json.encode

local goRequire = require
function require(name)
  if name == "json" then return json end
  local encoded = goRequire(name)
  if type(encoded) ~= "string" then return nil end
  return json.decode(encoded)
end
`

// luaHostHelper builds typed wrappers over the generic host.call envelope so a
// Lua plugin can use the same host.* vocabulary as the other runtimes.
const luaHostHelper = `
local function call(name, ...)
  local args = { ... }
  local encoded = #args == 0 and "[]" or json.encode(args)
  local raw = host.call(name, encoded)
  local reply = json.decode(raw)
  if type(reply) ~= "table" or not reply.ok then
    local why = type(reply) == "table" and reply.error or "host call failed"
    return nil, why
  end
  return reply.data
end

host.now = function() return call("now") end

host.log = {}
host.log.info = function(msg) call("log.info", msg) end
host.log.error = function(msg) call("log.error", msg) end
host.log.debug = function(msg) call("log.debug", msg) end

host.storage = {}
host.storage.get = function(key) return call("storage.get", key) end
host.storage.set = function(key, value) return call("storage.set", key, value) end

host.text = host.text or {}
host.text.clean = function(s) return call("text.clean", s) end
host.text.strip_html = function(s) return call("text.strip_html", s) end
host.text.trim = function(s) return call("text.trim", s) end
host.text.chapter_num = function(s) return call("text.chapter_num", s) end
host.text.date_epoch = function(s) return call("text.date_epoch", s) end
host.text.date_to_iso = function(s) return call("text.date_to_iso", s) end
host.text.normalize_status = function(s) return call("text.normalize_status", s) end
host.text.json_blob = function(s, marker) return call("text.json_blob", s, marker) end
host.text.unescape = function(s) return call("text.unescape", s) end

host.plugin = { param = function(key) return call("plugin.param", key) end }
host.config = { get = function(key) return call("config.get", key) end }
host.fs = {
  read = function(name) return call("fs.read", name) end,
  read_lib = function(name) return call("fs.read_lib", name) end,
}

host.codecs = {}
host.codecs.base64_encode = function(s) return call("codecs.base64_encode", s) end
host.codecs.base64_decode = function(s) return call("codecs.base64_decode", s) end
host.codecs.hex_encode = function(s) return call("codecs.hex_encode", s) end
host.codecs.hex_decode = function(s) return call("codecs.hex_decode", s) end

host.html = {}
host.html.select = function(doc, selector) return call("html.select", doc, selector) end
host.html.attr = function(doc, selector, attr) return call("html.attr", doc, selector, attr) end
host.html.xpath = function(expr, doc) return call("html.xpath", expr, doc) end

host.crypto = {
  vrf_sign = function(path, params, stages) return call("crypto.vrf_sign", path, params, stages) end,
  sha256_hex = function(s) return call("crypto.sha256_hex", s) end,
  md5_hex = function(s) return call("crypto.md5_hex", s) end,
  hmac_sha256_hex = function(data, key) return call("crypto.hmac_sha256_hex", data, key) end,
  aes_gcm_decrypt = function(iv, ct, tag, key) return call("crypto.aes_gcm_decrypt", iv, ct, tag, key) end,
}

host.browser = { call = function(tool, args) return call("browser.call", tool, args) end }

host.url_encode = function(s) return call("url_encode", s) end
host.url_decode = function(s) return call("url_decode", s) end

host.http = {}
local fetchTools = { "fetch", "fetch_body", "get_json", "post_json" }
for _, name in ipairs(fetchTools) do
  host.http[name] = function(url, opts)
    if opts == nil then return call("http." .. name, url) end
    return call("http." .. name, url, opts)
  end
end

io = {
  read = function(name) return call("fs.read", name) end,
  lines = function(name)
    local body, err = call("fs.read", name)
    if body == nil then return nil, err end
    local pos = 1
    return function()
      if pos > #body then return nil end
      local nl = string.find(body, "\n", pos, true)
      local line
      if nl then
        line = string.sub(body, pos, nl - 1)
        pos = nl + 1
      else
        line = string.sub(body, pos)
        pos = #body + 1
      end
      return line
    end
  end,
}

local re2_call = function(pattern, subject) return call("re2.match", pattern, subject) end
host.re2 = setmetatable({
  match = function(p, s) return call("re2.match", p, s) end,
  find = function(p, s) return call("re2.find", p, s) end,
  find_all = function(p, s) return call("re2.find_all", p, s) end,
  gmatch = function(p, s) return call("re2.gmatch", p, s) end,
  replace = function(p, r, s) return call("re2.replace", p, r, s) end,
  quote = function(s) return call("re2.quote", s) end,
  find_index = function(p, s) return call("re2.find_index", p, s) end,
  capture = function(p, s) return call("re2.capture", p, s) end,
}, { __call = function(_, p, s) return re2_call(p, s) end })
`

// luaInvokeHelper converts a handle result to the frozen JSON envelope while
// preserving plugins that already return a JSON string.
const luaInvokeHelper = `
local pluginHandle = handle
if type(pluginHandle) ~= "function" then error("plugin: lua: missing handle function") end
function __invoke(request)
  local result = pluginHandle(request)
  if type(result) == "string" then return result end
  return json.encode(result)
end
`
