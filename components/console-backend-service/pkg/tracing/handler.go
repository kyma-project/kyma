package tracing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"runtime"

	"github.com/golang/glog"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

type OpentracingHandler http.HandlerFunc

// type GQLBody struct {
// 	Query string
// }

type GQLBody struct {
	Query string
}

const (
	// InfoColor    = "\033[1;34m%s\033[0m"
	InfoColor    = "[1;34m"
	NoticeColor  = "[1;36m"
	WarningColor = "[1;33m"
	ErrorColor   = "[1;31m"
	DebugColor   = "[0;36m"
)

func printWithColor(color string, message string) {
	fmt.Printf("\033%s%s\033[0m", color, message)
}

func getColorForNumber(number uint64) string {
	if number > 0 {
		return ErrorColor
	}
	return "[0m"
}

func PrintMemUsage(queryName string, statsBefore runtime.MemStats, statsAfter runtime.MemStats) {
	alloc := bToMb(statsAfter.Alloc - statsBefore.Alloc)
	totalAlloc := bToMb(statsAfter.TotalAlloc - statsBefore.TotalAlloc)
	sys := bToMb(statsAfter.Sys - statsBefore.Sys)
	numGC := statsAfter.NumGC - statsBefore.NumGC

	printWithColor(WarningColor, queryName)
	fmt.Printf(" consumed: ")
	printWithColor(getColorForNumber(alloc), fmt.Sprintf("\tAlloc = %v MiB", alloc))
	printWithColor(getColorForNumber(totalAlloc), fmt.Sprintf("\tTotalAlloc = %v MiB", totalAlloc))
	printWithColor(getColorForNumber(sys), fmt.Sprintf("\tSys = %v MiB", sys))
	printWithColor(getColorForNumber(uint64(numGC)), fmt.Sprintf("\tNumGC = %v\n", numGC))

}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func NewWithParentSpan(spanName string, next http.HandlerFunc) OpentracingHandler {
	r, _ := regexp.Compile(`query (.*?)(\(| |\{})`)

	return func(writer http.ResponseWriter, request *http.Request) {

		queryName := "<unrecognized query>"
		var statsBefore runtime.MemStats
		runtime.ReadMemStats(&statsBefore)

		bodyCopy, err := ioutil.ReadAll(request.Body)
		request.Body = ioutil.NopCloser(bytes.NewReader(bodyCopy))

		if request.Method == "POST" && bodyCopy != nil {
			if err != nil {
				log.Fatal(err)
			}

			var body GQLBody
			json.Unmarshal([]byte(bodyCopy), &body)

			queryNameArray := r.FindStringSubmatch(body.Query)

			if len(queryNameArray) >= 2 {
				queryName = queryNameArray[1]
			}

		}

		spanContext, err := opentracing.GlobalTracer().Extract(
			opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(request.Header))
		if err != nil {
			glog.Warning("opentracing parent span headers extract", err)
			next(writer, request)
		}
		span := opentracing.StartSpan(spanName,
			opentracing.ChildOf(spanContext))
		defer span.Finish()

		ext.SpanKind.Set(span, "server")
		ext.Component.Set(span, spanName)
		ctx := opentracing.ContextWithSpan(request.Context(), span)
		next(writer, request.WithContext(ctx))

		var statsAfter runtime.MemStats
		runtime.ReadMemStats(&statsAfter)

		defer PrintMemUsage(queryName, statsBefore, statsAfter)
	}
}
