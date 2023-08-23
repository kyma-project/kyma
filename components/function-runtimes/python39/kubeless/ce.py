import io
import json
import os

import bottle
import requests
from cloudevents.http import from_http

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


def resolve_data_type(event_data):
    if type(event_data) is dict:
        return 'application/json'
    elif type(event_data) is str:
        return 'text/plain'


def build_cloud_event_attributes(req, data):
    event = from_http(req.headers, data)
    ceHeaders = {
        'data': event.data,
        # 'ce-type': event['type'],
        'ce-source': event['source'],
        # 'ce-eventtypeversion': event['eventtypeversion'],
        'ce-specversion': event['specversion'],
        'ce-id': event['id'],
        'ce-time': event['time'],
    }
    return ceHeaders


def has_ce_headers(headers):
    has = 'ce-type' in headers and 'ce-source' in headers
    return has


def is_cloud_event(req):
    return req.get_header('content-type') == 'application/cloudevents+json' or has_ce_headers(req.headers)


class Event:
    def __init__(self, req, tracer):
        self.ceHeaders = dict()
        self.tracer = tracer
        self.req = req
        data = req.body.read()
        picklable_req = PicklableBottleRequest(data, req.environ.copy())
        self.ceHeaders.update({
            'extensions': {'request': picklable_req}
        })

        if is_cloud_event(req):
            ce_headers = build_cloud_event_attributes(req, data)
            self.ceHeaders.update(ce_headers)
        else:
            if req.get_header('content-type') == 'application/json':
                data = req.json
                self.ceHeaders.update({'data': data})

    def __getitem__(self, item):
        return self.ceHeaders[item]

    def __setitem__(self, name, value):
        self.ceHeaders[name] = value

    def publishCloudEvent(self, data):
        return requests.post(
            publisher_proxy_address,
            data=json.dumps(data, default=str),
            headers={"Content-Type": "application/cloudevents+json"}
        )

    def build_response_cloud_event(self, event_id, event_type, event_data):
        return {
            'type': event_type,
            'source': self.ceHeaders['ce-source'],
            'eventtypeversion': self.ceHeaders['ce-eventtypeversion'],
            'specversion': self.ceHeaders['ce-specversion'],
            'id': event_id,
            'data': event_data,
            'datacontenttype': resolve_data_type(event_data)
        }
