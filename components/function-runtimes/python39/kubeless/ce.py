import requests
import bottle
import io
import os
import json

publisher_proxy_address = os.getenv('PUBLISHER_PROXY_ADDRESS')

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


class Event:
    ceHeaders = dict()
    tracer = None

    def __init__(self, req, tracer):
        data = req.body.read()
        picklable_req = PicklableBottleRequest(data, req.environ.copy())
        if req.get_header('content-type') == 'application/json':
            data = req.json

        self.req = req
        self.tracer = tracer
        self.ceHeaders = {
            'data': data,
            'ce-type': req.get_header('ce-type'),
            'ce-source': req.get_header('ce-source'),
            'ce-eventtypeversion': req.get_header('ce-eventtypeversion'),
            'ce-specversion': req.get_header('ce-specversion'),
            'ce-id': req.get_header('ce-id'),
            'ce-time': req.get_header('ce-time'),
            'extensions': {'request': picklable_req}
        }

    def __getitem__(self, item):
        return self.ceHeaders[item]

    def __setitem__(self, name, value):
        self.ceHeaders[name] = value

    def publishCloudEvent(self, data):
        return requests.post(
            publisher_proxy_address,
            data = json.dumps(data),
            headers = {"Content-Type": "application/cloudevents+json"}
            )
    
    def resolveDataType(self, event_data):
        if type(event_data) is dict:
            return 'application/json'
        elif type(event_data) is str:
            return 'text/plain'

    def buildResponseCloudEvent(self, event_id, event_type, event_data):
        return {
            'type': event_type,
            'source': self.ceHeaders['ce-source'],
            'eventtypeversion': self.ceHeaders['ce-eventtypeversion'],
            'specversion': self.ceHeaders['ce-specversion'],
            'id': event_id,
            'data': event_data,
            'datacontenttype': self.resolveDataType(event_data)
        }

