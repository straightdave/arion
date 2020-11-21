// Postgal
// auto-generated source code: main.go
// by arion (https://github.com/straightdave/arion)
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	protoempty "github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	gozip "github.com/straightdave/gozip/lib"
	"github.com/straightdave/lesphina/v2"
	trunks "github.com/straightdave/trunks-lib"
)

var (
	_les             *lesphina.Lesphina
	_renderData      *RenderData
	_hostList        []string
	_regexMethodPath = regexp.MustCompile(`\.[Invoke|NewStream].+?"(.+?)"`)
	_regexDurFormat  = regexp.MustCompile(`\d+[sm]`)
	_regexDataMark   = regexp.MustCompile(`(?im)<<arion:(\w*?):(.*?)>>`)

	// flags
	_ver            = flag.Bool("v", false, "print version info")
	_debug          = flag.Bool("debug", false, "print some debug info (for dev purpose)")
	_serve          = flag.Bool("serve", false, "browser mode")
	_arionHost      = flag.String("at", ":9999", "address to host Postgal in browser mode")
	_info           = flag.Bool("i", false, "print basic service info")
	_type           = flag.String("t", "", "data type name")
	_hosts          = flag.String("h", ":8087", "hosts of target service (commas to seperate multiple hosts)")
	_endpoint       = flag.String("e", "", "endpoint name (<svc_name>#<end_name> or just <end_name>) to execute")
	_json           = flag.Bool("json", false, "print response in JSON format")
	_isMassive      = flag.Bool("x", false, "stress test mode")
	_data           = flag.String("d", "", "request data string")
	_meta           = flag.String("meta", "", "gRPC metadata (format: 'key1=value1 key2=value2')")
	_maxRecv        = flag.Int("maxrecv", 0, "set max size of received message")
	_maxSend        = flag.Int("maxsend", 0, "set max size of sending message")
	_connPerHost    = flag.Uint64("C", 1, "[Unary stress test] number of concurrent connections per each host")
	_streamMode     = flag.String("m", "unary", "unary | client | server | bidirect")
	_streamingTimes = flag.Uint("n", 10, "how many times to send streaming msg to server in one connection")
	_numStream      = flag.Uint("N", 1, "how many concurrent streaming connections when doing stress test mode (-x)")
	_dataFile       = flag.String("df", "", "request data file. Data in the file will be sent line by line")
	_binFile        = flag.String("B", "", "protobuf binary data file")
	_loop           = flag.Bool("loop", false, "repeatly sending all requests in data file (-df)")
	_duration       = flag.Duration("duration", 10*time.Second, "test duration in stress test mode: 10s, 20m")
	_rate           = flag.Uint64("rate", 1, "expected QPS (query per second) in stress test mode")
	_worker         = flag.Uint64("worker", 10, "workers (max volume of goroutine pool)")
	_dumpToFile     = flag.String("dumpto", "", "dump massive call responses to file")
	_genBin         = flag.String("G", "", "the bin file name to generate")
)

var _version string

func init() {
	flag.Parse()
	_les = lesphina.Restore(_lesDump)
	_hostList = parseHostList(*_hosts)
	genRenderData()
}

func main() {
	if len(os.Args) == 1 {
		flag.PrintDefaults()
		return
	}

	if *_ver {
		var svcNames []string
		for _, svc := range _renderData.Services {
			svcNames = append(svcNames, svc.Name)
		}
		fmt.Printf("Service:\t%s\n", strings.Join(svcNames, ", "))
		fmt.Printf("Generated:\t%s\n", _version)
		return
	}

	if *_serve {
		realHtml := genHtml(gozip.DecompressString(html2))
		serveStaticContent("/", "text/html", realHtml)
		serveStaticContent("/m.css", "text/css", gozip.DecompressString(css))
		serveStaticContent("/m.js", "text/javascript", gozip.DecompressString(js))
		http.HandleFunc("/type", handleGetType)
		http.HandleFunc("/meta", handleGetLesphinaMeta)
		http.HandleFunc("/call", handleCallEndpoint)
		srv := &http.Server{Addr: *_arionHost}
		log.Fatal(srv.ListenAndServe())
	}

	if *_info {
		if *_endpoint != "" {
			genAndPrintEndpointInfo(*_endpoint)
			return
		}
		if *_type != "" {
			genAndPrintTypeInfo(*_type)
			return
		}
		genAndPrintOverallInfo()
		return
	}

	if *_endpoint != "" {
		req := &CallEndpointRequest{
			Hostlist:       _hostList,
			Endpoint:       *_endpoint,
			IsMassive:      *_isMassive,
			DataInJSON:     *_data,
			Metadata:       parseMetadata(*_meta),
			StreamMode:     guessStreamMode(*_streamMode),
			StreamingTimes: *_streamingTimes,
			NumStream:      *_numStream,
			RespInJSON:     *_json,
			DataFile:       *_dataFile,
			IsLoop:         *_loop,
			DumpFile:       *_dumpToFile,
			CallDuration:   *_duration,
			Rate:           *_rate,
			Worker:         *_worker,
			MaxRecv:        *_maxRecv,
			MaxSend:        *_maxSend,
			ConnPerHost:    *_connPerHost,
		}

		guessCallReqFields(req)

		if len(*_binFile) != 0 {
			d, err := ioutil.ReadFile(*_binFile)
			if err != nil {
				fmt.Printf("Failed to read bin file: %v\n", err)
				return
			}
			req.BinData = d
		}

		// gen bin data file
		if len(*_genBin) != 0 {
			err := req.serialize(*_genBin)
			if err != nil {
				fmt.Printf("Failed to generate bin file: %v\n", err)
			}
			return
		}

		if req.IsMassive {
			req.Massive()
			return
		}

		res, err := req.Call()
		if err != nil {
			fmt.Println("ERROR", err)
			return
		}
		fmt.Println(res)
	}
}

