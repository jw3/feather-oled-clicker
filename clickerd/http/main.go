package main

import (
	. "../common"
	"fmt"
	"github.com/go-yaml/yaml"
	ppc "github.com/jw3/ppc/cli"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/xujiajun/gorouter"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-lib/metrics"
	"strings"

	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
)


func main() {
	clickerConf, e := parseClickerConf()
	if e != nil {
		log.Fatalf("failed to read clicker configuration: %v", e)
	}

	cfg := ppc.NewConfiguration()

	traceCfg := jaegercfg.Configuration{
		ServiceName: "clicker-http",
		Sampler:     &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter:    &jaegercfg.ReporterConfig{
			LogSpans: true,
		},
	}

	jLogger := jaegerlog.StdLogger
	jMetricsFactory := metrics.NullFactory
	tracer, closer, err := traceCfg.NewTracer(
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
	)
	if err != nil {
		panic(err)
	}

	opentracing.SetGlobalTracer(tracer)
	defer closer.Close()

	items := make(chan Item)
	mux := gorouter.New()
	http.HandleFunc("/health", handleHealth)
	mux.POST("/click/:model", func(w http.ResponseWriter, r *http.Request) {
		sid := gorouter.GetParam(r, "model")
		id, e := strconv.Atoi(sid)
		if e != nil {
			log.Printf("Invalid numeric item id: %v", sid)
		}

		iid := id - 1
		if !(iid < len(clickerConf.Items)) {
			log.Printf("selected item id is out of bounds: %v", iid)
		}

		item := clickerConf.Items[iid]
		items <- item
	})


	go func() {
		for {
			item := <-items
			log.Printf("selected model: %v", item.Title)
			call(&item, cfg)
		}
	}()

	http.Handle("/", mux)
	log.Fatal(http.ListenAndServe(":9000", nil))
}


func handleHealth(writer http.ResponseWriter, _ *http.Request) {
	writer.WriteHeader(http.StatusOK)
}


func parseClickerConf() (Cfg, error) {
	cfg := Cfg{}
	json, e := ioutil.ReadFile("/usr/local/etc/clickerd.conf")
	if e == nil {
		e = yaml.UnmarshalStrict(json, &cfg)
	}
	return cfg, e
}


func call(item *Item, cfg *ppc.Config, tracer opentracing.Tracer) error {
	clientSpan := tracer.StartSpan("client")
	defer clientSpan.Finish()

	for _, m := range item.Modules {
		movementCommand := "move" // todo;; externalize
		endpoint := fmt.Sprintf("http://%s/devices/%s/%s", cfg.ApiUri, m.Id, movementCommand)

		v := url.Values{}
		v.Set("args", m.Model)
		req, _ := http.NewRequest("POST", endpoint, strings.NewReader(v.Encode()))

		ext.SpanKindRPCClient.Set(clientSpan)
		ext.HTTPUrl.Set(clientSpan, endpoint)
		ext.HTTPMethod.Set(clientSpan, "POST")
		tracer.Inject(clientSpan.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))

		if _, e := http.DefaultClient.Do(req); e != nil {
			log.Printf("move failed for %v", m.Id)
			return e
		}
	}
	return nil
}
