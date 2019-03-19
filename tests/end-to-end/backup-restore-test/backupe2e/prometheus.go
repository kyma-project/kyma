package backupe2e

// Test http request to the prometheus api before backup and after restore.
// An expected result at a specific point in time is considered a succeeded test.
//
// {
//   "status":"success",
//   "data":{
//      "resultType":"vector",
//      "result":[
//         {
//           "metric":{},
//           "value":[
//               1551421406.195,
//               "1.661"
//            ]
//          }
//      ]
//    }
// }
//
// {success {vector [{map[] [1.551424874014e+09 1.661]}]}}

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
)

func init() {
	prometheusTest, err := NewPrometheusTest()
	if err != nil {
		log.Fatal(err)
	}
	Register(prometheusTest)
}

const (
	domain       = "http://monitoring-prometheus.kyma-system"
	prometheusNS = "kyma-system"
	api          = "/api/v1/query?"
	metricsQuery = "max(sum(kube_pod_container_resource_requests_cpu_cores) by (instance))"
	port         = "9090"
	metricName   = "kube_pod_container_resource_requests_cpu_cores"
)

type queryResponse struct {
	Status string       `json:"status"`
	Data   responseData `json:"data"`
}

type responseData struct {
	ResultType string       `json:"resultType"`
	Result     []dataResult `json:"result"`
}

type dataResult struct {
	Metric interface{}   `json:"metric,omitempty"`
	Value  []interface{} `json:"value,omitempty"`
}

type prometheusTest struct {
	metricName, uuid string
	beforeBackup     queryResponse
	expectedResult   string
	finalResult      string
	apiQuery
	pointInTime
}

type pointInTime struct {
	floatValue      float64
	timeValue       time.Time
	formmattedValue string
}

type apiQuery struct {
	domain       string
	prometheusNS string
	api          string
	metricQuery  string
	port         string
}

func NewPrometheusTest() (*prometheusTest, error) {

	queryToApi := apiQuery{api: api, domain: domain, metricQuery: metricsQuery, port: port, prometheusNS: prometheusNS}

	return &prometheusTest{
		metricName: metricName,
		uuid:       uuid.New().String(),
		apiQuery:   queryToApi,
	}, nil
}

func (point *pointInTime) pointInTime(f float64) {
	point.floatValue = f

	t := time.Unix(int64(f), 0) //gives unix time stamp in utc
	point.timeValue = t

	point.formmattedValue = t.Format(time.RFC3339)
}

type Connector interface {
	connectToPrometheusApi(domain, port, api, query, pointInTime string) error
}

func (qresp *queryResponse) connectToPrometheusApi(domain, port, api, query, pointInTime string) error {

	values := url.Values{}
	values.Set("query", query)
	if pointInTime != "" {
		values.Set("time", pointInTime)
	}

	uri := domain + ":" + port + api
	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	url := uri + values.Encode()

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("http request to the api (%s) failed with '%s'\n", uri, err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unable to get a reponse from the api. \n http response was '%s' (%d) and not OK (200). Body:\n%s\n", resp.Status, resp.StatusCode, string(body))
	}

	return qresp.decodeQueryResponse(body)

}

func whatIsThisThing(something interface{}) (float64, string, error) {
	switch i := something.(type) {
	case float64:
		return float64(i), "", nil
	case string:
		return float64(0), string(i), nil
	default:
		return float64(0), "", errors.New("unknown value is of incompatible type")
	}
}

func (qresp *queryResponse) decodeQueryResponse(jresponse []byte) error {

	err := json.Unmarshal(jresponse, &qresp)
	if err != nil {
		return fmt.Errorf("http response can't be Unmarshal: %v", err)
	}

	return nil
}

func (pt *prometheusTest) CreateResources(namespace string) {
	qresp := &queryResponse{}
	err := qresp.connectToPrometheusApi(pt.domain, pt.port, pt.api, pt.metricQuery, "")
	So(err, ShouldBeNil)

	pt.beforeBackup = *qresp
	point := pointInTime{}
	if len(qresp.Data.Result) > 0 && len(qresp.Data.Result[0].Value) > 0 {
		values := qresp.Data.Result[0].Value
		for _, something := range values {

			f, s, err := whatIsThisThing(something)
			So(err, ShouldBeNil)

			if f != float64(0) {
				point.pointInTime(f)
				pt.pointInTime = point
			}

			if s != "" {
				pt.expectedResult = s
			}
		}

	}

}

func (pt *prometheusTest) TestResources(namespace string) {
	qresp := &queryResponse{}
	err := qresp.connectToPrometheusApi(pt.domain, pt.port, pt.api, pt.metricQuery, pt.pointInTime.formmattedValue)
	So(err, ShouldBeNil)

	if len(qresp.Data.Result) > 0 && len(qresp.Data.Result[0].Value) > 0 {
		values := qresp.Data.Result[0].Value
		for _, something := range values {

			_, s, err := whatIsThisThing(something)
			So(err, ShouldBeNil)
			if s != "" {
				pt.finalResult = s
			}

		}

	}

	So(strings.TrimSpace(pt.finalResult), ShouldEqual, strings.TrimSpace(pt.expectedResult))
}
