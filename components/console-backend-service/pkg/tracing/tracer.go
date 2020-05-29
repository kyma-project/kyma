// Base on https://github.com/99designs/gqlgen-contrib/blob/master/gqlopentracing/tracer.go,
// contains the following modifications:
// 	- shortens the name of the span created in StartOperationExecution method
//	- full graphQL query is stored in graphQL.query span tag
package tracing

import (
	"context"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

type Tracer interface {
	graphql.HandlerExtension
	graphql.OperationInterceptor
	graphql.FieldInterceptor
}

// TODO: Disabled for now as it needs to be adapted to new gqlgen API

// New returns Tracer for OpenTracing.
// see https://opentracing.io/
func New() Tracer {
	t := tracer(0)
	return &t
}

type tracer int

func (t *tracer) ExtensionName() string {
	return "Open tracing"
}

func (t *tracer) Validate(schema graphql.ExecutableSchema) error {
	return nil
}

func (t *tracer) InterceptOperation(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	ctx = t.StartOperationExecution(ctx)
	rsp := next(ctx)
	t.EndOperationExecution(ctx)
	return rsp
}

func (t *tracer) InterceptField(ctx context.Context, next graphql.Resolver) (res interface{}, err error) {
	return nil, nil
}

func (t *tracer) getSpanName(rawQuery string) string {

	index := strings.Index(rawQuery, "(")
	if index < 1 {
		return rawQuery
	}

	return rawQuery[:index]
}

func (t *tracer) StartOperationExecution(ctx context.Context) context.Context {
	requestContext := graphql.GetRequestContext(ctx)
	spanName := t.getSpanName(requestContext.RawQuery)
	span, ctx := opentracing.StartSpanFromContext(ctx, spanName)
	span.SetTag("graphQL.query", requestContext.RawQuery)
	ext.SpanKind.Set(span, "server")
	ext.Component.Set(span, "graphQL")

	return ctx
}

func (tracer) StartFieldExecution(ctx context.Context, field graphql.CollectedField) context.Context {
	span, ctx := opentracing.StartSpanFromContext(ctx, "unnamed")
	ext.SpanKind.Set(span, "server")
	ext.Component.Set(span, "graphQL")

	return ctx
}

func (tracer) StartFieldResolverExecution(ctx context.Context, rc *graphql.ResolverContext) context.Context {
	span := opentracing.SpanFromContext(ctx)
	span.SetOperationName(rc.Object + "_" + rc.Field.Name)
	span.SetTag("resolver.object", rc.Object)
	span.SetTag("resolver.field", rc.Field.Name)

	return ctx
}

func (tracer) EndFieldExecution(ctx context.Context) {
	//span := opentracing.SpanFromContext(ctx)
	//defer span.Finish()
	//
	//rc := graphql.GetResolverContext(ctx)
	//reqCtx := graphql.GetRequestContext(ctx)
	//
	//errList := reqCtx.GetErrors(rc)
	//if len(errList) != 0 {
	//	ext.Error.Set(span, true)
	//	span.LogFields(
	//		log.String("event", "error"),
	//	)
	//
	//	for idx, err := range errList {
	//		span.LogFields(
	//			log.String(fmt.Sprintf("error.%d.message", idx), err.Error()),
	//			log.String(fmt.Sprintf("error.%d.kind", idx), fmt.Sprintf("%T", err)),
	//		)
	//	}
	//}
}

func (tracer) EndOperationExecution(ctx context.Context) {
	span := opentracing.SpanFromContext(ctx)
	defer span.Finish()
}