func guessCallReqFields(req *CallEndpointRequest) {
	svcName, endName := parseEndpointName(req.Endpoint)
	method, err := getMethodPath(svcName, endName)
	if err != nil {
		return
	}

	if svcName == "" && len(_renderData.Services) > 1 {
		fmt.Println("mbiguous endpoint name; more than one services; please use <svc-name>#<end-name>")
		return
	}

	if svcName == "" && len(_renderData.Services) == 1 {
		svcName = _renderData.Services[0].Name
	}

	req.Method = method
	req.SvcName = svcName
	req.EndName = endName

	// guess req/resp type names
	reqType := req.EndName + "Request"
	respType := req.EndName + "Response"

outter:
	for _, svc := range _renderData.Services {
		if svc.Name == req.SvcName {
			for _, e := range svc.Endpoints {
				if e.Name == req.EndName {
					reqType = e.ReqType
					respType = e.RespType
					break outter
				}
			}
		}
	}
	req.RequestTypeName = reqType
	req.ResponseTypeName = respType
}

func guessStreamMode(input string) Mode {
	switch strings.ToLower(input) {
	case "server":
		return ServerSide
	case "client":
		return ClientSide
	case "bidirect":
		return Bidirectional
	default:
		return Unary
	}
}

func parseHostList(origin string) []string {
	origin = strings.Trim(origin, " ,")
	tmpList := strings.Split(origin, ",")
	if len(tmpList) == 0 {
		return []string{":8087"}
	}

	var result []string
	for _, h := range tmpList {
		h = strings.Trim(h, " ")
		indexOfColon := strings.Index(h, ":")
		if indexOfColon >= 0 {
			result = append(result, h)
		} else {
			// no colon, no port
			result = append(result, h+":8087")
		}
	}
	return result
}

func genAndPrintEndpointInfo(raw string) {
	svcName, endName := parseEndpointName(raw)
	for _, svc := range _renderData.Services {
		if svcName != "" && svc.Name != svcName {
			continue
		}

		for _, end := range svc.Endpoints {
			if end.Name == endName {
				fmt.Printf("%s#%s\n", svc.Name, end.Name)
				fmt.Println("- Request entity:", end.ReqType)
				for _, req := range end.ReqElements {
					fmt.Printf("--- %s %s (JSON field name: %s)\n", req.Name, req.Type, req.JSONFieldName)
				}
				fmt.Println("- Response entity:", end.RespType)
				for _, res := range end.RespElements {
					fmt.Printf("--- %s %s (JSON field name: %s)\n", res.Name, res.Type, res.JSONFieldName)
				}
				break
			}
		}
	}
}

func genAndPrintTypeInfo(typeName string) {
	for _, t := range _renderData.Reference {
		if t.Name == typeName {
			for _, f := range t.Fields {
				fmt.Printf("- %s %s (JSON field name: %s)\n", f.Name, f.Type, f.JSONFieldName)
			}
			break
		}
	}
}

func genAndPrintOverallInfo() {
	for _, svc := range _renderData.Services {
		fmt.Println(svc.Name)
		for _, e := range svc.Endpoints {
			fmt.Printf("> %s\n", e.Name)
		}
	}
}

func getFullHost(raw string) string {
	if strings.HasPrefix(raw, ":") {
		return "0.0.0.0" + raw
	}
	return raw
}

func getPort(rawAddress string) string {
	splits := strings.Split(rawAddress, ":")
	if len(splits) >= 2 {
		last := splits[len(splits)-1]
		last = strings.TrimSpace(last)
		if last != "" {
			return last
		}
	}
	return "9999"
}

type RenderData struct {
	ArionVersion string
	ArionPort    string
	TargetHost   string
	Services     []ServiceData
	Reference    []Struct
}

