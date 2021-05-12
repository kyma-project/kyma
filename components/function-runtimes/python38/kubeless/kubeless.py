#!/usr/bin/env python

import importlib
import io
import os
import sys
import threading
import bottle
import prometheus_client as prom
import queue

# The reason this file has an underscore prefix in its name is to avoid a
# name collision with the user-defined module.
module_name = os.getenv('MOD_NAME')
if module_name is None:
    print('MOD_NAME have to be provided', flush=True)
    exit(1)
current_mod = os.path.basename(__file__).split('.')[0]
if module_name == current_mod:
    print('Module cannot be named {} as current module'.format(current_mod), flush=True)
    exit(2)

sys.path.append('/kubeless')

mod = importlib.import_module(module_name)
func_name = os.getenv('FUNC_HANDLER')
if func_name is None:
    print('FUNC_HANDLER have to be provided', flush=True)
    exit(3)

func = getattr(mod, os.getenv('FUNC_HANDLER'))

func_port = os.getenv('FUNC_PORT', 8080)
timeout = float(os.getenv('FUNC_TIMEOUT', 180))
memfile_max = int(os.getenv('FUNC_MEMFILE_MAX', 100 * 1024 * 1024))
bottle.BaseRequest.MEMFILE_MAX = memfile_max

app = application = bottle.app()

function_context = {
    'function-name': func.__name__,
    'timeout': timeout,
    'runtime': os.getenv('FUNC_RUNTIME'),
    'memory-limit': os.getenv('FUNC_MEMORY_LIMIT'),
}


class PicklableBottleRequest(bottle.BaseRequest):
    '''Bottle request that can be pickled (serialized).

    `bottle.BaseRequest` is not picklable and therefore cannot be passed directly to a
    python multiprocessing `Process` when using the forkserver or spawn multiprocessing
    contexts. So, we selectively delete components that are not picklable.
    '''

    def __init__(self, data, *args, **kwargs):
        super().__init__(*args, **kwargs)
        # Bottle uses either `io.BytesIO` or `tempfile.TemporaryFile` to store the
        # request body depending on whether the length of the body is less than
        # `MEMFILE_MAX` or not, but `tempfile.TemporaryFile` is not picklable.
        # So, we override it to always store the body as `io.BytesIO`.
        self.environ['bottle.request.body'] = io.BytesIO(data)

    def __getstate__(self):
        env = self.environ.copy()

        # File-like objects are not picklable.
        del env['wsgi.errors']
        del env['wsgi.input']

        # bottle.ConfigDict is not picklable because it contains a lambda function.
        del env['bottle.app']
        del env['bottle.route']
        del env['route.handle']

        return env

    def __setstate__(self, env):
        setattr(self, 'environ', env)


@app.get('/healthz')
def healthz():
    return 'OK'


@app.get('/metrics')
def metrics():
    bottle.response.content_type = prom.CONTENT_TYPE_LATEST
    return prom.generate_latest(prom.REGISTRY)


@app.error(500)
def exception_handler():
    return 'Internal server error'


@app.route('/<:re:.*>', method=['GET', 'POST', 'PATCH', 'DELETE'])
def handler():
    req = bottle.request
    data = req.body.read()
    picklable_req = PicklableBottleRequest(data, req.environ.copy())
    if req.get_header('content-type') == 'application/json':
        data = req.json

    event = {
        'data': data,
        'ce-type': req.get_header('ce-type'),
        'ce-source': req.get_header('ce-source'),
        'ce-eventtypeversion': req.get_header('ce-eventtypeversion'),
        'ce-specversion': req.get_header('ce-specversion'),
        'ce-id': req.get_header('ce-id'),
        'ce-time': req.get_header('ce-time'),
        'extensions': {'request': picklable_req}
    }

    method = req.method
    func_calls.labels(method).inc()
    with func_errors.labels(method).count_exceptions():
        with func_hist.labels(method).time():
            que = queue.Queue()
            t = threading.Thread(target=lambda q, e: q.put(func(e,function_context)), args=(que,event))
            t.start()
            try:
                res = que.get(block=True, timeout=timeout)
            except queue.Empty:
                return bottle.HTTPError(408, "Timeout while processing the function")
            else:
                t.join()
                return res

def preload():
    """This is a no-op function used to start the forkserver."""
    pass


if __name__ == '__main__':
    import logging
    import multiprocessing as mp
    import requestlogger

    mp_context = os.getenv('MP_CONTEXT', 'forkserver')

    if mp_context == "fork":
        raise ValueError(
            '"fork" multiprocessing context is not supported because cherrypy is a '
            'multithreaded server and safely forking a multithreaded process is '
            'problematic'
        )
    if mp_context not in ["forkserver", "spawn"]:
        raise ValueError(
            f'"{mp_context}" is an invalid multiprocessing context. Possible values '
            'are "forkserver" and "spawn"'
        )

    try:
        ctx = mp.get_context(mp_context)

        if ctx.get_start_method() == 'forkserver':
            # Preload the current module and consequently also the user-defined module
            # so that all the child processes forked from the forkserver in response to
            # a request immediately have access to the global data in the user-defined
            # module without having to load it for every request.
            ctx.set_forkserver_preload([current_mod])

            # Start the forkserver before we start accepting requests.
            d = ctx.Process(target=preload)
            d.start()
            d.join()

    except ValueError:
        # Default to 'spawn' if 'forkserver' is unavailable.
        ctx = mp.get_context('spawn')
        logging.warn(
            f'"{mp_context}" multiprocessing context is unavailable. Using "spawn"'
        )

    func_hist = prom.Histogram(
        'function_duration_seconds', 'Duration of user function in seconds', ['method']
    )
    func_calls = prom.Counter(
        'function_calls_total', 'Number of calls to user function', ['method']
    )
    func_errors = prom.Counter(
        'function_failures_total', 'Number of exceptions in user function', ['method']
    )

    # added by Kyma team
    if os.getenv('KYMA_INTERNAL_LOGGER_ENABLED'):
        # default that has been used so far
        loggedapp = requestlogger.WSGILogger(
            app,
            [logging.StreamHandler(stream=sys.stdout)],
            requestlogger.ApacheFormatter(),
        )
    else:
        loggedapp = app
    # end of modified section

    bottle.run(
        loggedapp,
        server='cherrypy',
        host='0.0.0.0',
        port=func_port,
        # Set this flag to True to auto-reload the server after any source files change
        reloader=os.getenv('CHERRYPY_RELOADED', False),
        # Number of requests that can be handled in parallel (default = 10).
        numthreads=int(os.getenv('CHERRYPY_NUMTHREADS', 10)),
        quiet='KYMA_BOTTLE_QUIET_OPTION_DISABLED' not in os.environ,
    )
