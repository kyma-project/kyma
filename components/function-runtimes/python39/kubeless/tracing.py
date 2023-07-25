from contextlib import contextmanager
from typing import Iterator

from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider, _Span
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
from opentelemetry.propagate import extract
from opentelemetry.sdk.resources import (
    SERVICE_NAME,
    Resource
)
from opentelemetry.sdk.trace.export import (
    SimpleSpanProcessor,
)
from opentelemetry.sdk.trace.sampling import ( 
    DEFAULT_ON,
)

from opentelemetry.trace import context_api
from opentelemetry.trace.propagation import _SPAN_KEY

import os

# Tracing propagators are configured based on OTEL_PROPAGATORS env variable set in dockerfile
# https://opentelemetry.io/docs/instrumentation/python/manual/#using-environment-variables
def _setup_tracer(service_name: str) -> trace.Tracer:

    provider = TracerProvider(
        resource=Resource.create({SERVICE_NAME: service_name}),
        sampler=DEFAULT_ON,
    )
   
    tracecollector_endpoint = os.getenv('TRACE_COLLECTOR_ENDPOINT')

    if tracecollector_endpoint:
        span_processor = SimpleSpanProcessor(OTLPSpanExporter(endpoint=tracecollector_endpoint))
        provider.add_span_processor(span_processor)
 
    # Sets the global default tracer provider
    trace.set_tracer_provider(provider)

    # Creates a tracer from the global tracer provider
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
        