type Struct struct {
	Name   string
	Fields []ElementData
}

type ServiceData struct {
	Name      string
	Endpoints []EndpointData
}

type EndpointData struct {
	Name         string
	ReqType      string
	ReqElements  []ElementData
	RespType     string
	RespElements []ElementData
}

type ElementData struct {
	Name          string
	Type          string
	JSONFieldName string
}

func genRenderData() {
	_renderData = &RenderData{
		ArionVersion: _version,
		TargetHost:   getFullHost(_hostList[0]),
		ArionPort:    getPort(*_arionHost),
	}

	rawInterfaces := _les.Query().ByKind(lesphina.KindInterface).ByName("~Client").All()
	for _, intf := range rawInterfaces {
		svc, ok := intf.(*lesphina.Interface)
		if !ok {
			fmt.Printf("WARN: the obj [%+v] found is not interface, weird\n", intf)
			continue
		}

		svcData := ServiceData{
			Name: strings.TrimSuffix(svc.GetName(), "Client"),
		}

		for _, m := range svc.Methods {
			endData := EndpointData{
				Name: m.GetName(),
			}

			for _, p := range m.InParams() {
				lowerType := strings.ToLower(p.BaseType)
				if strings.Contains(lowerType, "context") || strings.Contains(lowerType, "option") {
					continue
				}

				// Unary
				if p.IsPointer && p.Name == "in" {
					endData.ReqType = p.BaseType
					reqstru := _les.Query().ByKind(lesphina.KindStruct).ByName(p.BaseType).First()
					if st, ok := reqstru.(*lesphina.Struct); ok {
						for _, f := range st.Fields {
							endData.ReqElements = append(endData.ReqElements, ElementData{
								Name:          f.Name,
								Type:          f.BaseType,
								JSONFieldName: f.JSONFieldName(),
							})
						}
						break
					}
					// if not found in lesphina, this req type would not count any sub elements
				}
				// Will check response parameters for streaming req/resp parameter types
			}

			for _, p := range m.OutParams() {
				// Unary
				if p.IsPointer {
					endData.RespType = p.BaseType
					reqstru := _les.Query().ByKind(lesphina.KindStruct).ByName(p.BaseType).First()
					if st, ok := reqstru.(*lesphina.Struct); ok {
						for _, f := range st.Fields {
							endData.RespElements = append(endData.RespElements, ElementData{
								Name:          f.Name,
								Type:          f.BaseType,
								JSONFieldName: f.JSONFieldName(),
							})
						}
						break
					}
				}

				// streaming (client-side or bi-dir) naming in the pattern of 'xxxx_xxxxClient'
				if strings.HasSuffix(p.BaseType, "Client") {
					o := _les.Query().ByKind(lesphina.KindInterface).ByName(p.BaseType).First()
					if o == nil {
						fmt.Printf("ERR: Cannot find the interface:%s\n", p.BaseType)
						return
					}

					if oi, ok := o.(*lesphina.Interface); ok {
						for _, m := range oi.Methods {

							// NOTE: relies heavily on grpc protoc naming
							// to collect request entity
							if m.Name == "Send" && len(m.In) > 0 {
								endData.ReqType = m.In[0].BaseType

								reqT := _les.Query().ByKind(lesphina.KindStruct).ByName(endData.ReqType).First()
								if t, ok := reqT.(*lesphina.Struct); ok {
									for _, f := range t.Fields {
										endData.ReqElements = append(endData.ReqElements, ElementData{
											Name:          f.Name,
											Type:          f.BaseType,
											JSONFieldName: f.JSONFieldName(),
										})
									}
								}
							}

							// to collect response entity
							if strings.Contains(strings.ToLower(m.Name), "recv") && len(m.Out) > 0 {
								endData.RespType = m.Out[0].BaseType

								respT := _les.Query().ByKind(lesphina.KindStruct).ByName(endData.RespType).First()
								if t, ok := respT.(*lesphina.Struct); ok {
									for _, f := range t.Fields {
										endData.RespElements = append(endData.RespElements, ElementData{
											Name:          f.Name,
											Type:          f.BaseType,
											JSONFieldName: f.JSONFieldName(),
										})
									}
								}
							}
						}
					}
				}
			}

			svcData.Endpoints = append(svcData.Endpoints, endData)
		}
		_renderData.Services = append(_renderData.Services, svcData)
	}

	rawAllStructs := _les.Query().ByKind(lesphina.KindStruct).All()
	for _, stru := range rawAllStructs {
		s, ok := stru.(*lesphina.Struct)
		if !ok {
			fmt.Println("the obj found is not Struct, weird")
			continue
		}

		// hide non-exporting structures
		if strings.Title(s.Name) != s.Name {
			continue
		}

		structData := Struct{Name: s.Name}
		for _, f := range s.Fields {
			structData.Fields = append(structData.Fields, ElementData{
				Name:          f.Name,
				Type:          f.RawType,
				JSONFieldName: f.JSONFieldName(),
			})
		}
		_renderData.Reference = append(_renderData.Reference, structData)
	}
}

