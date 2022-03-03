import os
import requests
from opentelemetry import trace
from opentelemetry.trace import context_api
from opentelemetry.exporter.jaeger.thrift import JaegerExporter
from opentelemetry.sdk.resources import SERVICE_NAME, Resource
from opentelemetry.instrumentation.requests import RequestsInstrumentor
from opentelemetry.sdk.trace import TracerProvider, _Span
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.propagate import set_global_textmap
from opentelemetry.propagators.b3 import B3MultiFormat
from contextlib import contextmanager
from opentelemetry.propagate import extract
from opentelemetry.trace.propagation import _SPAN_KEY
from typing import Iterator

podName = os.getenv('HOSTNAME')
serviceNamespace = os.getenv('SERVICE_NAMESPACE')
jaegerEndpoint = os.getenv('JAEGER_SERVICE_ENDPOINT')

def is_jaeger_available() -> bool:
    try:
        res = requests.get(jaegerEndpoint, timeout=2)
    
        # 405 is the right status code for the GET method if jaeger service exists 
        # because the only allowed method is POST and usage of other methods are not allowe
        # https://github.com/jaegertracing/jaeger/blob/7872d1b07439c3f2d316065b1fd53e885b26a66f/cmd/collector/app/handler/http_handler.go#L60
        if res.status_code == 405:
            return True
    except:
        pass

    return False

def get_tracer() -> trace.Tracer:
    set_global_textmap(B3MultiFormat())
    
    RequestsInstrumentor().instrument()

    # remove generated pods suffix ( two last sections )
    deploymentName = '-'.join(podName.split('-')[0:podName.count('-')-1])

    trace.set_tracer_provider(
        TracerProvider(
            resource=Resource.create({SERVICE_NAME: '.'.join([deploymentName, serviceNamespace])})
        )
    )

    jaeger_exporter = JaegerExporter(
        collector_endpoint = jaegerEndpoint + '?format=jaeger.thrift',
    )

    span_processor = BatchSpanProcessor(jaeger_exporter)

    trace.get_tracer_provider().add_span_processor(span_processor)

    return trace.get_tracer(__name__)

@contextmanager  # type: ignore
def set_req_context(req) -> Iterator[trace.Span]:
    '''Propagates incoming span from the request to the current context

    This method allows to set up a context in any thread based on the incoming request.
    By design, span context can't be moved between threads and because we run every function 
    in the separated thread we have to propagate the context manually.
    '''
    span = _Span(
        "request-span", 
        trace.get_current_span(
            extract(req.headers)
        ).get_span_context()
    )

    token = context_api.attach(context_api.set_value(_SPAN_KEY, span))
    try:
        yield span
    finally:
        context_api.detach(token)
