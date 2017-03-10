package grpccli

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/golang/protobuf/proto"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func RunE(
	addr, _input *string,
	method,
	inT string,
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

			in := reflect.New(proto.MessageType(inT).Elem()).Interface()
			if len(*_input) > 0 {
				if err := json.Unmarshal([]byte(*_input), in); err != nil {
					return err
				}
			}

			result := method.Call([]reflect.Value{
				reflect.ValueOf(context.Background()),
				reflect.ValueOf(in),
			})
			if len(result) != 2 {
				panic("service methods should always return 2 values")
			}
			if !result[1].IsNil() {
				return result[1].Interface().(error)
			}
			out := result[0].Interface()
			data, err := json.MarshalIndent(out, "", "    ")
			if err != nil {
				return err
			}
			fmt.Println(out)
			fmt.Println(string(data))
		}

		return nil
	}
}

func dial(addr string) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	return grpc.Dial(addr, opts...)
}
