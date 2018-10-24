from flask import Flask
from flask import request
from flask import jsonify
import threading
import os
import time
import argparse

app = Flask(__name__)


def getid(node):
    ret = "unknown"
    sl = node.split('~')
    if len(sl) > 2:
        ret = sl[2]
    return ret

# example
# /v1/listeners/istio-proxy/sidecar~10.32.1.20~httpbin-57db476f4d-svs9h.default~default.svc.cluster.local


@app.route('/v1/listeners/<cluster>/<node>', methods=['POST'])
def lds(cluster, node):
    op = insert_lua(request.get_json(), getid(node))
    return jsonify(op)

# example
# /v1/clusters/istio-proxy/sidecar~10.32.1.20~httpbin-57db476f4d-svs9h.default~default.svc.cluster.local


@app.route('/v1/clusters/<cluster>/<node>', methods=['POST'])
def cds(cluster, node):
    output = request.data
    return output

# example
# /v1/routes/15003/istio-proxy/sidecar~10.32.1.20~httpbin-57db476f4d-svs9h.default~default.svc.cluster.local


@app.route('/v1/routes/<name>/<cluster>/<node>', methods=['POST'])
def rds(name, cluster, node):
    output = request.data
    return output


"""
# example listener configuration
listeners:
- address: tcp://0.0.0.0:80
  bind_to_port: true
  filters:
  - name: http_connection_manager
    config:
      access_log:
      - path: /dev/stdout
      codec_type: auto
      filters:
      - name: mixer
        config: {}
      - name: lua
        config:
          inline_code: <code>

"""

#
# inserts lua as a filter in the http_connection_manager
#


def insert_lua(listeners, nodeid):
    for l in listeners.get("listeners", []):
        for f in l.get("filters", []):
            if f["name"] != "http_connection_manager":
                continue

            ff = f["config"].get("filters", [])
            ff.insert(0, lua_config(nodeid))

    return listeners


def lua_config(nodeid):
    s = FILE_STORE[SCRIPT]
    return {"name": "lua",
            "config":
            {"inline_code": s}}


DEFAULT_LUA_SCRIPT = """
-- Called on the request path.
function envoy_on_request(request_handle)
  request_handle:headers():add("x-lua-header", "true")
end

-- Called on the response path.
function envoy_on_response(response_handle)
  response_handle:headers():add("x-lua-resp-header", "{nodeid}")
end
"""

SCRIPT = "SCRIPT"
FILE_STORE = {
    SCRIPT: DEFAULT_LUA_SCRIPT
}


# polls for file chage
class poller(object):

    def __init__(self, filepath, cfg):
        self.filepath = filepath
        self.done = False
        self.cfg = cfg

    def cancel(self):
        self.done = True

    def __call__(self):
        modtime = 0
        while not self.done:
            modtime = self.read_if_changed(modtime)
            time.sleep(5)

    def read_if_changed(self, modtime):
        if not os.path.isfile(self.filepath):
            print self.filepath, "not found"
            os._exit(os.EX_OSFILE)

        new_modtime = os.path.getmtime(self.filepath)
        if new_modtime == modtime:
            return modtime
        
        with open(self.filepath, "rt") as fl:
            ls = fl.read()
            print "File updated"
            print ls
            self.cfg[SCRIPT] = ls
        return new_modtime


def get_args_parser():
    parser = argparse.ArgumentParser(
        description="Run pilot webhook")

    parser.add_argument("--script", help="path of the lua script to inject",
                        default="scripts/plugin.lua")
    parser.add_argument("--port", help="port to listen on",
                        type=int, default=5000)
    return parser


def main(args):
    p = poller(args.script, FILE_STORE)
    threading.Thread(target=p).start()
    ret = app.run(host="0.0.0.0", port=args.port)
    p.cancel()
    return ret


if __name__ == "__main__":
    import sys
    parser = get_args_parser()
    args = parser.parse_args()
    sys.exit(main(args))
