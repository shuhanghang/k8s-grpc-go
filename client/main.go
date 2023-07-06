package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/shuhanghang/k8s-grpc-go/pb"
	"github.com/shuhanghang/k8s-grpc-go/utils"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
)

const (
	exampleScheme = "example"
)

var (
	k8sEndpoint      utils.EndPoint
	resConn          resolver.ClientConn
	nameSpace        = flag.String("ns", "default", "NameSpace")
	endPointSelector = flag.String("ends", "name=server-svc", "EndpointSelector")
)

func init() {
	flag.Parse()
	k8sEndpoint = utils.EndPoint{NameSpace: *nameSpace, EndPointLabelSelector: *endPointSelector}
	k8sEndpoint.Init()
	resolver.Register(&exampleResolverBuilder{})

}

func makeRPCs(cc *grpc.ClientConn) {
	c := pb.NewExampleServiceClient(cc)
	for i := 1; i < 100; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
		defer cancel()
		r, err := c.Service(ctx, &pb.ExampleRequest{Req: "grpc"})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		log.Printf("%v", r.Result)
		time.Sleep(2 * time.Second)
	}
}

type exampleResolverBuilder struct{}

func (*exampleResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := &exampleResolver{
		// target: target,
		cc: cc,
		// addrsStore: map[string][]string{
		// 	exampleServiceName: {backendAddr, backendAddr2},
	}
	resConn = r.cc
	r.start()
	return r, nil
}
func (*exampleResolverBuilder) Scheme() string { return exampleScheme }

// exampleResolver is a
// Resolver(https://godoc.org/google.golang.org/grpc/resolver#Resolver).
type exampleResolver struct {
	// target     resolver.Target
	cc resolver.ClientConn
	// addrsStore map[string][]string
}

func (r *exampleResolver) start() {
	// addrStrs := r.addrsStore[r.target.Endpoint()]
	// addrs := make([]resolver.Address, len(addrStrs))
	// for i, s := range addrStrs {
	// 	addrs[i] = resolver.Address{Addr: s}
	// }
	addrs := k8sEndpoint.Get()

	r.cc.UpdateState(resolver.State{Addresses: addrs})
}
func (*exampleResolver) ResolveNow(o resolver.ResolveNowOptions) {}
func (*exampleResolver) Close()                                  {}

func main() {
	exampleConn, err := grpc.Dial(
		fmt.Sprintf("%s:///", exampleScheme),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer exampleConn.Close()

	var wg sync.WaitGroup
	go k8sEndpoint.Watch(resConn)
	makeRPCs(exampleConn)
	wg.Add(1)
	wg.Wait()
}