func genHtml(raw string) string {
	t, err := template.New("html2").Parse(raw)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, _renderData); err != nil {
		panic(err)
	}
	return buf.String()
}

func serveStaticContent(path, mime, content string) {
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		defer recoverHandling(w)
		w.Header().Add("Content-Type", mime)
		io.WriteString(w, content)
	})
}

func handleGetType(w http.ResponseWriter, r *http.Request) {
	defer recoverHandling(w)

	typeName := strings.TrimSpace(r.URL.Query().Get("n"))
	if typeName == "" {
		http.Error(w, "lacking of query keyword (n)", http.StatusBadRequest)
		return
	}

	t := _les.Query().ByName(typeName).First()
	switch t.GetKind() {
	case lesphina.KindFunction:
		tt := t.(*lesphina.Function)
		io.WriteString(w, tt.JSON())
	case lesphina.KindStruct:
		tt := t.(*lesphina.Struct)
		io.WriteString(w, tt.JSON())
	case lesphina.KindElement:
		tt := t.(*lesphina.Element)
		io.WriteString(w, tt.JSON())
	default:
		io.WriteString(w, `{"err":"unknown or unsupport type"}`)
	}
}

func handleGetLesphinaMeta(w http.ResponseWriter, r *http.Request) {
	defer recoverHandling(w)
	io.WriteString(w, _les.Meta.JSON())
}

