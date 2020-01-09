package tracing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

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
	InfoColor         = "[1;34m"
	NoticeColor       = "[1;36m"
	WarningColor      = "[1;33m"
	ErrorColor        = "[1;31m"
	DebugColor        = "[0;36m"
	NotImportantColor = "[1;90m"
	GoodColor         = "[1;92m"
)

func printWithColor(color string, message string) {
	fmt.Printf("\033%s%s\033[0m", color, message)
}

func getColorForNumber(number int64) string {
	if number > 0 {
		return ErrorColor
	}

	if number < 0 {
		return GoodColor
	}
	return "[0m"
}

const unrecognizedQuery = "<unrecognized query>"

func PrintMemUsage(queryName string, statsBefore runtime.MemStats, statsAfter runtime.MemStats) {
	alloc := bToMb(int64(statsAfter.Alloc - statsBefore.Alloc))
	totalAlloc := bToMb(int64(statsAfter.TotalAlloc - statsBefore.TotalAlloc))
	allocAfter := bToMb(int64(statsAfter.Alloc))
	numGC := int64(statsAfter.NumGC - statsBefore.NumGC)

	queryColor := InfoColor
	if queryName == unrecognizedQuery {
		queryColor = NotImportantColor
	}

	printWithColor(NotImportantColor, time.Now().Format(time.RFC850))
	printWithColor(queryColor, fmt.Sprintf("%-30v", queryName))
	fmt.Printf(" consumed: ")
	printWithColor(getColorForNumber(alloc), fmt.Sprintf("Alloc = %v MiB", alloc))
	printWithColor(getColorForNumber(totalAlloc), fmt.Sprintf("\tTotalAlloc = %v MiB", totalAlloc))
	printWithColor(getColorForNumber(numGC), fmt.Sprintf("\tNumGC = %v", numGC))
	printWithColor(string(allocAfter), fmt.Sprintf("\tAlloc after query = %v MiB\n", allocAfter))

}

func bToMb(b int64) int64 {
	return b / 1024 / 1024
}

func NewWithParentSpan(spanName string, next http.HandlerFunc) OpentracingHandler {
	r, _ := regexp.Compile(`query (.*?)(\(| |\{})`)

	return func(writer http.ResponseWriter, request *http.Request) {

		queryName := unrecognizedQuery
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
