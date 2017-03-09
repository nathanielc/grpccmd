package grpccli

import (
	"fmt"
	"reflect"

	"github.com/golang/protobuf/proto"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func RunE(
	addr *string,
	method,
	inT, outT string,
	newClient func(*grpc.ClientConn) interface{},
) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		conn, err := dial(*addr)
		if err != nil {
			return err
		}
		defer conn.Close()
		c := newClient(conn)
		cv := reflect.ValueOf(c)
		method := cv.MethodByName(method)
		if method.IsValid() {

			in := reflect.New(proto.MessageType(inT).Elem())

			result := method.Call([]reflect.Value{
				reflect.ValueOf(context.Background()),
				in,
			})
			if len(result) != 2 {
				panic("service methods should always return 2 values")
			}
			err := result[1].Interface().(error)
			if err != nil {
				return err
			}
			out := result[0].Interface()
			fmt.Println(out)
		}

		return nil
	}
}

func dial(addr string) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	return grpc.Dial(addr, opts...)
}