func handleCallEndpoint(w http.ResponseWriter, r *http.Request) {
	defer recoverHandling(w)

	if r.Method != "POST" {
		http.Error(w, "only support POST", http.StatusBadRequest)
		return
	}

	v, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	endpointName := v.Get("e")
	if endpointName == "" {
		http.Error(w, "lacking of query keyword (e)", http.StatusBadRequest)
		return
	}

	hostAddr := v.Get("h")
	if hostAddr == "" {
		http.Error(w, "lacking of query keyword (h)", http.StatusBadRequest)
		return
	}

	var respInJSON bool
	respFormat := v.Get("format")
	if strings.ToLower(respFormat) == "json" {
		respInJSON = true
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	data := string(body)

	// parse grpc metadata in http request header
	meta := make(metadata.MD)
	for k, v := range r.Header {
		if strings.HasPrefix(k, "Gmeta-") { // first letter auto converted to capitalized
			realKey := strings.TrimPrefix(k, "Gmeta-")
			meta.Set(realKey, v...) // v is []string
		}
	}

	call := &CallEndpointRequest{
		Hostlist:   []string{hostAddr},
		Endpoint:   endpointName,
		DataInJSON: data,
		Metadata:   meta,
		RespInJSON: respInJSON,
	}
	log.Printf("call endpoint: %s \n\thosts: %v \n\tdata: %s \n\tmeta: %#v",
		call.Endpoint, call.Hostlist, call.DataInJSON, call.Metadata)

	guessCallReqFields(call)

	res, err := call.Call()
	if err != nil {
		log.Println("Err:", err.Error())
		io.WriteString(w, err.Error())
		return
	}
	io.WriteString(w, res)
}

func parseEndpointName(rawName string) (string, string) {
	var svcName, endName string
	i := strings.LastIndex(rawName, "#")
	if i > 0 {
		svcName = rawName[:i]
		endName = rawName[i+1:]
	} else {
		endName = rawName
	}
	return svcName, endName
}

type Mode int

const (
	Unary Mode = iota
	ClientSide
	ServerSide
	Bidirectional
)

type CallEndpointRequest struct {
	Hostlist         []string
	Endpoint         string
	Method           string
	SvcName          string
	EndName          string
	RequestTypeName  string
	ResponseTypeName string
	IsMassive        bool
	DataInJSON       string
	BinData          []byte
	DataFile         string
	Metadata         metadata.MD
	DialTimeout      time.Duration
	StreamMode       Mode
	StreamingTimes   uint
	NumStream        uint
	RespInJSON       bool
	CallDuration     time.Duration
	Rate             uint64
	Worker           uint64
	IsLoop           bool
	DumpFile         string
	MaxRecv          int
	MaxSend          int
	ConnPerHost      uint64
}

// Call ...
func (req *CallEndpointRequest) Call() (response string, err error) {
	max := func(du1, du2 time.Duration) time.Duration {
		if du1 > du2 {
			return du1
		}
		return du2
	}

	var cos []grpc.CallOption
	if req.MaxRecv > 0 {
		cos = append(cos, grpc.MaxCallRecvMsgSize(req.MaxRecv))
	}
	if req.MaxSend > 0 {
		cos = append(cos, grpc.MaxCallSendMsgSize(req.MaxSend))
	}

	var conn *grpc.ClientConn
	if len(cos) > 0 {
		conn, err = grpc.Dial(
			req.Hostlist[0],
			grpc.WithDefaultCallOptions(cos...),
			grpc.WithInsecure(),
			grpc.WithBlock(),
			grpc.WithTimeout(max(req.DialTimeout, 5*time.Second)))
	} else {
		conn, err = grpc.Dial(
			req.Hostlist[0],
			grpc.WithInsecure(),
			grpc.WithBlock(),
			grpc.WithTimeout(max(req.DialTimeout, 5*time.Second)))
	}
	if err != nil {
		return
	}
	defer conn.Close()

	ctx := metadata.NewOutgoingContext(context.Background(), req.Metadata)

	var replyData proto.Message
	switch req.StreamMode {

	case ClientSide:
		replyData, err = req.sendClientStream(ctx, conn)

	case ServerSide:
		err = req.sendServerStream(ctx, conn)
		if err != nil {
			log.Printf("ERR: failed to do server streaming: %v", err)
		}
		return

	case Bidirectional:
		err = req.sendBidirectStream(ctx, conn)
		if err != nil {
			log.Printf("ERR: failed to do server streaming: %v", err)
		}
		return

	default:
		replyData, err = req.unaryCall(ctx, conn)
	}

	// for unary call, will print its response

	if err != nil {
		log.Printf("ERR failed to call: %v", err)
		return
	}

	if respMeta, ok := metadata.FromIncomingContext(ctx); ok {
		fmt.Println("Metadata from the incoming context (response):")
		fmt.Printf("%+v\n", respMeta)
	}

	if req.RespInJSON {
		jsonbytes, err := json.MarshalIndent(replyData, "", "    ")
		if err != nil {
			fmt.Printf("ERR: failed to marshal response: %v\n", err)
			return "", err
		}
		response = string(jsonbytes)
	} else {
		response = replyData.String()
	}
	return
}

// Massive ...
func (req *CallEndpointRequest) Massive() {
	fmt.Printf("Massive Call on %s ...\n", req.Endpoint)

	if req.StreamMode != Unary {
		wg := &sync.WaitGroup{}
		for i := uint(0); i < req.NumStream; i++ {
			wg.Add(1)
			go func(index uint) {
				// log.Printf("start worker: %d", index)
				defer wg.Done()

				defer func() {
					if r := recover(); r != nil {
						log.Printf("[# %d] panic recovered: %v", index, r.(error))
					}
				}()

				conn, err := grpc.Dial(
					req.Hostlist[0], // only use the first host
					grpc.WithInsecure(),
					grpc.WithBlock(),
					grpc.WithTimeout(5*time.Second))

				if err != nil {
					log.Printf("failed to dial to server: %v", err)
					return
				}
				defer conn.Close()

				switch req.StreamMode {
				case ClientSide:
					replyData, err := req.sendClientStream(context.Background(), conn)
					if err != nil {
						log.Printf("[# %d] ERR: %v", index, err)
						return
					}
					fmt.Printf("[# %d] response: %+v\n", index, replyData)
				case ServerSide:
					err := req.sendServerStream(context.Background(), conn)
					if err != nil {
						log.Printf("[# %d] ERR: %v", index, err)
						return
					}
					fmt.Printf("[# %d] server streaming finished\n", index)
				case Bidirectional:
					fmt.Println("Bidirectional streaming stress test is in my plan :)")
					return
				}
			}(i)
		}
		wg.Wait()
		return
	}

	// Unary stress call
	var target *trunks.Gtarget
	if req.DataFile != "" {
		target = prepMultipleReqTarget(req)
	} else if req.DataInJSON != "" {
		target = prepSingleReqTarget(req)
	} else {
		fmt.Println("Please specify either -d or -df")
		return
	}

	fmt.Printf("%d connection(s) per host, %d host(s)\n", req.ConnPerHost, len(req.Hostlist))

	burner, err := trunks.NewBurner(
		req.Hostlist,
		trunks.WithMaxRecvSize(req.MaxRecv),
		trunks.WithMaxSendSize(req.MaxSend),
		trunks.WithNumConnPerHost(req.ConnPerHost),
		trunks.WithLooping(req.IsLoop),
		trunks.WithNumWorker(req.Worker),
		trunks.WithDumpFile(req.DumpFile),
		trunks.WithMetadata(req.Metadata))
	if err != nil {
		panic(err)
	}
	defer burner.Close()

	fmt.Println("Start at", time.Now().Format(time.UnixDate))
	var metrics trunks.Metrics
	startT := time.Now()
	for res := range burner.Burn(target, req.Rate, req.CallDuration) {
		metrics.Add(res)
	}
	dur := time.Since(startT)
	metrics.Close()

	if err := burner.WaitDumpDone(); err != nil {
		fmt.Printf("WARN: failed to dump response to file: %v\n", err)
	}

	fmt.Println("----------------------")
	fmt.Printf("Duration: %v\n", dur.Seconds())
	fmt.Printf("Actual Rate: %f\n", metrics.Rate)
	fmt.Printf("Total Requests: %d\n", metrics.Requests)
	fmt.Printf("Success: %f\n", metrics.Success)
	fmt.Printf("Mean: %s\n", metrics.Latencies.Mean)
	fmt.Printf("50th: %s\n", metrics.Latencies.P50)
	fmt.Printf("95th: %s\n", metrics.Latencies.P95)
	fmt.Printf("99th: %s\n", metrics.Latencies.P99)
	fmt.Printf("Max: %s\n", metrics.Latencies.Max)
}

// massive call
// read request from command line '-d'
// in this case, it supports unique/random value
func prepSingleReqTarget(req *CallEndpointRequest) *trunks.Gtarget {
	reply := getVarByTypeName(req.ResponseTypeName)
	if reply == nil {
		log.Println("WARN response type not found; treating as protobuf.Empty")
		reply = &protoempty.Empty{}
	}

	reqData := getVarByTypeName(req.RequestTypeName)
	if reqData == nil {
		log.Println("WARN request type not found; treating as protobuf.Empty")
		reqData = &protoempty.Empty{}

		return &trunks.Gtarget{
			MethodName: req.Method,
			Requests:   []proto.Message{reqData},
			Response:   reply,
		}
	}

	// see if data has random/unique marks
	// to generate lots of requests based on that
	if strings.Contains(*_data, "<<arion:") {
		// total amount of requests to be generated
		total := *_rate * uint64((*_duration).Seconds())
		if total > 10000000 {
			// Dirty solution:
			// suppose each request is 100 byte large, 10000000 requests are ~1G
			// we consider this as the bar for now
			panic("Amount is too large (> 10000000) to generate unique requests")
		}

		var requests []proto.Message
		genData := genUniqueData(*_data, total)
		for _, d := range genData {
			r := getVarByTypeName(req.RequestTypeName) // must not nil
			if err := jsonpb.UnmarshalString(d, r); err != nil {
				panicf("ERR unmarshalling requests [%s] failed: %v\n", d, err)
			}
			requests = append(requests, r)
		}
		log.Printf("Generated %d requests\n", len(requests))

		return &trunks.Gtarget{
			MethodName: req.Method,
			Requests:   requests,
			Response:   reply,
		}
	}

	if err := jsonpb.UnmarshalString(req.DataInJSON, reqData); err != nil {
		panic(err)
	}

	return &trunks.Gtarget{
		MethodName: req.Method,
		Requests:   []proto.Message{reqData},
		Response:   reply,
	}
}

// massive call
// read multiple requests from a file
func prepMultipleReqTarget(req *CallEndpointRequest) *trunks.Gtarget {
	reply := getVarByTypeName(req.ResponseTypeName)
	if reply == nil {
		log.Println("WARN response type not found; treating as protobuf.Empty")
		reply = &protoempty.Empty{}
	}

	bts, err := ioutil.ReadFile(*_dataFile)
	if err != nil {
		panic(err)
	}

	var reqs []proto.Message
	content := string(bts)
	for _, line := range strings.Split(content, "\n") {
		req := getVarByTypeName(req.RequestTypeName)
		if req == nil {
			continue
		}
		if err = jsonpb.UnmarshalString(line, req); err != nil {
			continue
		}
		reqs = append(reqs, req)
	}
	fmt.Printf("with %d requests\n", len(reqs))

	return &trunks.Gtarget{
		MethodName: req.Method,
		Requests:   reqs,
		Response:   reply,
	}
}

func (req *CallEndpointRequest) serialize(binFile string) error {
	if len(req.DataInJSON) < 3 { // magic number
		return fmt.Errorf("No data in JSON")
	}
	reqData, _ := getReqReplyData(req)
	if reqData != nil {
		raw, err := proto.Marshal(reqData)
		if err != nil {
			return err
		}
		fmt.Println(raw)
		return ioutil.WriteFile(binFile, raw, 0644)
	}
	return nil
}

// get accurate method path by parsing pb source
// with svc_name and end_name
func getMethodPath(svcName, endName string) (res string, err error) {
	qCandidates := _les.Query().ByKind(lesphina.KindFunction).ByName(endName).All()
	if len(qCandidates) < 1 {
		return "", fmt.Errorf("No function found by name: %s", endName)
	}

	// fmt.Println("DEBUG: method found:", len(qCandidates))

	var target *lesphina.Function
	for _, c := range qCandidates {
		if f, ok := c.(*lesphina.Function); ok {
			if len(f.Recv) < 1 {
				continue
			}

			// recv type should end with string 'Client'
			if !strings.HasSuffix(f.Recv[0].BaseType, "Client") {
				continue
			}

			if svcName != "" {
				// if svcName is given, return the first candidate under this svc
				if strings.Contains(strings.ToLower(f.Recv[0].BaseType), strings.ToLower(svcName)) {
					target = f
					break
				}
			} else {
				// if svcName is not given, return the first candidate
				target = f
				break
			}
		}
	}

	if target == nil {
		return "", fmt.Errorf("Endpoint Not Found (svcName: %s, endName: %s)", svcName, endName)
	}

	rawBody := target.RawBody
	if rawBody == "" {
		return "", fmt.Errorf("Blank endpoint")
	}

	m := _regexMethodPath.FindAllStringSubmatch(rawBody, 1)
	if len(m) > 0 && len(m[0]) > 1 {
		// fmt.Println("DEBUG: grpc method:", m[0][1])
		return m[0][1], nil
	}

	return "", fmt.Errorf("No match in rawBody")
}

func recoverHandling(w http.ResponseWriter) {
	if r := recover(); r != nil {
		io.WriteString(w, r.(error).Error())
	}
}

func panicf(f string, v ...interface{}) {
	panic(fmt.Sprintf(f, v...))
}

// --- unique value stuff
type replacetask struct {
	mark    string
	mode    string
	typ     string
	strsize int
}

// how to use index to create unique string or int value
// TODO: improve this
func (t *replacetask) do(rawStr *string, index uint64) {
	realvalue := ""
	switch t.typ {
	case "string":
		realvalue = fmt.Sprintf("%d", index)
	case "int":
		realvalue = fmt.Sprintf("%d", index)
	}
	*rawStr = strings.Replace(*rawStr, t.mark, realvalue, 1)
}

// gen random value (now it is only supporting string type)
// and only support the first and only one mark
// { ... <<arion:unique:string:10>> ... } => { ... "0000000123" ...}
func genUniqueData(rawValue string, amount uint64) []string {
	matches := _regexDataMark.FindAllStringSubmatch(rawValue, -1)

	var tasks []*replacetask
	var err error
	for _, subm := range matches {
		task := &replacetask{
			mark: subm[0],
			mode: strings.ToLower(subm[1]),
		}
		spls := strings.Split(subm[2], ":")
		switch strings.ToLower(spls[0]) {
		case "string":
			task.typ = "string"
			if len(spls) > 1 {
				task.strsize, err = strconv.Atoi(spls[1])
				if err != nil {
					panic(err)
				}
			}
		case "int":
			task.typ = "int"
			// TODO: size of number type
		default:
			// only supporting int and string types
			panic("non-supported type")
		}
		tasks = append(tasks, task)
	}

	// apply
	var result []string
	for i := uint64(0); i < amount; i++ {
		tempStr := rawValue // copy original data
		for _, t := range tasks {
			t.do(&tempStr, i)
		}
		result = append(result, tempStr)
	}
	return result
}

func parseMetadata(origin string) metadata.MD {
	md := make(metadata.MD)
	if origin != "" {
		for _, s := range strings.Split(origin, " ") {
			s = strings.Trim(s, " ")
			if s == "" {
				continue
			}
			i := strings.Index(s, "=")
			if i >= 0 {
				md.Append(s[:i], s[i+1:])
			}
		}
	}
	return md
}

// UnaryCall ...
func (req *CallEndpointRequest) unaryCall(ctx context.Context, conn *grpc.ClientConn) (proto.Message, error) {
	reqData, replyData := getReqReplyData(req)
	err := conn.Invoke(ctx, req.Method, reqData, replyData)
	if err != nil {
		log.Printf("ERR: failed to invoke unary request: %v", err)
		return nil, err
	}
	return replyData, nil
}

func (req *CallEndpointRequest) sendClientStream(ctx context.Context, conn *grpc.ClientConn) (proto.Message, error) {
	desc := &grpc.StreamDesc{
		ClientStreams: true,
	}

	stream, err := conn.NewStream(ctx, desc, req.Method)
	if err != nil {
		log.Printf("ERR: cannot create Stream: %v", err)
		return nil, err
	}

	reqData, replyData := getReqReplyData(req)

	if req.IsMassive {
		timeoutCh := time.After(req.CallDuration)
	outter:
		for {
			if err := stream.SendMsg(reqData); err != nil {
				log.Printf("ERR: failed to send streaming msg: %v", err)
				return nil, err
			}
			select {
			case <-timeoutCh:
				break outter
			default:
			}
		}
	} else {
		for i := uint(0); i < req.StreamingTimes; i++ {
			if err := stream.SendMsg(reqData); err != nil {
				log.Printf("ERR: failed to send streaming msg: %v", err)
				return nil, err
			}
		}
	}

	if err := stream.CloseSend(); err != nil {
		log.Printf("ERR: failed to close sending streaming msg: %v", err)
		return nil, err
	}

	if err := stream.RecvMsg(replyData); err != nil {
		log.Printf("ERR: failed to receive streaming response: %v", err)
		return nil, err
	}
	return replyData, nil
}

func (req *CallEndpointRequest) sendServerStream(ctx context.Context, conn *grpc.ClientConn) error {
	desc := &grpc.StreamDesc{
		ServerStreams: true,
	}

	stream, err := conn.NewStream(ctx, desc, req.Method)
	if err != nil {
		log.Printf("ERR: cannot create Stream: %v", err)
		return err
	}

	reqData, _ := getReqReplyData(req)

	if !req.IsMassive {
		return sendAndDrainStream(stream, reqData, req.ResponseTypeName)
	}

	// if is massive and before test duration run out
	// we should re-send the request again to get responses
	timeoutCh := time.After(req.CallDuration)
outter:
	for {
		if err := sendAndDrainStream(stream, reqData, req.ResponseTypeName); err != nil {
			return err
		}

		select {
		case <-timeoutCh:
			break outter
		default:
			// re-new the stream
			stream, err = conn.NewStream(ctx, desc, req.Method)
			if err != nil {
				log.Printf("ERR: cannot create Stream: %v", err)
				return err
			}
		}
	}
	log.Println("server streaming finished")
	return nil
}

func (req *CallEndpointRequest) sendBidirectStream(ctx context.Context, conn *grpc.ClientConn) error {
	if req.IsMassive {
		return fmt.Errorf("stress test for bidirectional streaming is not supported")
	}

	desc := &grpc.StreamDesc{
		ClientStreams: true,
		ServerStreams: true,
	}

	stream, err := conn.NewStream(ctx, desc, req.Method)
	if err != nil {
		log.Printf("ERR: cannot create Stream: %v", err)
		return err
	}

	// reader
	waitc := make(chan struct{})
	go func() {
		defer func() {
			waitc <- struct{}{}
			close(waitc)
		}()

		i := 0
		for {
			m := getVarByTypeName(req.ResponseTypeName)
			err := stream.RecvMsg(m)
			if err == io.EOF {
				fmt.Println("read done.")
				return
			}
			if err != nil {
				log.Printf("Failed to receive msg: %v", err)
				return
			}
			fmt.Println(i, m.String())
			i++
		}
	}()

	// writer
	reqData, _ := getReqReplyData(req)

	for i := uint(0); i < req.StreamingTimes; i++ {
		if err := stream.SendMsg(reqData); err != nil {
			log.Printf("Failed to send a request: %v", err)
			return err
		}
	}

	if err := stream.CloseSend(); err != nil {
		log.Printf("Failed to close: %v", err)
		return err
	}
	fmt.Println("finished sending")

	<-waitc
	return nil
}

// used in server side streaming
func sendAndDrainStream(stream grpc.ClientStream, reqData proto.Message, responseTypeName string) error {
	// send request once
	err := stream.SendMsg(reqData)
	if err != nil {
		log.Printf("ERR: failed to send request: %v", err)
		return err
	}

	// receive streaming responses
	for {
		m := getVarByTypeName(responseTypeName)
		err := stream.RecvMsg(m)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("ERR: failed to receive response msg: %v", err)
			return err
		}
		fmt.Println(m.String())
	}
	return nil
}

func getReqReplyData(req *CallEndpointRequest) (reqData, replyData proto.Message) {
	reqData = getVarByTypeName(req.RequestTypeName)
	if reqData == nil {
		// Need to improve:
		// Currently treating not-found structure as empty
		// and ignoring given data
		// In the future we may improve Lesphina
		log.Printf("WARN request type [%v] not found; use protobuf.Empty instead", req.RequestTypeName)
		reqData = &protoempty.Empty{}
	} else if len(req.DataInJSON) > 0 {
		err := jsonpb.UnmarshalString(req.DataInJSON, reqData)
		if err != nil {
			log.Printf("ERR unmarshall reqData failed: %v", err)
			return
		}
	} else if len(req.BinData) > 0 {
		err := proto.Unmarshal(req.BinData, reqData)
		if err != nil {
			log.Printf("ERR failed to unmarshall BinData as request: %v", err)
			return
		}
	} else {
		log.Fatalf("No data provided (JSON data or BinData) to send!")
	}

	replyData = getVarByTypeName(req.ResponseTypeName)
	if replyData == nil {
		log.Printf("WARN response type [%v] not found; treating as protobuf.Empty", req.ResponseTypeName)
		replyData = &protoempty.Empty{}
	}
	return
}
